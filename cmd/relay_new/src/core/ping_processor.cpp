#include "includes.h"
#include "ping_processor.hpp"

#include "encoding/write.hpp"

namespace core
{
  PingProcessor::PingProcessor(os::Socket& socket, relay::relay_t& relay, volatile bool& handle)
   : mSocket(socket), mRelay(relay), mHandle(handle)
  {}

  PingProcessor::~PingProcessor()
  {
    stop();
  }

  void PingProcessor::listen()
  {
    while (this->mHandle) {
      relay::relay_platform_mutex_acquire(mRelay.mutex);

      if (mRelay.relays_dirty) {
        legacy::relay_manager_update(mRelay.relay_manager, mRelay.num_relays, mRelay.relay_ids, mRelay.relay_addresses);
        mRelay.relays_dirty = false;
      }

      double current_time = relay::relay_platform_time();

      struct ping_data_t
      {
        uint64_t sequence;
        legacy::relay_address_t address;
      };

      int num_pings = 0;
      ping_data_t pings[MAX_RELAYS];

      for (int i = 0; i < mRelay.relay_manager->num_relays; ++i) {
        if (mRelay.relay_manager->relay_last_ping_time[i] + RELAY_PING_TIME <= current_time) {
          pings[num_pings].sequence = relay_ping_history_ping_sent(mRelay.relay_manager->relay_ping_history[i], current_time);
          pings[num_pings].address = mRelay.relay_manager->relay_addresses[i];
          mRelay.relay_manager->relay_last_ping_time[i] = current_time;
          num_pings++;
        }
      }

      relay_platform_mutex_release(mRelay.mutex);

      for (int i = 0; i < num_pings; ++i) {
        uint8_t packet_data[9];
        packet_data[0] = RELAY_PING_PACKET;
        uint8_t* p = packet_data + 1;
        encoding::write_uint64(&p, pings[i].sequence);
        mSocket.send(pings[i].address, packet_data, 9);
      }

      relay::relay_platform_sleep(1.0 / 100.0);
    }
  }

  void PingProcessor::stop()
  {
    mHandle = false;
  }
}  // namespace core
