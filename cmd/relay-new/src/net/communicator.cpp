#include "includes.h"
#include "communicator.hpp"

#include "util.hpp"
#include "util/logger.hpp"
#include "util/throughput_logger.hpp"

#include "core/ping_history.hpp"
#include "core/replay_protection.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

#include "relay/relay_continue_token.hpp"
#include "relay/relay_platform.hpp"
#include "relay/relay_route_token.hpp"

namespace net
{
  Communicator::Communicator(os::Socket& socket, relay::relay_t& relay, volatile bool& handle, std::ostream& output)
   : mSocket(socket), mRelay(relay), mHandle(handle), mLogger(output)
  {
    initPingThread();
    initRecvThread();
  }

  Communicator::~Communicator()
  {
    stop();
  }

  void Communicator::stop()
  {
    mHandle = false;
    mPingThread->join();
    mRecvThread->join();
    mLogger.stop();
  }

  void Communicator::initPingThread()
  {
    mPingThread = std::make_unique<std::thread>([this] {
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
    });
  }

  void Communicator::initRecvThread()
  {
    mRecvThread = std::make_unique<std::thread>([this] {
      std::array<uint8_t, RELAY_MAX_PACKET_BYTES> packetData;

      while (this->mHandle) {
        legacy::relay_address_t from;
        const int packet_bytes = mSocket.recv(from, packetData.data(), sizeof(uint8_t) * packetData.size());

        if (packet_bytes == 0) {
          mLogger.addToPkt0();
          continue;
        } else if (packetData[0] == RELAY_PING_PACKET && packet_bytes == 9) {
          this->handleRelayPingPacket(packetData, packet_bytes, from);
        } else if (packetData[0] == RELAY_PONG_PACKET && packet_bytes == 9) {
          this->handleRelayPongPacket(packetData, packet_bytes, from);
        } else if (packetData[0] == RELAY_ROUTE_REQUEST_PACKET) {
          this->handleRouteRequestPacket(packetData, packet_bytes, from);
        } else if (packetData[0] == RELAY_ROUTE_RESPONSE_PACKET) {
          this->handleRouteResponsePacket(packetData, packet_bytes);
        } else if (packetData[0] == RELAY_CONTINUE_REQUEST_PACKET) {
          this->handleContinueRequestPacket(packetData, packet_bytes);
        } else if (packetData[0] == RELAY_CONTINUE_RESPONSE_PACKET) {
          this->handleContinueResponsePacket(packetData, packet_bytes);
        } else if (packetData[0] == RELAY_CLIENT_TO_SERVER_PACKET) {
          this->handleClientToServerPacket(packetData, packet_bytes);
        } else if (packetData[0] == RELAY_SERVER_TO_CLIENT_PACKET) {
          this->handleServerToClientPacket(packetData, packet_bytes);
        } else if (packetData[0] == RELAY_SESSION_PING_PACKET) {
          this->handleSessionPingPacket(packetData, packet_bytes);
        } else if (packetData[0] == RELAY_SESSION_PONG_PACKET) {
          this->handleSessionPongPacket(packetData, packet_bytes);
        } else if (packetData[0] == RELAY_NEAR_PING_PACKET) {
          this->handleNearPingPacket(packetData, packet_bytes, from);
        } else {
          LogDebug("Received unknown packet type: ", std::hex, (int)packetData[0]);
          mLogger.addToUnknown(packet_bytes);
        }
      }
    });
  }

  void Communicator::handleRelayPingPacket(
   std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from)
  {
    mLogger.addToRelayPingPacket(size);

    // mark the 0'th index as a pong and send it back from where it came
    packet[0] = RELAY_PONG_PACKET;
    if (!mSocket.send(from, packet.data(), 9)) {
      Log("Failed to send data");
    }
  }

  void Communicator::handleRelayPongPacket(
   std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from)
  {
    mLogger.addToRelayPongPacket(size);
    relay_platform_mutex_acquire(mRelay.mutex);

    // read the uint from the packet - this could be brought out of the mutex
    const uint8_t* p = packet.data() + 1;
    uint64_t sequence = encoding::read_uint64(&p);

    // process the pong time
    relay_manager_process_pong(mRelay.relay_manager, &from, sequence);
    relay_platform_mutex_release(mRelay.mutex);
  }

