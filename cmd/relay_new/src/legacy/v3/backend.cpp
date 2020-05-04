#include "includes.h"
#include "backend.hpp"
#include "packet_send.hpp"

using namespace std::chrono_literals;

namespace
{
  const uint8_t PacketType = 123;
}

namespace legacy
{
  namespace v3
  {
    Backend::Backend(util::Receiver<core::GenericPacket<>>& receiver, util::Env& env, os::Socket& socket)
     : mReceiver(receiver), mEnv(env), mSocket(socket)
    {}

    auto Backend::init() -> bool
    {
      bool success = false;
      for (int i = 0; i < 60; i++) {
        if (tryInit()) {
          success = true;
          break;
        }
      }
      return success;
    }

    // clang-format off
    /*
     * Init response appears to be
     *
     *{
     *  "Timestamp": number, // unsigned, millisecond resolution, needs to be converted to nanos and have RTT compensation (t = backend_timestamp / 1000000 - (received - requested) / 2)
     *  "Token": string // base64 composed of an address (which one?) and hmac (of the address?)
     *}
     *
     * Config response appears to be
     *
     *{
     *  "Group": string // read-only once it's saved
     *}
     */
    // clang-format on

    auto Backend::tryInit() -> bool
    {
      // prep

      util::JSON doc;
      if (!buildInitJSON(doc)) {
        Log("could not build v3 init json");
        return false;
      }

      std::string data = doc.toString();
      std::vector<uint8_t> buff(data.begin(), data.end());

      net::Address backendAddr;
      if (!backendAddr.resolve(mEnv.RelayV3BackendHostname, mEnv.RelayV3BackendPort)) {
        Log("Could not resolve the v3 backend hostname to an ip address");
        return 1;
      }

      BackendToken token;
      BackendRequest request;

      core::GenericPacket<> packet;

      // send request
      if (!packet_send(mSocket, token.Address, token, PacketType::InitRequest, request, packet)) {
        Log("failed to send init packet");
        return false;
      }

      // wait a seconds for the response to come in
      std::this_thread::sleep_for(1s);

      // receive response
      mReceiver.recv(packet);

      std::string resp(packet.Buffer.begin() + 1, packet.Buffer.begin() + packet.Len);
      if (!doc.parse(resp)) {
        Log("v3 init resp parse error: ", doc.err());
        return false;
      }

      // process response

      LogDebug("hey shit worked");

      return true;
    }

    auto Backend::updateCycle(const volatile bool& handle) -> bool
    {
      while (handle) {
        if (!update()) {
          return false;
        }

        std::this_thread::sleep_for(10s);
      }

      return true;
    }

    /*
     * Update response appears to be
     *
     *{
     *  "PingTargets": [
     *  {
     *    "Address": string, // address of the relay
     *    "Id": number, // id of the relay -- not going to be the same ids as the new backend
     *    "Group": string, // not entirely sure
     *    "PingToken": string // base64 token composed of ??
     *  },
     *  ...
     * ]
     *}
     */

    auto Backend::update() -> bool
    {
      util::JSON doc;

      if (!buildUpdateJSON(doc)) {
        Log("could not build v3 update json");
        return false;
      }

      std::string data = doc.toString();
      std::vector<uint8_t> buff(data.begin(), data.end());

      core::GenericPacket<> packet;
      mSocket.recv(packet);

      std::string resp(packet.Buffer.begin() + 1, packet.Buffer.begin() + packet.Len);
      if (!doc.parse(resp)) {
        Log("v3 update resp parse error: ", doc.err());
        return false;
      }

      return true;
    }

    auto Backend::buildInitJSON(util::JSON& doc) -> bool
    {
      doc.set("init", "value");
      return true;
    }

    auto Backend::buildConfigJSON(util::JSON& doc) -> bool
    {
      doc.set("config", "value");
      return true;
    }

    auto Backend::buildUpdateJSON(util::JSON& doc) -> bool
    {
      doc.set("update", "value");
      return true;
    }
  }  // namespace v3
}  // namespace legacy
