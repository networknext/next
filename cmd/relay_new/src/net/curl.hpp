#ifndef NET_CURL_HPP
#define NET_CURL_HPP

#include "util/logger.hpp"

namespace net
{
  class CurlWrapper
  {
    CurlWrapper();
    ~CurlWrapper();

    CURL* mHandle;

    template <typename RespType>
    static size_t curlWriteFunction(char* ptr, size_t size, size_t nmemb, void* userdata)
    {
      auto respBuff = reinterpret_cast<RespType*>(userdata);
      auto index = respBuff->size();
      respBuff->resize(respBuff->size() + size * nmemb);
      std::copy(ptr, ptr + size * nmemb, respBuff->begin() + index);
      return size * nmemb;
    }

   public:
    template <typename ReqType, typename RespType>
    static bool SendTo(
     const std::string hostname, const std::string endpoint, const ReqType& request, RespType& response, size_t& bytesSent);
  };

  /*
   * Sends data to the specified hostname and endpoint
   * request can be anything that supplies ReqType::data() and ReqType::size()
   * response can be anything that supplies RespType::resize() and is compatable with std::copy()
   */
  template <typename ReqType, typename RespType>
  bool CurlWrapper::SendTo(
   const std::string hostname, const std::string endpoint, const ReqType& request, RespType& response, size_t& bytesSent)
  {
    static CurlWrapper wrapper;

    curl_slist* slist = curl_slist_append(nullptr, "Content-Type:application/json");

    std::stringstream ss;
    ss << hostname << endpoint;
    auto url = ss.str();

    curl_easy_setopt(wrapper.mHandle, CURLOPT_BUFFERSIZE, 102400L);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_URL, url.c_str());
    curl_easy_setopt(wrapper.mHandle, CURLOPT_NOPROGRESS, 1L);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_POSTFIELDS, request.data());
    curl_easy_setopt(wrapper.mHandle, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)request.size());
    curl_easy_setopt(wrapper.mHandle, CURLOPT_HTTPHEADER, slist);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_USERAGENT, "network next relay");
    curl_easy_setopt(wrapper.mHandle, CURLOPT_MAXREDIRS, 50L);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_TCP_KEEPALIVE, 1L);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_TIMEOUT_MS, long(1000));
    curl_easy_setopt(wrapper.mHandle, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(wrapper.mHandle, CURLOPT_WRITEFUNCTION, &curlWriteFunction<RespType>);

    CURLcode ret = curl_easy_perform(wrapper.mHandle);

    curl_slist_free_all(slist);
    slist = nullptr;

    if (ret != 0) {
      Log("curl request for '", hostname, endpoint, "' had an error: ", ret);
      return false;
    }

    if (!curl_easy_getinfo(wrapper.mHandle, CURLINFO_REQUEST_SIZE, &bytesSent)) {
      Log("curl could not get the last request size");  // non-critical failure, don't return false
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
#endif
