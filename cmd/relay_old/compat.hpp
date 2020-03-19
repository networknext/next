#include <curl/curl.h>
#include <string>
#include <cstdio>
#include <array>
#include <cinttypes>
#include <sodium.h>
#include <sstream>
#include "json.hpp"

/*
  This file contains most of the things needed to make this relay version work with the new backend
*/

namespace compat
{
  const long HttpSuccess = 200;
  const uint32_t InitRequestMagic = 0x9083708f;
  const unsigned char Base64TableEncode[65] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
  const uint8_t RelayTokenSize = 32;

  // clang-format off
const int Base64TableDecode[256] =
{
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0,  0,
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0, 62, 63, 62, 62, 63, 52, 53, 54, 55,
    56, 57, 58, 59, 60, 61,  0,  0,  0,  0,  0,  0,  0,  0,  1,  2,  3,  4,  5,  6,
    7,  8,  9,  10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,  0,
    0,  0,  0,  63,  0, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
    41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
};
  // clang-format on

  int base64_encode_data(const uint8_t* input, size_t input_length, char* output, size_t output_size)
  {
    assert(input);
    assert(output);
    assert(output_size > 0);

    char* pos;
    const uint8_t* end;
    const uint8_t* in;

    size_t output_length = 4 * ((input_length + 2) / 3);  // 3-byte blocks to 4-byte

    if (output_length < input_length) {
      return -1;  // integer overflow
    }

    if (output_length >= output_size) {
      return -1;  // not enough room in output buffer
    }

    end = input + input_length;
    in = input;
    pos = output;
    while (end - in >= 3) {
      *pos++ = Base64TableEncode[in[0] >> 2];
      *pos++ = Base64TableEncode[((in[0] & 0x03) << 4) | (in[1] >> 4)];
      *pos++ = Base64TableEncode[((in[1] & 0x0f) << 2) | (in[2] >> 6)];
      *pos++ = Base64TableEncode[in[2] & 0x3f];
      in += 3;
    }

    if (end - in) {
      *pos++ = Base64TableEncode[in[0] >> 2];
      if (end - in == 1) {
        *pos++ = Base64TableEncode[(in[0] & 0x03) << 4];
        *pos++ = '=';
      } else {
        *pos++ = Base64TableEncode[((in[0] & 0x03) << 4) | (in[1] >> 4)];
        *pos++ = Base64TableEncode[(in[1] & 0x0f) << 2];
      }
      *pos++ = '=';
    }

    output[output_length] = '\0';

    return int(output_length);
  }

  class CurlWrapper
  {
   public:
    CurlWrapper();
    ~CurlWrapper();

    static auto curl_write_function(char* ptr, size_t size, size_t nmem, void* std_str) -> size_t;

    CURL* curl;
  };

  inline CurlWrapper::CurlWrapper()
  {
    curl_global_init(0);
    curl = curl_easy_init();
  }

  inline CurlWrapper::~CurlWrapper()
  {
    curl_easy_cleanup(curl);
    curl_global_cleanup();
  }

  inline auto CurlWrapper::curl_write_function(char* ptr, size_t size, size_t nmemb, void* std_str) -> size_t
  {
    std::string& str = *reinterpret_cast<std::string*>(std_str);
    str.reserve(size * nmemb);
    str.append(ptr, ptr + size * nmemb);
    return size * nmemb;
  }

