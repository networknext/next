#include "includes.h"
#include "v3_backend.hpp"

using namespace std::chrono_literals;

namespace core
{
  V3Backend::V3Backend(const net::Address& addr, os::Socket& socket, util::Channel<GenericPacket<>>& channel): mAddr(addr), mSocket(socket), mChannel(channel) {}

  auto V3Backend::init() -> bool
  {
    static std::string data = "foo";
    static std::vector<uint8_t> buff(data.begin(), data.end());
    mSocket.send(mAddr, buff.data(), buff.size());

    GenericPacket<> packet;
    mChannel.recv(packet);

    std::string resp(packet.Buffer.begin() + 1, packet.Buffer.begin() + packet.Len);

    LogDebug("\n\nInit: ", resp, "\n\n");

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
    mChannel.recv(packet);

    std::string resp(packet.Buffer.begin() + 1, packet.Buffer.begin() + packet.Len);

    LogDebug("\n\nUpdated: ", resp, "\n\n");

    return true;
  }
}  // namespace core
