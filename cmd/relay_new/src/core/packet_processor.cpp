#include "includes.h"
#include "packet_processor.hpp"

#include "encoding/read.hpp"

#include "relay/relay_continue_token.hpp"
#include "relay/relay_platform.hpp"
#include "relay/relay_route_token.hpp"

namespace core
{
  PacketProcessor::PacketProcessor(
   os::Socket& socket, relay::relay_t& relay, volatile bool& handle, util::ThroughputLogger* logger)
   : mSocket(socket), mRelay(relay), mHandle(handle), mLogger(logger)
  {}

  PacketProcessor::~PacketProcessor()
  {
    stop();
  }

  void PacketProcessor::listen()
  {
    std::array<uint8_t, RELAY_MAX_PACKET_BYTES> packetData;

    while (this->mHandle) {
      legacy::relay_address_t from;
      const int packet_bytes = mSocket.recv(from, packetData.data(), sizeof(uint8_t) * packetData.size());

      if (packet_bytes == 0) {
        if (mLogger != nullptr) {
          mLogger->addToPkt0();
        }
        continue;
      }

      if (packetData[0] == RELAY_PING_PACKET && packet_bytes == 9) {
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
        if (mLogger != nullptr) {
          mLogger->addToUnknown(packet_bytes);
        }
      }
    }
  }

  void PacketProcessor::stop()
  {
    mHandle = false;
  }

  void PacketProcessor::handleRelayPingPacket(
   std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from)
  {
    if (mLogger != nullptr) {
      mLogger->addToRelayPingPacket(size);
    }

    // mark the 0'th index as a pong and send it back from where it came
    packet[0] = RELAY_PONG_PACKET;
    if (!mSocket.send(from, packet.data(), 9)) {
      Log("Failed to send data");
    }
  }

  void PacketProcessor::handleRelayPongPacket(
   std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from)
  {
    if (mLogger != nullptr) {
      mLogger->addToRelayPongPacket(size);
    }

    relay_platform_mutex_acquire(mRelay.mutex);

    // read the uint from the packet - this could be brought out of the mutex
    const uint8_t* p = packet.data() + 1;
    uint64_t sequence = encoding::read_uint64(&p);

    // process the pong time
    relay_manager_process_pong(mRelay.relay_manager, &from, sequence);
    relay_platform_mutex_release(mRelay.mutex);
  }

  void PacketProcessor::handleRouteRequestPacket(
   std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from)
  {
    if (mLogger != nullptr) {
      mLogger->addToRouteReq(size);
    }

    if (size < int(1 + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES * 2)) {
      Log("ignoring route request. bad packet size (", size, ")");
      return;
    }

    // ignore the header byte of the packet
    uint8_t* p = &packet[1];
    relay::relay_route_token_t token;

    if (relay::relay_read_encrypted_route_token(&p, &token, mRelay.router_public_key, mRelay.relay_private_key) != RELAY_OK) {
      Log("ignoring route request. could not read route token");
      return;
    }

    // don't do anything if the token is expired - probably should log something here
    if (token.expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    // create a new session and add it to the session map
    uint64_t hash = token.session_id ^ token.session_version;
    mSessionMap.Lock.lock();
    auto iter = mSessionMap.find(hash);
    auto end = mSessionMap.end();
    mSessionMap.Lock.unlock();
    if (iter == end) {
      // create the session
      auto session = std::make_shared<Session>();
      assert(session);

      // fill it with data in the token
      session->ExpireTimestamp = token.expire_timestamp;
      session->SessionID = token.session_id;
      session->SessionVersion = token.session_version;
      session->ClientToServerSeq = 0;
      session->ServerToClientSeq = 0;
      session->KbpsUp = token.kbps_up;
      session->KbpsDown = token.kbps_down;
      session->PrevAddr = from;
      session->NextAddr = token.next_address;

      // store it
      memcpy(session->private_key, token.private_key, crypto_box_SECRETKEYBYTES);
      relay_replay_protection_reset(&session->ClientToServerProtection);
      relay_replay_protection_reset(&session->ServerToClientProtection);
      mSessionMap.Lock.lock();
      mSessionMap[hash] = session;
      mSessionMap.Lock.unlock();

      printf("session created: %" PRIx64 ".%d\n", token.session_id, token.session_version);
    }

    // remove this part of the token by offseting it the request packet bytes
    packet[RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES] = RELAY_ROUTE_REQUEST_PACKET;
    mSocket.send(
     token.next_address, packet.data() + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES, size - RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES);
  }

  void PacketProcessor::handleRouteResponsePacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    if (mLogger != nullptr) {
      mLogger->addToRouteResp(size);
    }

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

    mSessionMap.Lock.lock();
    auto iter = mSessionMap.find(hash);
    auto end = mSessionMap.end();
    mSessionMap.Lock.unlock();
    if (iter == end) {
      return;
    }

    mSessionMap.Lock.lock();
    auto session = mSessionMap[hash];
    mSessionMap.Lock.unlock();

    if (session->ExpireTimestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->ServerToClientSeq) {
      return;
    }

    session->ServerToClientSeq = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }

