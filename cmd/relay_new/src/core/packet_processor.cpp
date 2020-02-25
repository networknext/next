#include "includes.h"
#include "packet_processor.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

#include "relay/relay_platform.hpp"
#include "relay/relay.hpp"

#include "core/route_token.hpp"
#include "core/continue_token.hpp"

namespace core
{
  PacketProcessor::PacketProcessor(os::Socket& socket,
   const util::Clock& relayClock,
   const crypto::Keychain& keychain,
   const core::RouterInfo& routerInfo,
   core::SessionMap& sessions,
   core::RelayManager& relayManager,
   volatile bool& handle,
   util::ThroughputLogger* logger)
   : mSocket(socket),
     mRelayClock(relayClock),
     mKeychain(keychain),
     mRouterInfo(routerInfo),
     mSessionMap(sessions),
     mRelayManager(relayManager),
     mShouldProcess(handle),
     mLogger(logger),
     mSender(socket)
  {}

  void PacketProcessor::process(std::condition_variable& var, std::atomic<bool>& readyToReceive)
  {
    static std::atomic<int> listenCounter;
    int listenIndx = listenCounter.fetch_add(1);
    (void)listenIndx;

    GenericPacket packetData;

    LogDebug("listening for packets {", listenIndx, '}');

    readyToReceive = true;
    var.notify_one();

    while (this->mShouldProcess) {
      net::Address from;
      const int packet_bytes = mSocket.recv(from, packetData.data(), sizeof(uint8_t) * packetData.size());

      // timeout
      if (packet_bytes == 0) {
        continue;
      }

      LogDebug("got packet on {", listenIndx, "} / type: ", static_cast<unsigned int>(packetData[0]));

      if (packetData[0] == RELAY_PING_PACKET && packet_bytes == RELAY_PING_PACKET_BYTES) {
        this->handleRelayPingPacket(packetData, packet_bytes);
      } else if (packetData[0] == RELAY_PONG_PACKET && packet_bytes == RELAY_PING_PACKET_BYTES) {
        this->handleRelayPongPacket(packetData, packet_bytes);
      } else if (packetData[0] == RELAY_ROUTE_REQUEST_PACKET) {
        this->handleRouteRequestPacket(packetData, packet_bytes, from);
      } else if (packetData[0] == RELAY_ROUTE_RESPONSE_PACKET) {
        this->handleRouteResponsePacket(packetData, packet_bytes, from);
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
        LogDebug("received unknown packet type: ", std::hex, (int)packetData[0]);
        if (mLogger != nullptr) {
          mLogger->addToUnknown(packet_bytes);
        }
      }
    }
  }

  void PacketProcessor::handleRelayPingPacket(GenericPacket& packet, const int size)
  {
    if (mLogger != nullptr) {
      mLogger->addToRelayPingPacket(size);
    }

    net::Address addr;  // where it actually came from
    (void)addr;

    // mark the 0'th index as a pong and send it back from where it came
    packet[0] = RELAY_PONG_PACKET;  // set the identifier byte as pong
    size_t index = 1;               // skip the identifier byte
    uint64_t sequence = encoding::ReadUint64(packet, index);
    (void)sequence;
    size_t addrIndx = index;
    encoding::ReadAddress(packet, index, addr);  // pings are sent on a different port, need to read actual address
    LogDebug("got ping packet from ", addr);

    encoding::WriteAddress(packet, addrIndx, mSocket.getAddress());

    if (!mSocket.send(addr, packet.data(), RELAY_PING_PACKET_BYTES)) {
      Log("failed to send data");
    }
  }

  void PacketProcessor::handleRelayPongPacket(GenericPacket& packet, const int size)
  {
    if (mLogger != nullptr) {
      mLogger->addToRelayPongPacket(size);
    }

    net::Address addr;  // the actual from

    size_t index = 1;  // skip the identifier byte
    uint64_t sequence = encoding::ReadUint64(packet, index);
    // pings are sent on a different port, need to read actual address to stay consistent
    encoding::ReadAddress(packet, index, addr);
    LogDebug("got pong packet from ", addr);

    // process the pong time
    mRelayManager.processPong(addr, sequence);
  }

