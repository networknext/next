#ifndef UTIL_THROUGHPUT_LOGGER_HPP
#define UTIL_THROUGHPUT_LOGGER_HPP

#include "util/console.hpp"

using namespace std::chrono_literals;

namespace util
{
  struct ThroughputStats
  {
    ThroughputStats() = default;

    void add(size_t count);

    size_t PacketCount = 0;
    size_t ByteCount = 0;

    void reset();

    ThroughputStats operator+(const ThroughputStats& other);
  };

  class ThroughputLogger
  {
   public:
    ThroughputLogger(std::ostream& output);
    ~ThroughputLogger();

    void addToRelayPingPacket(size_t count);
    void addToRelayPongPacket(size_t count);
    void addToRouteReq(size_t count);
    void addToRouteResp(size_t count);
    void addToContReq(size_t count);
    void addToContResp(size_t count);
    void addToCliToServ(size_t count);
    void addToServToCli(size_t count);
    void addToSessionPing(size_t count);
    void addToSessionPong(size_t count);
    void addToNearPing(size_t count);
    void addToUnknown(size_t count);

    void stop();

   private:
    std::atomic<bool> mAlive;
    std::ostream& mOutput;
    util::Console mConsole;

    std::mutex mLock;
    std::unique_ptr<std::thread> mPrintThread;

    std::size_t mEmptyPacketsTotal = 0;

    ThroughputStats mRelayPing;
    ThroughputStats mRelayPong;

    ThroughputStats mRouteReq;
    ThroughputStats mRouteResp;

    ThroughputStats mContReq;
    ThroughputStats mContResp;

    ThroughputStats mCliToServ;
    ThroughputStats mServToCli;

    ThroughputStats mSessionPing;
    ThroughputStats mSessionPong;

    ThroughputStats mNearPing;

    ThroughputStats mUnknown;

    void reset();
  };

  inline void ThroughputStats::add(size_t count)
  {
    this->ByteCount += count;
    this->PacketCount++;
  }

  inline ThroughputStats ThroughputStats::operator+(const ThroughputStats& other)
  {
    ThroughputStats retval;
    retval.ByteCount = this->ByteCount + other.ByteCount;
    retval.PacketCount = this->PacketCount + other.PacketCount;
    return retval;
  }

  inline void ThroughputStats::reset()
  {
    ByteCount = 0;
    PacketCount = 0;
  }

