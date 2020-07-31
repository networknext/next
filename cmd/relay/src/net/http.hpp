#pragma once

#include "util/logger.hpp"

namespace net
{
  namespace beast = boost::beast;
  namespace http = beast::http;
  namespace network = boost::asio;
  using tcp = network::ip::tcp;

  class BeastWrapper
  {
    BeastWrapper();
    ~BeastWrapper();

    // The io_context is required for all I/O
    network::io_context ioc;

    // These objects perform our I/O
    tcp::resolver resolver;
    beast::tcp_stream stream;

   public:
    static auto SendTo(
     const std::string hostname, const std::string endpoint, const std::string& request, std::string& response) -> bool;
  };

  inline BeastWrapper::BeastWrapper(): resolver(ioc), stream(ioc) {}

  inline BeastWrapper::~BeastWrapper()
  {
    // Gracefully close the socket
    beast::error_code ec;
    stream.socket().shutdown(tcp::socket::shutdown_both, ec);

    // not_connected happens sometimes
    // so don't bother reporting it

    if (ec && ec != beast::errc::not_connected) {
      Log("error shutting down socket");
    }
  }

  inline auto BeastWrapper::SendTo(
   const std::string hostname, const std::string endpoint, const std::string& request, std::string& response) -> bool
  {
    static BeastWrapper wrapper;

    try {
      std::string proto;
      std::string name;

      auto protopos = hostname.find_first_of(':');
      proto = hostname.substr(0, protopos);

      auto namepos = hostname.find_last_of('/');
      name = hostname.substr(namepos + 1);

      auto portpos = name.find_first_of(':');
      if (portpos != std::string::npos) {
        proto = name.substr(portpos + 1);
        name = name.substr(0, portpos);
      }

      // Look up the domain name
      auto const results = wrapper.resolver.resolve(name, proto);

      // Make the connection on the IP address we get from a lookup
      wrapper.stream.connect(results);

      // Set up an HTTP PUT request message
      http::request<http::string_body> req;
      req.method(http::verb::post);
      req.target(endpoint);
      req.version(11);
      req.content_length(request.size());
      req.body() = request;
      req.set(http::field::host, name);
      req.set(http::field::user_agent, "network next relay");
      req.set(http::field::content_type, "application/json");
      req.set(http::field::timeout, "10");

      // Send the HTTP request to the remote host
      beast::error_code ec;
      http::write(wrapper.stream, req, ec);

      if (ec) {
        Log("failed to send http request: ", ec);
        return false;
      }

      // This buffer is used for reading and must be persisted
      beast::flat_buffer buffer;

      // Declare a container to hold the response
      http::response<http::string_body> res;

      // Receive the HTTP response
      http::read(wrapper.stream, buffer, res, ec);
      if (ec) {
        Log("failed to send http request: ", ec);
        return false;
      }

      // Check the status code
      if (res.result() != http::status::ok) {
        Log("http call to '", hostname, endpoint, "' did not return a success, code: ", res.result_int());
        return false;
      }

      // Copy the response
      response = res.body().data();

      return true;
    } catch (std::exception& e) {
      Log("could not send http request: ", e.what());
      return false;
    }
  }

  // Previously curl was used for http communication, however it's a pain to link statically
  // If for some reason we have to go back, this was the code used

  // class CurlWrapper
  //{
  //  CurlWrapper();
  //  ~CurlWrapper();

  //  CURL* mHandle;

  //  template <typename RespType>
  //  static size_t curlWriteFunction(char* ptr, size_t size, size_t nmemb, void* userdata)
  //  {
  //    auto respBuff = reinterpret_cast<RespType*>(userdata);
  //    auto index = respBuff->size();
  //    respBuff->resize(respBuff->size() + size * nmemb);
  //    std::copy(ptr, ptr + size * nmemb, respBuff->begin() + index);
  //    return size * nmemb;
  //  }

  // public:
  //  template <typename ReqType, typename RespType>
  //  static bool SendTo(const std::string hostname, const std::string endpoint, const ReqType& request, RespType& response);
  //};

  // inline CurlWrapper::CurlWrapper()
  //{
  //  curl_global_init(0);
  //  mHandle = curl_easy_init();
  //}

  // inline CurlWrapper::~CurlWrapper()
  //{
  //  curl_easy_cleanup(mHandle);
  //  curl_global_cleanup();
  //}

  ///*
  // * Sends data to the specified hostname and endpoint
  // * request can be anything that supplies ReqType::data() and ReqType::size()
  // * response can be anything that supplies RespType::resize() and is compatable with std::copy()
  // */
  // template <typename ReqType, typename RespType>
  // bool CurlWrapper::SendTo(const std::string hostname, const std::string endpoint, const ReqType& request, RespType&
  // response)
  //{
  //  static CurlWrapper wrapper;

  //  curl_slist* slist = curl_slist_append(nullptr, "Content-Type:application/json");

  //  std::stringstream ss;
  //  ss << hostname << endpoint;
  //  auto url = ss.str();

  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_BUFFERSIZE, 102400L);
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_URL, url.c_str());
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_NOPROGRESS, 1L);
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_POSTFIELDS, request.data());
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)request.size());
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_HTTPHEADER, slist);
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_USERAGENT, "network next relay");
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_MAXREDIRS, 50L);
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS);
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_TCP_KEEPALIVE, 1L);
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_TIMEOUT_MS, long(10000));
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_WRITEDATA, &response);
  //  curl_easy_setopt(wrapper.mHandle, CURLOPT_WRITEFUNCTION, &curlWriteFunction<RespType>);

  //  CURLcode ret = curl_easy_perform(wrapper.mHandle);

  //  curl_slist_free_all(slist);
  //  slist = nullptr;

  //  if (ret != 0) {
  //    Log("curl request for '", hostname, endpoint, "' had an error: ", ret);
  //    return false;
  //  }

  //  long code = 0;
  //  curl_easy_getinfo(wrapper.mHandle, CURLINFO_RESPONSE_CODE, &code);
  //  if (code < 200 || code >= 300) {
  //    Log("http call to '", hostname, endpoint, "' did not return a success, code: ", code);
  //    return false;
  //  }

  //  return true;
  //}

}  // namespace net
