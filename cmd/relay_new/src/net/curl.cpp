#include "includes.h"
#include "curl.hpp"

#include "util/logger.hpp"

namespace
{
  size_t curlWriteFunction(char* ptr, size_t size, size_t nmemb, void* userdata)
  {
    auto dataVec = reinterpret_cast<std::vector<uint8_t>*>(userdata);
    dataVec->resize(size * nmemb);
    std::copy(ptr, ptr + size * nmemb, dataVec->begin());
    return size * nmemb;
  }
}  // namespace

namespace net
{
  CurlWrapper::CurlWrapper()
  {
    curl_global_init(0);
    mHandle = curl_easy_init();
  }

  CurlWrapper::~CurlWrapper()
  {
    curl_easy_cleanup(mHandle);
    curl_global_cleanup();
  }

  bool CurlWrapper::SendTo(
   const std::string hostname, const std::string endpoint, const std::vector<uint8_t>& msg, std::vector<uint8_t>& resp)
  {
    static CurlWrapper wrapper;

    curl_slist* slist = curl_slist_append(nullptr, "Content-Type:application/json");

    std::stringstream ss;
    ss << hostname << endpoint;
    auto url = ss.str();

    curl_easy_setopt(wrapper.mHandle, CURLOPT_BUFFERSIZE, 102400L);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_URL, url.c_str());
    curl_easy_setopt(wrapper.mHandle, CURLOPT_NOPROGRESS, 1L);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_POSTFIELDS, msg.data());
    curl_easy_setopt(wrapper.mHandle, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)msg.size());
    curl_easy_setopt(wrapper.mHandle, CURLOPT_HTTPHEADER, slist);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_USERAGENT, "network next relay");
    curl_easy_setopt(wrapper.mHandle, CURLOPT_MAXREDIRS, 50L);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_TCP_KEEPALIVE, 1L);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_TIMEOUT_MS, long(1000));
    curl_easy_setopt(wrapper.mHandle, CURLOPT_WRITEDATA, &resp);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_WRITEFUNCTION, &curlWriteFunction);

    CURLcode ret = curl_easy_perform(wrapper.mHandle);

    curl_slist_free_all(slist);
    slist = nullptr;

    if (ret != 0) {
      Log("curl request for '", hostname, endpoint, "' had an error error: ", ret);
      return false;
    }

    long code = 0;
    curl_easy_getinfo(wrapper.mHandle, CURLINFO_RESPONSE_CODE, &code);
    if (code < 200 || code >= 300) {
      Log("http call to '", hostname, endpoint, "' did not return a success, code: ", code);
      return false;
    }

    return true;
  }
}  // namespace net
