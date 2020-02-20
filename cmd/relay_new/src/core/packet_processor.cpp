#include "includes.h"
#include "packet_processor.hpp"

#include "encoding/read.hpp"

#include "relay/relay_platform.hpp"
#include "relay/relay.hpp"

namespace core
{
  PacketProcessor::PacketProcessor(const util::Clock& relayClock,
   const crypto::Keychain& keychain,
   const core::RouterInfo& routerInfo,
   core::SessionMap& sessions,
   core::RelayManager& relayManager,
   volatile bool& handle,
   util::ThroughputLogger* logger)
   : mRelayClock(relayClock),
     mKeychain(keychain),
     mRouterInfo(routerInfo),
     mSessionMap(sessions),
     mRelayManager(relayManager),
     mShouldProcess(handle),
     mLogger(logger)
  {}

  void PacketProcessor::listen(os::Socket& socket)
  {
    static std::atomic<int> listenCounter;
    int listenIndx = listenCounter.fetch_add(1);

    LogDebug("Listening for packets {", listenIndx, '}');

    std::array<uint8_t, RELAY_MAX_PACKET_BYTES> packetData;

    while (this->mShouldProcess) {
      net::Address from;
      const int packet_bytes = socket.recv(from, packetData.data(), sizeof(uint8_t) * packetData.size());

      if (packet_bytes == 0) {
        continue;
      }

      // LogDebug("Got packet on {", listenIndx, '}');

      if (packetData[0] == RELAY_PING_PACKET && packet_bytes == 9) {
        this->handleRelayPingPacket(socket, packetData, packet_bytes, from);
      } else if (packetData[0] == RELAY_PONG_PACKET && packet_bytes == 9) {
        this->handleRelayPongPacket(packetData, packet_bytes, from);
      } else if (packetData[0] == RELAY_ROUTE_REQUEST_PACKET) {
        this->handleRouteRequestPacket(socket, packetData, packet_bytes, from);
      } else if (packetData[0] == RELAY_ROUTE_RESPONSE_PACKET) {
        this->handleRouteResponsePacket(socket, packetData, packet_bytes, from);
      } else if (packetData[0] == RELAY_CONTINUE_REQUEST_PACKET) {
        this->handleContinueRequestPacket(socket, packetData, packet_bytes);
      } else if (packetData[0] == RELAY_CONTINUE_RESPONSE_PACKET) {
        this->handleContinueResponsePacket(socket, packetData, packet_bytes);
      } else if (packetData[0] == RELAY_CLIENT_TO_SERVER_PACKET) {
        this->handleClientToServerPacket(socket, packetData, packet_bytes);
      } else if (packetData[0] == RELAY_SERVER_TO_CLIENT_PACKET) {
        this->handleServerToClientPacket(socket, packetData, packet_bytes);
      } else if (packetData[0] == RELAY_SESSION_PING_PACKET) {
        this->handleSessionPingPacket(socket, packetData, packet_bytes);
      } else if (packetData[0] == RELAY_SESSION_PONG_PACKET) {
        this->handleSessionPongPacket(socket, packetData, packet_bytes);
      } else if (packetData[0] == RELAY_NEAR_PING_PACKET) {
        this->handleNearPingPacket(socket, packetData, packet_bytes, from);
      } else {
        LogDebug("Received unknown packet type: ", std::hex, (int)packetData[0]);
        if (mLogger != nullptr) {
          mLogger->addToUnknown(packet_bytes);
        }
      }
    }
  }

  void PacketProcessor::handleRelayPingPacket(
   os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, net::Address& from)
  {
    if (mLogger != nullptr) {
      mLogger->addToRelayPingPacket(size);
    }

    // mark the 0'th index as a pong and send it back from where it came
    packet[0] = RELAY_PONG_PACKET;
    if (!socket.send(from, packet.data(), 9)) {
      Log("Failed to send data");
    }
  }

  void PacketProcessor::handleRelayPongPacket(
   std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, net::Address& from)
  {
    if (mLogger != nullptr) {
      mLogger->addToRelayPongPacket(size);
    }

    // read the uint from the packet - this could be brought out of the mutex
    const uint8_t* p = packet.data() + 1;
    uint64_t sequence = encoding::read_uint64(&p);

    // process the pong time
    mRelayManager.processPong(from, sequence);
  }

