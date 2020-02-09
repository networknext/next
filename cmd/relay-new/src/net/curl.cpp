#include "curl.hpp"

#include <cassert>
#include <cstring>

namespace net
{
  size_t curl_buffer_write_function(char* ptr, size_t size, size_t nmemb, void* userdata)
  {
    curl_buffer_t* buffer = (curl_buffer_t*)userdata;
    assert(buffer);
    assert(size == 1);
    if (int(buffer->size + size * nmemb) > buffer->max_size)
      return 0;
    memcpy(buffer->data + buffer->size, ptr, size * nmemb);
    buffer->size += size * nmemb;
    return size * nmemb;
  }
}  // namespace net