#ifndef NET_CURL_HPP
#define NET_CURL_HPP

#include <cstddef>
#include <cinttypes>

namespace net
{
  struct curl_buffer_t
  {
    int size;
    int max_size;
    uint8_t* data;
  };

  size_t curl_buffer_write_function(char* ptr, size_t size, size_t nmemb, void* userdata);
}  // namespace net
#endif