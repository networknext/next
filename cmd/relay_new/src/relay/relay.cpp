#include "includes.h"
#include "relay.hpp"

#include "encoding/binary.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"

#include "util.hpp"

#include "net/curl.hpp"

#include "core/relay_stats.hpp"

int relay_debug = 0;
namespace relay
{
  int relay_initialize()
  {
    if (relay::relay_platform_init() != RELAY_OK) {
      Log("failed to initialize platform");
      return RELAY_ERROR;
    }

    if (sodium_init() == -1) {
      Log("failed to initialize sodium");
      return RELAY_ERROR;
    }

    const char* relay_debug_env = relay::relay_platform_getenv("RELAY_DEBUG");
    if (relay_debug_env) {
      // TODO replace this flag with a makefile compile-time define
      relay_debug = atoi(relay_debug_env);
    }

    return RELAY_OK;
  }

  void relay_term()
  {
    relay::relay_platform_term();
  }

  int relay_init(CURL* curl,
   const char* hostname,
   uint8_t* relay_token,
   const char* relay_address,
   const uint8_t* router_public_key,
   const uint8_t* relay_private_key,
   uint64_t* router_timestamp)
  {
    const uint32_t init_request_magic = 0x9083708f;

    uint32_t init_request_version = 0;

    uint8_t init_data[1024];
    memset(init_data, 0, sizeof(init_data));

    unsigned char nonce[crypto_box_NONCEBYTES];
    encoding::relay_random_bytes(nonce, crypto_box_NONCEBYTES);

    uint8_t* p = init_data;

    encoding::write_uint32(&p, init_request_magic);
    encoding::write_uint32(&p, init_request_version);
    encoding::write_bytes(&p, nonce, crypto_box_NONCEBYTES);
    encoding::write_string(&p, relay_address, RELAY_MAX_ADDRESS_STRING_LENGTH);

    uint8_t* q = p;

    encoding::write_bytes(&p, relay_token, RELAY_TOKEN_BYTES);

    int encrypt_length = int(p - q);

    if (crypto_box_easy(q, q, encrypt_length, nonce, router_public_key, relay_private_key) != 0) {
      LogDebug("could not encrypt relay token");
      return RELAY_ERROR;
    }

    int init_length = (int)(p - init_data) + encrypt_length + crypto_box_MACBYTES;

    struct curl_slist* slist = curl_slist_append(NULL, "Content-Type:application/octet-stream");

    net::curl_buffer_t init_response_buffer;
    init_response_buffer.size = 0;
    init_response_buffer.max_size = 1024;
    init_response_buffer.data = (uint8_t*)alloca(init_response_buffer.max_size);

    char init_url[1024];
    sprintf(init_url, "%s/relay_init", hostname);

    curl_easy_setopt(curl, CURLOPT_BUFFERSIZE, 102400L);
    curl_easy_setopt(curl, CURLOPT_URL, init_url);
    curl_easy_setopt(curl, CURLOPT_NOPROGRESS, 1L);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, init_data);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)init_length);
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, slist);
    curl_easy_setopt(curl, CURLOPT_USERAGENT, "network next relay");
    curl_easy_setopt(curl, CURLOPT_MAXREDIRS, 50L);
    curl_easy_setopt(curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS);
    curl_easy_setopt(curl, CURLOPT_TCP_KEEPALIVE, 1L);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT_MS, long(1000));
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &init_response_buffer);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, &net::curl_buffer_write_function);

    CURLcode ret = curl_easy_perform(curl);

    curl_slist_free_all(slist);
    slist = NULL;

    if (ret != 0) {
      LogDebug("curl error: ", ret);
      return RELAY_ERROR;
    }

    long code;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &code);
    if (code != 200) {
      LogDebug("http call not success, code: ", code);
      return RELAY_ERROR;
    }

    if (init_response_buffer.size < 4) {
      Log("error: bad relay init response size. too small to have valid data (", init_response_buffer.size, ")");
      return RELAY_ERROR;
    }

    const uint8_t* r = init_response_buffer.data;

    uint32_t version = encoding::read_uint32(&r);

    const uint32_t init_response_version = 0;

    if (version != init_response_version) {
      Log("error: bad relay init response version. expected ", init_response_version, ", got ", version);
      return RELAY_ERROR;
    }

    if (init_response_buffer.size != 4 + 8 + RELAY_TOKEN_BYTES) {
      Log("error: bad relay init response size. expected ", RELAY_TOKEN_BYTES, " bytes, got ", init_response_buffer.size);
      return RELAY_ERROR;
    }

    *router_timestamp = encoding::read_uint64(&r);

    memcpy(relay_token, init_response_buffer.data + 4 + 8, RELAY_TOKEN_BYTES);

    return RELAY_OK;
  }

  int relay_update(CURL* curl,
   const char* hostname,
   const uint8_t* relay_token,
   const char* relay_address,
   uint8_t* update_response_memory,
   core::RelayManager& manager)
  {
    // build update data

    uint32_t update_version = 0;

    uint8_t update_data[10 * 1024];  // TODO pass this in like response memory is

    uint8_t* p = update_data;
    encoding::write_uint32(&p, update_version);
    encoding::write_string(&p, relay_address, 256);
    encoding::write_bytes(&p, relay_token, RELAY_TOKEN_BYTES);

    core::RelayStats stats;
    manager.getStats(stats);

    encoding::write_uint32(&p, stats.NumRelays);
    for (unsigned int i = 0; i < stats.NumRelays; ++i) {
      encoding::write_uint64(&p, stats.IDs[i]);
      encoding::write_float32(&p, stats.RTT[i]);
      encoding::write_float32(&p, stats.Jitter[i]);
      encoding::write_float32(&p, stats.PacketLoss[i]);
    }

    int update_data_length = (int)(p - update_data);

    // post it to backend

    struct curl_slist* slist = curl_slist_append(NULL, "Content-Type:application/octet-stream");

    net::curl_buffer_t update_response_buffer;
    update_response_buffer.size = 0;
    update_response_buffer.max_size = RESPONSE_MAX_BYTES;
    update_response_buffer.data = (uint8_t*)update_response_memory;

    char update_url[1024];
    sprintf(update_url, "%s/relay_update", hostname);

    curl_easy_setopt(curl, CURLOPT_BUFFERSIZE, 102400L);
    curl_easy_setopt(curl, CURLOPT_URL, update_url);
    curl_easy_setopt(curl, CURLOPT_NOPROGRESS, 1L);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, update_data);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)update_data_length);
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, slist);
    curl_easy_setopt(curl, CURLOPT_USERAGENT, "network next relay");
    curl_easy_setopt(curl, CURLOPT_MAXREDIRS, 50L);
    curl_easy_setopt(curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS);
    curl_easy_setopt(curl, CURLOPT_TCP_KEEPALIVE, 1L);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT_MS, long(1000));
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &update_response_buffer);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, &net::curl_buffer_write_function);

    CURLcode ret = curl_easy_perform(curl);

    curl_slist_free_all(slist);
    slist = NULL;

    if (ret != 0) {
      Log("error: could not post relay update. curl error: ", ret, '\n');
      return RELAY_ERROR;
    }

    long code;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &code);
    if (code != 200) {
      Log("error: relay update response was ", code, ", expected 200\n");
      return RELAY_ERROR;
    }

    // parse update response

    const uint8_t* q = update_response_buffer.data;

    uint32_t version = encoding::read_uint32(&q);

    const uint32_t update_response_version = 0;

    if (version != update_response_version) {
      Log("error: bad relay update response version. expected ", update_response_version, ", got ", version, '\n');
      return RELAY_ERROR;
    }

    uint32_t num_relays = encoding::read_uint32(&q);

    if (num_relays > MAX_RELAYS) {
      Log("error: too many relays to ping. max is ", MAX_RELAYS, ", got ", num_relays, '\n');
      return RELAY_ERROR;
    }

    bool error = false;

    struct relay_ping_data_t
    {
      uint64_t id;
      net::Address address;
    };

    relay_ping_data_t relay_ping_data[MAX_RELAYS];

    for (uint32_t i = 0; i < num_relays; ++i) {
      char address_string[RELAY_MAX_ADDRESS_STRING_LENGTH];
      relay_ping_data[i].id = encoding::read_uint64(&q);
      encoding::read_string(&q, address_string, RELAY_MAX_ADDRESS_STRING_LENGTH);
      if (!relay_ping_data[i].address.parse(address_string)) {
        error = true;
        break;
      }
    }

    if (error) {
      Log("error: error while reading set of relays to ping in update response\n");
      return RELAY_ERROR;
    }

    // TODO can avoid this loop entirely by just moving ping data to it's own file and having manager take it as a param
    std::array<uint64_t, MAX_RELAYS> relayIDs;
    std::array<net::Address, MAX_RELAYS> relayAddresses;
    for (unsigned int i = 0; i < num_relays; ++i) {
      relayIDs[i] = relay_ping_data[i].id;
      relayAddresses[i] = relay_ping_data[i].address;
    }

    manager.update(num_relays, relayIDs, relayAddresses);

    return RELAY_OK;
  }

  int relay_write_header(int direction,
   uint8_t type,
   uint64_t sequence,
   uint64_t session_id,
   uint8_t session_version,
   const uint8_t* private_key,
   uint8_t* buffer,
   int buffer_length)
  {
    assert(private_key);
    assert(buffer);
    assert(RELAY_HEADER_BYTES <= buffer_length);

    (void)buffer_length;

    uint8_t* start = buffer;

    (void)start;

    if (direction == RELAY_DIRECTION_SERVER_TO_CLIENT) {
      // high bit must be set
      assert(sequence & (1ULL << 63));
    } else {
      // high bit must be clear
      assert((sequence & (1ULL << 63)) == 0);
    }

    if (type == RELAY_SESSION_PING_PACKET || type == RELAY_SESSION_PONG_PACKET || type == RELAY_ROUTE_RESPONSE_PACKET ||
        type == RELAY_CONTINUE_RESPONSE_PACKET) {
      // second highest bit must be set
      assert(sequence & (1ULL << 62));
    } else {
      // second highest bit must be clear
      assert((sequence & (1ULL << 62)) == 0);
    }

    encoding::write_uint8(&buffer, type);

    encoding::write_uint64(&buffer, sequence);

    uint8_t* additional = buffer;
    const int additional_length = 8 + 2;

    encoding::write_uint64(&buffer, session_id);
    encoding::write_uint8(&buffer, session_version);
    encoding::write_uint8(&buffer, 0);  // todo: remove this once we fully switch to new relay

    uint8_t nonce[12];
    {
      uint8_t* p = nonce;
      encoding::write_uint32(&p, 0);
      encoding::write_uint64(&p, sequence);
    }

    unsigned long long encrypted_length = 0;

    int result = crypto_aead_chacha20poly1305_ietf_encrypt(
     buffer, &encrypted_length, buffer, 0, additional, (unsigned long long)additional_length, NULL, nonce, private_key);

    if (result != 0)
      return RELAY_ERROR;

    buffer += encrypted_length;

    assert(int(buffer - start) == RELAY_HEADER_BYTES);

    return RELAY_OK;
  }

  int relay_peek_header(int direction,
   uint8_t* type,
   uint64_t* sequence,
   uint64_t* session_id,
   uint8_t* session_version,
   const uint8_t* buffer,
   int buffer_length)
  {
    uint8_t packet_type;
    uint64_t packet_sequence;

    assert(buffer);

    if (buffer_length < RELAY_HEADER_BYTES)
      return RELAY_ERROR;

    packet_type = encoding::read_uint8(&buffer);

    packet_sequence = encoding::read_uint64(&buffer);

    if (direction == RELAY_DIRECTION_SERVER_TO_CLIENT) {
      // high bit must be set
      if (!(packet_sequence & (1ULL << 63)))
        return RELAY_ERROR;
    } else {
      // high bit must be clear
      if (packet_sequence & (1ULL << 63))
        return RELAY_ERROR;
    }

    *type = packet_type;

    if (*type == RELAY_SESSION_PING_PACKET || *type == RELAY_SESSION_PONG_PACKET || *type == RELAY_ROUTE_RESPONSE_PACKET ||
        *type == RELAY_CONTINUE_RESPONSE_PACKET) {
      // second highest bit must be set
      assert(packet_sequence & (1ULL << 62));
    } else {
      // second highest bit must be clear
      assert((packet_sequence & (1ULL << 62)) == 0);
    }

    *sequence = packet_sequence;
    *session_id = encoding::read_uint64(&buffer);
    *session_version = encoding::read_uint8(&buffer);

    return RELAY_OK;
  }

  int relay_verify_header(int direction, const uint8_t* private_key, uint8_t* buffer, int buffer_length)
  {
    assert(private_key);
    assert(buffer);

    if (buffer_length < RELAY_HEADER_BYTES) {
      return RELAY_ERROR;
    }

    const uint8_t* p = buffer;

    uint8_t packet_type = encoding::read_uint8(&p);

    uint64_t packet_sequence = encoding::read_uint64(&p);

    if (direction == RELAY_DIRECTION_SERVER_TO_CLIENT) {
      // high bit must be set
      if (!(packet_sequence & (1ULL << 63))) {
        return RELAY_ERROR;
      }
    } else {
      // high bit must be clear
      if (packet_sequence & (1ULL << 63)) {
        return RELAY_ERROR;
      }
    }

    if (packet_type == RELAY_SESSION_PING_PACKET || packet_type == RELAY_SESSION_PONG_PACKET ||
        packet_type == RELAY_ROUTE_RESPONSE_PACKET || packet_type == RELAY_CONTINUE_RESPONSE_PACKET) {
      // second highest bit must be set
      assert(packet_sequence & (1ULL << 62));
    } else {
      // second highest bit must be clear
      assert((packet_sequence & (1ULL << 62)) == 0);
    }

    const uint8_t* additional = p;

    const int additional_length = 8 + 2;

    uint64_t packet_session_id = encoding::read_uint64(&p);
    uint8_t packet_session_version = encoding::read_uint8(&p);
    uint8_t packet_session_flags = encoding::read_uint8(&p);  // todo: remove once we fully switch over to new relay

    (void)packet_session_id;
    (void)packet_session_version;
    (void)packet_session_flags;

    uint8_t nonce[12];
    {
      uint8_t* q = nonce;
      encoding::write_uint32(&q, 0);
      encoding::write_uint64(&q, packet_sequence);
    }

    unsigned long long decrypted_length;

    int result = crypto_aead_chacha20poly1305_ietf_decrypt(buffer + 19,
     &decrypted_length,
     NULL,
     buffer + 19,
     (unsigned long long)crypto_aead_chacha20poly1305_IETF_ABYTES,
     additional,
     (unsigned long long)additional_length,
     nonce,
     private_key);

    if (result != 0) {
      return RELAY_ERROR;
    }

    return RELAY_OK;
  }

  uint64_t relay_clean_sequence(uint64_t sequence)
  {
    uint64_t mask = ~((1ULL << 63) | (1ULL << 62));
    return sequence & mask;
  }
}  // namespace relay
