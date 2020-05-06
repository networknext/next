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

    auto Backend::init() -> bool
    {
      // prep

      LogDebug("building init json");
      util::JSON doc;
      if (!buildInitJSON(doc)) {
        Log("could not build v3 init json");
        return false;
      }

      std::string data = doc.toString();
      std::vector<uint8_t> buff(data.begin(), data.end());

      LogDebug("resolving backend addr");
      net::Address backendAddr;
      if (!backendAddr.resolve(mEnv.RelayV3BackendHostname, mEnv.RelayV3BackendPort)) {
        Log("Could not resolve the v3 backend hostname to an ip address");
        return 1;
      }

      BackendToken token;

      std::string initJSON = doc.toString();
      core::GenericPacket<> packet;
      std::copy(initJSON.begin(), initJSON.end(), packet.Buffer.begin());
      packet.Len = initJSON.length();
      packet.Addr = backendAddr;

      uint8_t attempts = 0;
      bool done = false;
      do {
        attempts++;
        // send request
        LogDebug("sending init packet") if (!packet_send(mSocket, backendAddr, token, PacketType::InitRequest, packet))
        {
          Log("failed to send init packet");
          return false;
        }

        // wait a second for the response to come in
        std::this_thread::sleep_for(1s);

        // receive response if it exists, if not resend
        LogDebug("checking for response");

        if (mReceiver.hasItems()) {
          mReceiver.recv(packet);
          done = true;
        }
      } while (!done && !mSocket.closed() && !mReceiver.closed() && attempts <= 60);

      if (mSocket.closed() || mReceiver.closed() || attempts > 60) {
        LogDebug("could not init relay");
        return false;
      }

      return true;

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
      return true;
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

      for (int i = 0; i < 250; i++) {
        std::stringstream ss;
        ss << "value" << i;
        auto str = ss.str();
        doc.set("init", str.c_str());
      }

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