  void PacketProcessor::handleRouteRequestPacket(
   os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, net::Address& from)
  {
    LogDebug("got route request from ", from);
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

    if (relay::relay_read_encrypted_route_token(
         &p, &token, mKeychain.RouterPublicKey.data(), mKeychain.RelayPrivateKey.data()) != RELAY_OK) {
      Log("ignoring route request. could not read route token");
      return;
    }

    // don't do anything if the token is expired - probably should log something here
    if (tokenIsExpired(token)) {
      Log("ignoring route request. token expired");
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

      // printf("session created: %" PRIx64 ".%d\n", token.session_id, token.session_version);
      std::stringstream ss;
      ss << std::hex << token.session_id << '.' << std::dec << static_cast<unsigned int>(token.session_version);
      Log("session created: ", ss.str());
    }

    // remove this part of the token by offseting it the request packet bytes
    packet[RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES] = RELAY_ROUTE_REQUEST_PACKET;
    socket.send(
     token.next_address, packet.data() + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES, size - RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES);
    LogDebug("sent token to ", token.next_address);
  }

  void PacketProcessor::handleRouteResponsePacket(
   os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, net::Address& from)
  {
    LogDebug("got route response from ", from);
    if (mLogger != nullptr) {
      mLogger->addToRouteResp(size);
    }

    if (size != RELAY_HEADER_BYTES) {
      Log("ignoring route response, header byte count invalid: ", size, " != ", RELAY_HEADER_BYTES);
      return;
    }

    uint8_t type;
    uint64_t sequence;
    uint64_t session_id;
    uint8_t session_version;
    if (relay::relay_peek_header(
         RELAY_DIRECTION_SERVER_TO_CLIENT, &type, &sequence, &session_id, &session_version, packet.data(), size) != RELAY_OK) {
      Log("ignoring route response, relay header could not be read");
      return;
    }

    uint64_t hash = session_id ^ session_version;

    mSessionMap.Lock.lock();
    auto iter = mSessionMap.find(hash);
    auto end = mSessionMap.end();
    mSessionMap.Lock.unlock();

    if (iter == end) {
      Log("ignoring route response, could not find session");
      return;
    }

    mSessionMap.Lock.lock();
    auto session = mSessionMap[hash];
    mSessionMap.Lock.unlock();

    if (sessionIsExpired(session)) {
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

    socket.send(session->PrevAddr, packet.data(), size);
    LogDebug("sent response to ", session->PrevAddr);
  }

  void PacketProcessor::handleContinueRequestPacket(
   os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
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
    if (relay_read_encrypted_continue_token(&p, &token, mKeychain.RouterPublicKey.data(), mKeychain.RelayPrivateKey.data()) !=
        RELAY_OK) {
      Log("ignoring continue request. could not read continue token");
      return;
    }

    if (tokenIsExpired(token)) {
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

    if (sessionIsExpired(session)) {
      return;
    }

    if (session->ExpireTimestamp != token.expire_timestamp) {
      printf("session continued: %" PRIx64 ".%d\n", token.session_id, token.session_version);
    }

    session->ExpireTimestamp = token.expire_timestamp;
    packet[RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES] = RELAY_CONTINUE_REQUEST_PACKET;

    socket.send(
     session->NextAddr, packet.data() + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES, size - RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES);
  }

  void PacketProcessor::handleContinueResponsePacket(
   os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
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

    if (sessionIsExpired(session)) {
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
    socket.send(session->PrevAddr, packet.data(), size);
  }

  void PacketProcessor::handleClientToServerPacket(
   os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    LogDebug("got client to server packet");
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

    if (sessionIsExpired(session)) {
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

    socket.send(session->NextAddr, packet.data(), size);
    LogDebug("sent client packet to ", session->NextAddr);
  }

  void PacketProcessor::handleServerToClientPacket(
   os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
  {
    LogDebug("got server to client packet");
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

    if (sessionIsExpired(session)) {
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

    socket.send(session->PrevAddr, packet.data(), size);
    LogDebug("sent server packet to ", session->PrevAddr);
  }

  void PacketProcessor::handleSessionPingPacket(
   os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
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

    if (sessionIsExpired(session)) {
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

    socket.send(session->NextAddr, packet.data(), size);
  }

  void PacketProcessor::handleSessionPongPacket(
   os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size)
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

    if (sessionIsExpired(session)) {
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

    socket.send(session->PrevAddr, packet.data(), size);
  }

  void PacketProcessor::handleNearPingPacket(
   os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, net::Address& from)
  {
    if (mLogger != nullptr) {
      mLogger->addToNearPing(size);
    }

    if (size != 1 + 8 + 8 + 8 + 8) {
      return;
    }

    packet[0] = RELAY_NEAR_PONG_PACKET;
    socket.send(from, packet.data(), size - 16);  // TODO why 16?
  }
}  // namespace core