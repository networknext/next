#include "includes.h"
#include "backend.hpp"

using namespace std::chrono_literals;

namespace legacy
{
  namespace v3
  {
    Backend::Backend(const net::Address& addr, os::Socket& socket): mAddr(addr), mSocket(socket) {}

    auto Backend::init() -> bool
    {
      util::JSON doc;
      if (!buildInitJSON(doc)) {
        Log("could not build v3 init json");
        return false;
      }

      std::string data = doc.toString();
      std::vector<uint8_t> buff(data.begin(), data.end());
      mSocket.send(mAddr, buff.data(), buff.size());

      core::GenericPacket<> packet;
      mSocket.recv(packet);

      std::string resp(packet.Buffer.begin() + 1, packet.Buffer.begin() + packet.Len);
      if (!doc.parse(resp)) {
        Log("v3 init resp parse error: ", doc.err());
        return false;
      }

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

    auto Backend::update() -> bool
    {
      util::JSON doc;

      if (!buildUpdateJSON(doc)) {
        Log("could not build v3 update json");
        return false;
      }

      std::string data = doc.toString();
      std::vector<uint8_t> buff(data.begin(), data.end());
      mSocket.send(mAddr, buff.data(), buff.size());

      core::GenericPacket<> packet;
      mSocket.recv(packet);

      std::string resp(packet.Buffer.begin() + 1, packet.Buffer.begin() + packet.Len);
      if (!doc.parse(resp)) {
        Log("v3 update resp parse error: ", doc.err());
        return false;
      }

      return true;
    }
  }  // namespace v3
}  // namespace legacy
