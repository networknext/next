#pragma once
#include "net/http.hpp"
namespace testing
{
  class MockHttpClient: public net::IHttpClient
  {
   public:
    bool success = true;            // The request was a success
    std::vector<uint8_t> request;   // The request that was sent
    std::vector<uint8_t> response;  // The response that should be received
    std::string hostname;           // The hostname used
    std::string endpoint;           // The endpoint to hit

    auto send_request(
     const std::string hostname,
     const std::string endpoint,
     const std::vector<uint8_t>& request,
     std::vector<uint8_t>& response) -> bool override;
  };

  INLINE auto MockHttpClient::send_request(
   const std::string hname, const std::string endpt, const std::vector<uint8_t>& req, std::vector<uint8_t>& resp)
   -> bool
  {
    this->request.assign(req.begin(), req.end());
    resp.assign(this->response.begin(), this->response.end());
    this->hostname = hname;
    this->endpoint = endpt;
    return this->success;
  }
}  // namespace testing