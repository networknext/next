#include "includes.h"
#include "curl.hpp"

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
}  // namespace net