  void PacketProcessor::handleRouteRequestPacket(GenericPacket& packet, const int size, net::Address& from)
  {
    LogDebug("got route request from ", from);
    if (mLogger != nullptr) {
      mLogger->addToRouteReq(size);
    }

    if (size < int(1 + RouteToken::EncryptedByteSize * 2)) {
      Log("ignoring route request. bad packet size (", size, ")");
      return;
    }

    // ignore the header byte of the packet
    size_t index = 1;
    core::RouteToken token;

    if (!token.readEncrypted(packet, index, mKeychain.RouterPublicKey, mKeychain.RelayPrivateKey)) {
      Log("ignoring route request. could not read route token");
      return;
    }

    // don't do anything if the token is expired - probably should log something here
    if (tokenIsExpired(token)) {
      Log("ignoring route request. token expired");
      return;
    }

    // create a new session and add it to the session map
    uint64_t hash = token.key();

    core::SessionMap::iterator iter, end;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      iter = mSessionMap.find(hash);
      end = mSessionMap.end();
    }

    if (iter == end) {
      // create the session
      auto session = std::make_shared<Session>();
      assert(session);

      // fill it with data in the token
      session->ExpireTimestamp = token.ExpireTimestamp;
      session->SessionID = token.SessionID;
      session->SessionVersion = token.SessionVersion;
      session->ClientToServerSeq = 0;
      session->ServerToClientSeq = 0;
      session->KbpsUp = token.KbpsUp;
      session->KbpsDown = token.KbpsDown;
      session->PrevAddr = from;
      session->NextAddr = token.NextAddr;

      // store it
      std::copy(token.PrivateKey.begin(), token.PrivateKey.end(), session->PrivateKey.begin());
      relay_replay_protection_reset(&session->ClientToServerProtection);
      relay_replay_protection_reset(&session->ServerToClientProtection);

      {
        std::lock_guard<std::mutex> lk(mSessionMap.Lock);
        mSessionMap[hash] = session;
      }

      std::stringstream ss;
      ss << std::hex << token.SessionID << '.' << std::dec << static_cast<unsigned int>(token.SessionVersion);
      Log("session created: ", ss.str());
    }  // TODO else what?