  inline ThroughputLogger::ThroughputLogger(std::ostream& output): mAlive(true), mOutput(output), mConsole(mOutput)
  {
    mPrintThread = std::make_unique<std::thread>([this] {
      while (this->mAlive) {
        std::this_thread::sleep_for(1s);

        ThroughputStats relayPing;
        ThroughputStats relayPong;

        ThroughputStats routeReq;
        ThroughputStats routeResp;

        ThroughputStats contReq;
        ThroughputStats contResp;

        ThroughputStats cliToServ;
        ThroughputStats servToCli;

        ThroughputStats sessionPing;
        ThroughputStats sessionPong;

        ThroughputStats nearPing;
        ThroughputStats unknown;

        ThroughputStats total;

        {
          std::lock_guard<std::mutex> lk(mLock);
          relayPing = mRelayPing;
          relayPong = mRelayPong;

          routeReq = mRouteReq;
          routeResp = mRouteResp;

          contReq = mContReq;
          contResp = mContResp;

          cliToServ = mCliToServ;
          servToCli = mServToCli;

          sessionPing = mSessionPing;
          sessionPong = mSessionPong;

          nearPing = mNearPing;
          unknown = mUnknown;

          this->reset();
        }

        total = relayPing + relayPong + routeReq + routeResp + contReq + contResp + cliToServ + servToCli + sessionPing +
                sessionPong + nearPing;

        mConsole.write("\n------------------------------------------------\n\n");

        // Total
        mConsole.log("Total Bytes received: ", total.ByteCount, "/s");
        mConsole.log("Total Packets received: ", total.PacketCount, "/s\n");

        mConsole.log("Total Unknown Bytes received: ", unknown.ByteCount, "/s");
        mConsole.log("Total Unknown Packets received: ", unknown.PacketCount, "/s\n");

        // relay
        mConsole.log("Relay Ping Bytes received: ", relayPing.ByteCount, "/s");
        mConsole.log("Relay Ping Packets received: ", relayPing.PacketCount, "/s\n");

        mConsole.log("Relay Pong Bytes received: ", relayPong.ByteCount, "/s");
        mConsole.log("Relay Pong Packets received: ", relayPong.PacketCount, "/s\n");

        // route
        mConsole.log("Route Req Bytes received: ", routeReq.ByteCount, "/s");
        mConsole.log("Route Req Packets received: ", routeReq.PacketCount, "/s\n");

        mConsole.log("Route Resp Bytes received: ", routeResp.ByteCount, "/s");
        mConsole.log("Route Resp Packets received: ", routeResp.PacketCount, "/s\n");

        // cont
        mConsole.log("Cont Req Bytes received: ", contReq.ByteCount, "/s");
        mConsole.log("Cont Req Packets received: ", contReq.PacketCount, "/s\n");

        mConsole.log("Cont Resp Bytes received: ", contResp.ByteCount, "/s");
        mConsole.log("Cont Resp Packets received: ", contResp.PacketCount, "/s\n");

        // cli to serv | serv to cli
        mConsole.log("Cli To Serv Bytes received: ", cliToServ.ByteCount, "/s");
        mConsole.log("Cli To Serv Packets received: ", cliToServ.PacketCount, "/s\n");

        mConsole.log("Serv To Cli Bytes received: ", servToCli.ByteCount, "/s");
        mConsole.log("Serv To Cli Packets received: ", servToCli.PacketCount, "/s\n");

        // session
        mConsole.log("Session Ping Bytes received: ", sessionPing.ByteCount, "/s");
        mConsole.log("Session Ping Packets received: ", sessionPing.PacketCount, "/s\n");

        mConsole.log("Session Pong Bytes received: ", sessionPong.ByteCount, "/s");
        mConsole.log("Session Pong Packets received: ", sessionPong.PacketCount, "/s\n");

        // near
        mConsole.log("Near Ping Bytes received: ", nearPing.ByteCount, "/s");
        mConsole.log("Near Ping Packets received: ", nearPing.PacketCount, "/s\n");

        mConsole.flush();
      }
    });
  }

  inline ThroughputLogger::~ThroughputLogger()
  {
    stop();
  }

  inline void ThroughputLogger::addToRelayPingPacket(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mRelayPing.add(count);
  }

  inline void ThroughputLogger::addToRelayPongPacket(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mRelayPong.add(count);
  }

  inline void ThroughputLogger::addToRouteReq(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mRouteReq.add(count);
  }

  inline void ThroughputLogger::addToRouteResp(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mRouteResp.add(count);
  }

  inline void ThroughputLogger::addToContReq(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mContReq.add(count);
  }

  inline void ThroughputLogger::addToContResp(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mContResp.add(count);
  }

  inline void ThroughputLogger::addToCliToServ(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mCliToServ.add(count);
  }

  inline void ThroughputLogger::addToServToCli(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mServToCli.add(count);
  }

  inline void ThroughputLogger::addToSessionPing(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mSessionPing.add(count);
  }

  inline void ThroughputLogger::addToSessionPong(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mSessionPong.add(count);
  }

  inline void ThroughputLogger::addToNearPing(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mNearPing.add(count);
  }

  inline void ThroughputLogger::addToUnknown(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mUnknown.add(count);
  }

  inline void ThroughputLogger::stop()
  {
    mAlive = false;
    if (mPrintThread && mPrintThread->joinable()) {
      mPrintThread->join();
      mPrintThread = nullptr;
    }
  }

  inline void ThroughputLogger::reset()
  {
    mEmptyPacketsTotal = 0;
    mRelayPing.reset();
    mRelayPong.reset();
    mRouteReq.reset();
    mRouteResp.reset();
    mContReq.reset();
    mContResp.reset();
    mCliToServ.reset();
    mServToCli.reset();
    mSessionPing.reset();
    mSessionPong.reset();
    mNearPing.reset();
    mUnknown.reset();
  }
}  // namespace util
#endif
