#ifndef NET_BUFFERED_SENDER_HPP
#define NET_BUFFERED_SENDER_HPP

#include "os/platform.hpp"

namespace net
{
  template <size_t MaxCapacity, size_t TimeoutInMicroseconds>
  class BufferedSender
  {
    static_assert(MaxCapacity > 0 && MaxCapacity <= 1024);  // 1024 is the hard limit for sendmmsg()

   public:
    BufferedSender(const os::Socket& socket);
    ~BufferedSender();

    void queue(const net::Address& addr, const uint8_t* data, size_t len);

    void autoSend();

   private:
    const os::Socket& mSocket;
    std::atomic<size_t> mNextIndex;

    std::array<mmsghdr, MaxCapacity> mHeaders;

    // running out of stack space, these need to be vectors
    std::vector<iovec> mIOVecBuff;
    std::vector<std::array<uint8_t, RELAY_MAX_PACKET_BYTES>> mDataBuff;
    std::vector<std::array<uint8_t, sizeof(sockaddr_in6)>> mAddrBuff;

    std::unique_ptr<std::thread> mSendThread;
    std::mutex mLock;

    std::atomic<bool> mShouldAutoSend;

    void sendAll();
  };

  template <size_t MaxCapacity, size_t TimeoutInMicroseconds>
  BufferedSender<MaxCapacity, TimeoutInMicroseconds>::BufferedSender(const os::Socket& socket)
   : mSocket(socket), mNextIndex(0), mShouldAutoSend(true)
  {
    mIOVecBuff.resize(MaxCapacity);
    mDataBuff.resize(MaxCapacity);
    mAddrBuff.resize(MaxCapacity);

    for (size_t i = 0; i < MaxCapacity; i++) {
      auto& header = mHeaders[i];
      auto& addr = mAddrBuff[i];
      auto& vec = mIOVecBuff[i];

      // iovec ptr assignment
      {
        vec.iov_base = mDataBuff[i].data();

        header.msg_hdr.msg_iovlen = 1;
        header.msg_hdr.msg_iov = &vec;
      }

      // address ptr assignment
      {
        header.msg_hdr.msg_name = addr.data();
      }
    }

    // for dev purposes, if max is 1 then it's the same as sending immediately, or if timeout is 0 always wait
    if (MaxCapacity > 1 && TimeoutInMicroseconds > 1) {
      mSendThread = std::make_unique<std::thread>([this] {
        while (mSocket.isOpen()) {
          std::this_thread::sleep_for(1us * TimeoutInMicroseconds);
          autoSend();
          mShouldAutoSend = true;
        }
      });
    }
  }

  template <size_t MaxCapacity, size_t TimeoutInMicroseconds>
  BufferedSender<MaxCapacity, TimeoutInMicroseconds>::~BufferedSender()
  {
    if (mSendThread) {
      mSendThread->join();
    }
  }

  template <size_t MaxCapacity, size_t TimeoutInMicroseconds>
  void BufferedSender<MaxCapacity, TimeoutInMicroseconds>::queue(const net::Address& addr, const uint8_t* data, size_t len)
  {
    assert(len <= RELAY_MAX_PACKET_BYTES);
    std::lock_guard<std::mutex> lk(mLock);
    auto& vec = mIOVecBuff[mNextIndex]; // header and iovec map to the same index

    vec.iov_len = len;
    std::copy(data, data + len, reinterpret_cast<uint8_t*>(vec.iov_base));

    addr.to(mHeaders[mNextIndex++]);  // be careful here, the increment must come last, but before the if

    if (mNextIndex == MaxCapacity) {
      LogDebug("reached max capacity");
      sendAll();
      mShouldAutoSend = false;
    }
  }

  template <size_t MaxCapacity, size_t TimeoutInMicroseconds>
  void BufferedSender<MaxCapacity, TimeoutInMicroseconds>::autoSend()
  {
    std::lock_guard<std::mutex> lk(mLock);
    if (mShouldAutoSend) {
      LogDebug("auto sending");
      sendAll();
    }
  }

  template <size_t MaxCapacity, size_t TimeoutInMicroseconds>
  void BufferedSender<MaxCapacity, TimeoutInMicroseconds>::sendAll()
  {
    if (mNextIndex > 0) {
      // mNextIndex also keeps track of how many messages are to be sent
      // after the call it contains the number of successfully sent messages
      if (!mSocket.multisend(mHeaders, mNextIndex)) {
        Log("failed to send buffered messages, sent ", mNextIndex, " messages");
      }

      mNextIndex = 0;
    }
  }
}  // namespace net
#endif