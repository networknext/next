#include "includes.h"
#include "v3_backend.hpp"

using namespace std::chrono_literals;

namespace core
{
  V3Backend::V3Backend(const net::Address& addr, os::Socket& socket): mAddr(addr), mSocket(socket) {}

  auto V3Backend::init() -> bool
  {
    static std::string data = "foo";
    static std::vector<uint8_t> buff(data.begin(), data.end());
    mSocket.send(mAddr, buff.data(), buff.size());

    GenericPacket<> packet;
    mSocket.recv(packet);

    std::string resp(packet.Buffer.begin() + 1, packet.Buffer.begin() + packet.Len);

    LogDebug("Init: ", resp);

    return true;
  }

  auto V3Backend::updateCycle(const volatile bool& handle) -> bool
  {
    while (handle) {
      if (!update()) {
        return false;
      }

      std::this_thread::sleep_for(10s);
    }

    return true;
  }

  auto V3Backend::update() -> bool
  {
    static std::string data = "bar";
    static std::vector<uint8_t> buff(data.begin(), data.end());
    mSocket.send(mAddr, buff.data(), buff.size());

    GenericPacket<> packet;
    mSocket.recv(packet);

    std::string resp(packet.Buffer.begin() + 1, packet.Buffer.begin() + packet.Len);

    LogDebug("Updated: ", resp);

    return true;
  }
}  // namespace core