    // remove this part of the token by offseting it the request packet bytes
    packet[RouteToken::EncryptedByteSize] = RELAY_ROUTE_REQUEST_PACKET;
    mSocket.send(token.NextAddr, packet.data() + RouteToken::EncryptedByteSize, size - RouteToken::EncryptedByteSize);
    LogDebug("sent route request to ", token.NextAddr);
  }

  void PacketProcessor::handleRouteResponsePacket(GenericPacket& packet, const int size, net::Address& from)
  {
    (void)from;
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

    core::SessionMap::iterator iter, end;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      iter = mSessionMap.find(hash);
      end = mSessionMap.end();
    }

    if (iter == end) {
      Log("ignoring route response, could not find session");
      return;
    }

    core::SessionPtr session;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      session = mSessionMap[hash];
    }

    if (sessionIsExpired(session)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->ServerToClientSeq) {
      return;
    }

    session->ServerToClientSeq = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), packet.data(), size) !=
        RELAY_OK) {
      return;
    }

    mSocket.send(session->PrevAddr, packet.data(), size);
    LogDebug("sent response to ", session->PrevAddr);
  }

  void PacketProcessor::handleContinueRequestPacket(GenericPacket& packet, const int size)
  {
    if (mLogger != nullptr) {
      mLogger->addToContReq(size);
    }

    if (size < int(1 + ContinueToken::EncryptedByteSize * 2)) {
      Log("ignoring continue request. bad packet size (", size, ")");
      return;
    }

    size_t index = 1;
    core::ContinueToken token;
    if (!token.readEncrypted(packet, index, mKeychain.RouterPublicKey, mKeychain.RelayPrivateKey)) {
      Log("ignoring continue request. could not read continue token");
      return;
    }

    if (tokenIsExpired(token)) {
      return;
    }

    uint64_t hash = token.key();

    core::SessionMap::iterator iter, end;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      iter = mSessionMap.find(hash);
      end = mSessionMap.end();
    }

    if (iter == end) {
      return;
    }

    core::SessionPtr session;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      session = mSessionMap[hash];
    }

    if (sessionIsExpired(session)) {
      return;
    }

    if (session->ExpireTimestamp != token.ExpireTimestamp) {
      std::stringstream ss;
      ss << std::hex << token.SessionID << '.' << std::dec << static_cast<unsigned int>(token.SessionVersion);
      Log("session continued: ", ss.str());
    }

    session->ExpireTimestamp = token.ExpireTimestamp;
    packet[ContinueToken::EncryptedByteSize] = RELAY_CONTINUE_REQUEST_PACKET;

    mSocket.send(session->NextAddr, packet.data() + ContinueToken::EncryptedByteSize, size - ContinueToken::EncryptedByteSize);
  }

  void PacketProcessor::handleContinueResponsePacket(GenericPacket& packet, const int size)
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

    core::SessionMap::iterator iter, end;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      iter = mSessionMap.find(hash);
      end = mSessionMap.end();
    }

    if (iter == end) {
      return;
    }

    core::SessionPtr session;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      session = mSessionMap[hash];
    }

    if (sessionIsExpired(session)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);

    if (clean_sequence <= session->ServerToClientSeq) {
      return;
    }

    session->ServerToClientSeq = clean_sequence;

    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), packet.data(), size) !=
        RELAY_OK) {
      return;
    }

    mSocket.send(session->PrevAddr, packet.data(), size);
  }

  void PacketProcessor::handleClientToServerPacket(GenericPacket& packet, const int size)
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

    core::SessionMap::iterator iter, end;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      iter = mSessionMap.find(hash);
      end = mSessionMap.end();
    }

    if (iter == end) {
      return;
    }

    core::SessionPtr session;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      session = mSessionMap[hash];
    }

    if (sessionIsExpired(session)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (relay_replay_protection_already_received(&session->ClientToServerProtection, clean_sequence)) {
      return;
    }

    relay_replay_protection_advance_sequence(&session->ClientToServerProtection, clean_sequence);
    if (relay::relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, session->PrivateKey.data(), packet.data(), size) !=
        RELAY_OK) {
      return;
    }

    mSocket.send(session->NextAddr, packet.data(), size);
    LogDebug("sent client packet to ", session->NextAddr);
  }

  void PacketProcessor::handleServerToClientPacket(GenericPacket& packet, const int size)
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

    core::SessionMap::iterator iter, end;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      iter = mSessionMap.find(hash);
      end = mSessionMap.end();
    }

    if (iter == end) {
      return;
    }

    core::SessionPtr session;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      session = mSessionMap[hash];
    }

    if (sessionIsExpired(session)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (relay_replay_protection_already_received(&session->ServerToClientProtection, clean_sequence)) {
      return;
    }

    relay_replay_protection_advance_sequence(&session->ServerToClientProtection, clean_sequence);
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), packet.data(), size) !=
        RELAY_OK) {
      return;
    }

    mSocket.send(session->PrevAddr, packet.data(), size);
    LogDebug("sent server packet to ", session->PrevAddr);
  }

  void PacketProcessor::handleSessionPingPacket(GenericPacket& packet, const int size)
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

    core::SessionMap::iterator iter, end;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      iter = mSessionMap.find(hash);
      end = mSessionMap.end();
    }

    if (iter == end) {
      return;
    }

    core::SessionPtr session;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      session = mSessionMap[hash];
    }

    if (sessionIsExpired(session)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->ClientToServerSeq) {
      return;
    }

    session->ClientToServerSeq = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, session->PrivateKey.data(), packet.data(), size) !=
        RELAY_OK) {
      return;
    }

    mSocket.send(session->NextAddr, packet.data(), size);
  }

  void PacketProcessor::handleSessionPongPacket(GenericPacket& packet, const int size)
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

    core::SessionMap::iterator iter, end;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      iter = mSessionMap.find(hash);
      end = mSessionMap.end();
    }

    if (iter == end) {
      return;
    }

    core::SessionPtr session;
    {
      std::lock_guard<std::mutex> lk(mSessionMap.Lock);
      session = mSessionMap[hash];
    }

    if (sessionIsExpired(session)) {
      return;
    }

    uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
    if (clean_sequence <= session->ServerToClientSeq) {
      return;
    }

    session->ServerToClientSeq = clean_sequence;
    if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), packet.data(), size) !=
        RELAY_OK) {
      return;
    }

    mSocket.send(session->PrevAddr, packet.data(), size);
  }

  void PacketProcessor::handleNearPingPacket(GenericPacket& packet, const int size, net::Address& from)
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