  inline auto next_curl_send(const char* url, const char* data, size_t len, std::string& resp) -> bool
  {
    static CurlWrapper wrapper;
    auto& curl = wrapper.curl;

    auto slist = curl_slist_append(nullptr, "Content-Type:application/json");

    // curl config
    {
      curl_easy_setopt(curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS);
      curl_easy_setopt(curl, CURLOPT_HTTPHEADER, slist);
      curl_easy_setopt(curl, CURLOPT_URL, url);
      curl_easy_setopt(curl, CURLOPT_POSTFIELDS, data);
      curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)len);
      curl_easy_setopt(curl, CURLOPT_USERAGENT, "network next relay");
      curl_easy_setopt(curl, CURLOPT_TCP_KEEPALIVE, 1L);
      curl_easy_setopt(curl, CURLOPT_TIMEOUT_MS, long(1000));
      curl_easy_setopt(curl, CURLOPT_NOPROGRESS, 1L);
      curl_easy_setopt(curl, CURLOPT_BUFFERSIZE, 102400L);
      curl_easy_setopt(curl, CURLOPT_WRITEDATA, &resp);
      curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, &CurlWrapper::curl_write_function);
    }

    CURLcode ret = curl_easy_perform(curl);

    if (ret != 0) {
      printf("curl error: %u\n", ret);
      return false;
    }

    long code;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &code);
    if (code < HttpSuccess || code >= 300) {
      printf("curl response not success: %ld\n", code);
      std::exit(1);
    }

    curl_slist_free_all(slist);

    return true;
  }

  // magic
  // nonce
  // address
  // encrypted token
  inline auto next_curl_init(const char* relayBackendAddr,
   const std::string& relayAddress,
   int port,
   uint8_t* routerPublicKey,
   uint8_t* relayPrivateKey,
   std::string& response) -> bool
  {
    // nonce
    uint8_t nonce[crypto_box_NONCEBYTES];
    randombytes_buf(nonce, sizeof(nonce));
    char base64Nonce[sizeof(nonce) * 2] = {};
    base64_encode_data(nonce, sizeof(nonce), base64Nonce, sizeof(base64Nonce));
    std::string base64NonceStr = base64Nonce;

    // encrypted token
    // just used for proving encryption works between the relay & backend, garbage data
    uint8_t token[RelayTokenSize];
    uint8_t encryptedToken[RelayTokenSize + crypto_box_MACBYTES];
    if (crypto_box_easy(encryptedToken, token, sizeof(token), nonce, routerPublicKey, relayPrivateKey) != 0) {
      printf("could not encrypt relay token\n");
      return false;
    }

    char base64Token[sizeof(encryptedToken) * 2] = {};
    base64_encode_data(encryptedToken, sizeof(encryptedToken), base64Token, sizeof(base64Token));
    std::string base64TokenStr = base64Token;

    json::JSON doc;
    doc.set(InitRequestMagic, "magic_request_protection");
    doc.set(base64NonceStr, "nonce");
    doc.set(relayAddress, "relay_address");
    doc.set(port, "relay_port");
    doc.set(base64TokenStr, "encrypted_token");

    auto prettyStr = doc.toPrettyString();
    printf("sending this data to relay backend: %s\n", prettyStr.c_str());

    auto data = doc.toString();
    std::stringstream ss;
    ss << relayBackendAddr << "/relay_init_json";
    auto strurl = ss.str();
    if (!next_curl_send(strurl.c_str(), data.c_str(), data.length(), response)) {
      printf("could not send relay init\n");
      return false;
    }

    printf("response data: %s\n", response.c_str());

    return true;
  }

  inline auto next_curl_update(
   const char* relayBackendAddr, const char* inputJson, const std::string& relayAddress, int port, const char* relayName, json::JSON& respDoc) -> bool
  {
    json::JSON doc;
    if (!doc.parse(inputJson)) {
      printf("failed to reopen json for http update, need valid json");
      return false;
    }

    doc.set(relayAddress, "relay_address"); // here the relay address contains the port, unlike with init
    doc.set(relayName, "relay_name");

    auto prettyStr = doc.toPrettyString();
    printf("sending this data to the relay backend update endpoint: %s\n", prettyStr.c_str());

    std::string resp;
    auto data = doc.toString();
    std::stringstream ss;
    ss << relayBackendAddr << "/relay_update_json";
    auto strurl = ss.str();
    auto res = next_curl_send(strurl.c_str(), data.c_str(), data.size(), resp);

    printf("update response: %s\n", resp.c_str());

    if (!respDoc.parse(resp)) {
      printf("could not update json response");
      return false;
    }

    return res;
  }
}  // namespace compat