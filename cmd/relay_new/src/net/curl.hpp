#ifndef NET_CURL_HPP
#define NET_CURL_HPP

namespace net
{
  class CurlWrapper
  {
    CurlWrapper();
    ~CurlWrapper();

    CURL* mHandle;

   public:
    static bool SendTo(
     const std::string hostname, const std::string endpoint, const std::vector<uint8_t>& msg, std::vector<uint8_t>& resp);
  };
}  // namespace net
#endif