    mSocket.send(session->PrevAddr, packet.data(), size);
  }

  void PacketProcessor::handleContinueRequestPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    if (mLogger != nullptr) {
      mLogger->addToContReq(size);
    }

    if (size < int(1 + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES * 2)) {
      Log("ignoring continue request. bad packet size (", size, ")");
      return;
    }
    uint8_t* p = &packet[1];
    relay::relay_continue_token_t token;
    if (relay_read_encrypted_continue_token(&p, &token, mRelay.router_public_key, mRelay.relay_private_key) != RELAY_OK) {
      Log("ignoring continue request. could not read continue token");
      return;
    }
    if (token.expire_timestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }
    uint64_t hash = token.session_id ^ token.session_version;
    mSessionMap.Lock.lock();
    auto iter = mSessionMap.find(hash);
    auto end = mSessionMap.end();
    mSessionMap.Lock.unlock();
    if (iter == end) {
      return;
    }

    mSessionMap.Lock.lock();
    auto session = mSessionMap[hash];
    mSessionMap.Lock.unlock();

    if (session->ExpireTimestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }
    if (session->ExpireTimestamp != token.expire_timestamp) {
      printf("session continued: %" PRIx64 ".%d\n", token.session_id, token.session_version);
    }
    session->ExpireTimestamp = token.expire_timestamp;
    packet[RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES] = RELAY_CONTINUE_REQUEST_PACKET;
    mSocket.send(
     session->NextAddr, packet.data() + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES, size - RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES);
  }

  void PacketProcessor::handleContinueResponsePacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    if (mLogger != nullptr) {
      mLogger->addToContResp(size);
    }

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
    mSessionMap.Lock.lock();
    auto iter = mSessionMap.find(hash);
    auto end = mSessionMap.end();
    mSessionMap.Lock.unlock();
    if (iter == end) {
      return;
    }

    mSessionMap.Lock.lock();
    auto session = mSessionMap[hash];
    mSessionMap.Lock.unlock();

    if (session->ExpireTimestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }
    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->ServerToClientSeq) {
      return;
    }
    session->ServerToClientSeq = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }
    mSocket.send(session->PrevAddr, packet.data(), size);
  }

  void PacketProcessor::handleClientToServerPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    if (mLogger != nullptr) {
      mLogger->addToCliToServ(size);
    }

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
    mSessionMap.Lock.lock();
    auto iter = mSessionMap.find(hash);
    auto end = mSessionMap.end();
    mSessionMap.Lock.unlock();
    if (iter == end) {
      return;
    }

    mSessionMap.Lock.lock();
    auto session = mSessionMap[hash];
    mSessionMap.Lock.unlock();

    if (session->ExpireTimestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (relay_replay_protection_already_received(&session->ClientToServerProtection, clean_sequence)) {
      return;
    }

    relay_replay_protection_advance_sequence(&session->ClientToServerProtection, clean_sequence);
    if (relay::relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }

    mSocket.send(session->NextAddr, packet.data(), size);
  }

  void PacketProcessor::handleServerToClientPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    if (mLogger != nullptr) {
      mLogger->addToServToCli(size);
    }

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
    mSessionMap.Lock.lock();
    auto iter = mSessionMap.find(hash);
    auto end = mSessionMap.end();
    mSessionMap.Lock.unlock();
    if (iter == end) {
      return;
    }

    mSessionMap.Lock.lock();
    auto session = mSessionMap[hash];
    mSessionMap.Lock.unlock();

    if (session->ExpireTimestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (relay_replay_protection_already_received(&session->ServerToClientProtection, clean_sequence)) {
      return;
    }

    relay_replay_protection_advance_sequence(&session->ServerToClientProtection, clean_sequence);
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }

    mSocket.send(session->PrevAddr, packet.data(), size);
  }

  void PacketProcessor::handleSessionPingPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    if (mLogger != nullptr) {
      mLogger->addToSessionPing(size);
    }

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
    mSessionMap.Lock.lock();
    auto iter = mSessionMap.find(hash);
    auto end = mSessionMap.end();
    mSessionMap.Lock.unlock();
    if (iter == end) {
      return;
    }

    mSessionMap.Lock.lock();
    auto session = mSessionMap[hash];
    mSessionMap.Lock.unlock();

    if (session->ExpireTimestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->ClientToServerSeq) {
      return;
    }

    session->ClientToServerSeq = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }

    mSocket.send(session->NextAddr, packet.data(), size);
  }

  void PacketProcessor::handleSessionPongPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    if (mLogger != nullptr) {
      mLogger->addToSessionPong(size);
    }

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
    mSessionMap.Lock.lock();
    auto iter = mSessionMap.find(hash);
    auto end = mSessionMap.end();
    mSessionMap.Lock.unlock();
    if (iter == end) {
      return;
    }

    mSessionMap.Lock.lock();
    auto session = mSessionMap[hash];
    mSessionMap.Lock.unlock();

    if (session->ExpireTimestamp < relay::relay_timestamp(&mRelay)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->ServerToClientSeq) {
      return;
    }

    session->ServerToClientSeq = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet.data(), size) != RELAY_OK) {
      return;
    }

    mSocket.send(session->PrevAddr, packet.data(), size);
  }

  void PacketProcessor::handleNearPingPacket(
   std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from)
  {
    if (mLogger != nullptr) {
      mLogger->addToNearPing(size);
    }

    if (size != 1 + 8 + 8 + 8 + 8) {
      return;
    }

    packet[0] = RELAY_NEAR_PONG_PACKET;
    mSocket.send(from, packet.data(), size - 16);  // TODO why 16?
  }
}  // namespace core