#pragma once
#include "net/http.hpp"
namespace testing
{
  class MockHttpClient: public net::IHttpClient
  {
   public:
    bool Success = true;            // The request was a success
    std::vector<uint8_t> Request;   // The request that was sent
    std::vector<uint8_t> Response;  // The Response that should be received
    std::string Hostname;           // The hostname used
    std::string Endpoint;           // The endpoint to hit

    auto send_request(
     const std::string hostname,
     const std::string endpoint,
     const std::vector<uint8_t>& request,
     std::vector<uint8_t>& response) -> bool override;
  };

  INLINE auto MockHttpClient::send_request(
   const std::string hostname, const std::string endpoint, const std::vector<uint8_t>& request, std::vector<uint8_t>& response)
   -> bool
  {
    Request.assign(request.begin(), request.end());
    response.assign(Response.begin(), Response.end());
    Hostname = hostname;
    Endpoint = endpoint;
    return Success;
  }
}  // namespace testing