  void Communicator::handleRouteRequestPacket(
   std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from)
  {
    mLogger.addToRouteReq(size);

    if (size < int(1 + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES * 2)) {
      relay_printf("ignoring route request. bad packet size (%d)", size);
      return;
    }

    // ignore the header byte of the packet
    uint8_t* p = &packet[1];
    relay::relay_route_token_t token;

    if (relay::relay_read_encrypted_route_token(&p, &token, mRelay.router_public_key, mRelay.relay_private_key) != RELAY_OK) {
      relay_printf("ignoring route request. could not read route token");
      return;
    }

    // don't do anything if the token is expired - probably should log something here
    if (token.expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    // create a new session and add it to the session map
    uint64_t hash = token.session_id ^ token.session_version;
    if (mRelay.sessions->find(hash) == mRelay.sessions->end()) {
      // create the session
      relay::relay_session_t* session = (relay::relay_session_t*)malloc(sizeof(relay::relay_session_t));
      assert(session);

      // fill it with data in the token
      session->expire_timestamp = token.expire_timestamp;
      session->session_id = token.session_id;
      session->session_version = token.session_version;
      session->client_to_server_sequence = 0;
      session->server_to_client_sequence = 0;
      session->kbps_up = token.kbps_up;
      session->kbps_down = token.kbps_down;
      session->prev_address = from;
      session->next_address = token.next_address;

      // store it
      memcpy(session->private_key, token.private_key, crypto_box_SECRETKEYBYTES);
      relay_replay_protection_reset(&session->replay_protection_client_to_server);
      relay_replay_protection_reset(&session->replay_protection_server_to_client);
      mRelay.sessions->insert(std::make_pair(hash, session));  // TODO dear god why

      printf("session created: %" PRIx64 ".%d\n", token.session_id, token.session_version);
    }

    // remove this part of the token by offseting it the request packet bytes
    packet[RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES] = RELAY_ROUTE_REQUEST_PACKET;
    mSocket.send(
     token.next_address, packet.data() + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES, size - RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES);
  }

  void Communicator::handleRouteResponsePacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    mLogger.addToRouteResp(size);
    if (size != RELAY_HEADER_BYTES) {
      return;
    }

    uint8_t type;
    uint64_t sequence;
    uint64_t session_id;
    uint8_t session_version;
    if (relay::relay_peek_header(
         RELAY_DIRECTION_SERVER_TO_CLIENT, &type, &sequence, &session_id, &session_version, packet.data(), size) != RELAY_OK) {
      return;
    }

    uint64_t hash = session_id ^ session_version;
    relay::relay_session_t* session = (*(mRelay.sessions))[hash];  // TODO just use a reference for this
    if (!session) {                                                // TODO and use find() for this
      return;
    }

    if (session->expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->server_to_client_sequence) {
      return;
    }

    session->server_to_client_sequence = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }

    mSocket.send(session->prev_address, packet.data(), size);
  }

  void Communicator::handleContinueRequestPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    mLogger.addToContReq(size);
    if (size < int(1 + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES * 2)) {
      relay_printf("ignoring continue request. bad packet size (%d)", size);
      return;
    }
    uint8_t* p = &packet[1];
    relay::relay_continue_token_t token;
    if (relay_read_encrypted_continue_token(&p, &token, mRelay.router_public_key, mRelay.relay_private_key) != RELAY_OK) {
      relay_printf("ignoring continue request. could not read continue token");
      return;
    }
    if (token.expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }
    uint64_t hash = token.session_id ^ token.session_version;
    relay::relay_session_t* session = (*(mRelay.sessions))[hash];
    if (!session) {
      return;
    }
    if (session->expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }
    if (session->expire_timestamp != token.expire_timestamp) {
      printf("session continued: %" PRIx64 ".%d\n", token.session_id, token.session_version);
    }
    session->expire_timestamp = token.expire_timestamp;
    packet[RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES] = RELAY_CONTINUE_REQUEST_PACKET;
    mSocket.send(
     session->next_address, packet.data() + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES, size - RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES);
  }

  void Communicator::handleContinueResponsePacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    mLogger.addToContResp(size);
    if (size != RELAY_HEADER_BYTES) {
      return;
    }
    uint8_t type;
    uint64_t sequence;
    uint64_t session_id;
    uint8_t session_version;
    if (relay::relay_peek_header(
         RELAY_DIRECTION_SERVER_TO_CLIENT, &type, &sequence, &session_id, &session_version, packet.data(), size) != RELAY_OK) {
      return;
    }
    uint64_t hash = session_id ^ session_version;
    relay::relay_session_t* session = (*(mRelay.sessions))[hash];
    if (!session) {
      return;
    }
    if (session->expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }
    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->server_to_client_sequence) {
      return;
    }
    session->server_to_client_sequence = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }
    mSocket.send(session->prev_address, packet.data(), size);
  }

  void Communicator::handleClientToServerPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    mLogger.addToCliToServ(size);

    if (size <= RELAY_HEADER_BYTES || size > RELAY_HEADER_BYTES + RELAY_MTU) {
      return;
    }

    uint8_t type;
    uint64_t sequence;
    uint64_t session_id;
    uint8_t session_version;
    if (relay::relay_peek_header(
         RELAY_DIRECTION_CLIENT_TO_SERVER, &type, &sequence, &session_id, &session_version, packet.data(), size) != RELAY_OK) {
      return;
    }

    uint64_t hash = session_id ^ session_version;
    relay::relay_session_t* session = (*(mRelay.sessions))[hash];
    if (!session) {
      return;
    }

    if (session->expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (relay_replay_protection_already_received(&session->replay_protection_client_to_server, clean_sequence)) {
      return;
    }

    relay_replay_protection_advance_sequence(&session->replay_protection_client_to_server, clean_sequence);
    if (relay::relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }

    mSocket.send(session->next_address, packet.data(), size);
  }

  void Communicator::handleServerToClientPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    mLogger.addToServToCli(size);
    if (size <= RELAY_HEADER_BYTES || size > RELAY_HEADER_BYTES + RELAY_MTU) {
      return;
    }

    uint8_t type;
    uint64_t sequence;
    uint64_t session_id;
    uint8_t session_version;
    if (relay::relay_peek_header(
         RELAY_DIRECTION_SERVER_TO_CLIENT, &type, &sequence, &session_id, &session_version, packet.data(), size) != RELAY_OK) {
      return;
    }

    uint64_t hash = session_id ^ session_version;
    relay::relay_session_t* session = (*(mRelay.sessions))[hash];
    if (!session) {
      return;
    }

    if (session->expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (relay_replay_protection_already_received(&session->replay_protection_server_to_client, clean_sequence)) {
      return;
    }

    relay_replay_protection_advance_sequence(&session->replay_protection_server_to_client, clean_sequence);
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }

    mSocket.send(session->prev_address, packet.data(), size);
  }

  void Communicator::handleSessionPingPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    mLogger.addToSessionPing(size);

    if (size > RELAY_HEADER_BYTES + 32) {
      return;
    }

    uint8_t type;
    uint64_t sequence;
    uint64_t session_id;
    uint8_t session_version;
    if (relay::relay_peek_header(
         RELAY_DIRECTION_CLIENT_TO_SERVER, &type, &sequence, &session_id, &session_version, packet.data(), size) != RELAY_OK) {
      return;
    }

    uint64_t hash = session_id ^ session_version;
    relay::relay_session_t* session = (*(mRelay.sessions))[hash];
    if (!session) {
      return;
    }

    if (session->expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->client_to_server_sequence) {
      return;
    }

    session->client_to_server_sequence = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }

    mSocket.send(session->next_address, packet.data(), size);
  }

  void Communicator::handleSessionPongPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    mLogger.addToSessionPong(size);
    if (size > RELAY_HEADER_BYTES + 32) {
      return;
    }

    uint8_t type;
    uint64_t sequence;
    uint64_t session_id;
    uint8_t session_version;
    if (relay::relay_peek_header(
         RELAY_DIRECTION_SERVER_TO_CLIENT, &type, &sequence, &session_id, &session_version, packet.data(), size) != RELAY_OK) {
      return;
    }

    uint64_t hash = session_id ^ session_version;
    relay::relay_session_t* session = (*(mRelay.sessions))[hash];
    if (!session) {
      return;
    }

    if (session->expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->server_to_client_sequence) {
      return;
    }

    session->server_to_client_sequence = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }

    mSocket.send(session->prev_address, packet.data(), size);
  }

  void Communicator::handleNearPingPacket(
   std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from)
  {
    mLogger.addToNearPing(size);

    if (size != 1 + 8 + 8 + 8 + 8) {
      return;
    }

    packet[0] = RELAY_NEAR_PONG_PACKET;
    mSocket.send(from, packet.data(), size - 16); // TODO why 16?
  }
}  // namespace net