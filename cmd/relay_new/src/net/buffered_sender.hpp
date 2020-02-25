#ifndef NET_BUFFERED_SENDER_HPP
#define NET_BUFFERED_SENDER_HPP

#include "net/message.hpp"
#include "os/platform.hpp"

namespace net
{
  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  class BufferedSender
  {
    static_assert(MaxCapacity > 0 && MaxCapacity <= 1024);

   public:
    BufferedSender(const os::Socket& socket);
    ~BufferedSender();

    /* Calls swap() on the msg, such that the passed in value should not be used after this */
    void queue(net::Message& msg);
    void queue(const net::Address& addr, std::vector<uint8_t>& data);
    template <size_t DataBuffSize>
    void queue(const net::Address& addr, std::array<uint8_t, DataBuffSize>& data, size_t len);
    void queue(const net::Address& addr, uint8_t* data, size_t len);

    void autoSend();

   private:
    const os::Socket& mSocket;
    std::atomic<size_t> mNextIndex;
    std::vector<net::Message> mBuffer;

    std::unique_ptr<std::thread> mSendThread;
    std::mutex mLock;

    std::atomic<bool> mShouldAutoSend;

    void sendAll();
  };

  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  BufferedSender<MaxCapacity, TimeoutInMilliseconds>::BufferedSender(const os::Socket& socket)
   : mSocket(socket), mNextIndex(0), mShouldAutoSend(true)
  {
    mBuffer.resize(MaxCapacity);
    mSendThread = std::make_unique<std::thread>([this] {
      while (mSocket.isOpen()) {
        std::this_thread::sleep_for(1ms * TimeoutInMilliseconds);
        autoSend();
        mShouldAutoSend = true;
      }
    });
  }

  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  BufferedSender<MaxCapacity, TimeoutInMilliseconds>::~BufferedSender()
  {
    mSendThread->join();
  }

  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  void BufferedSender<MaxCapacity, TimeoutInMilliseconds>::queue(net::Message& msg)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mBuffer[mNextIndex++].swap(msg);

    if (mNextIndex == MaxCapacity) {
      LogDebug("reached max capacity");
      sendAll();
      mShouldAutoSend = false;
    }
  }

  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  inline void BufferedSender<MaxCapacity, TimeoutInMilliseconds>::queue(const net::Address& addr, std::vector<uint8_t>& data)
  {
    net::Message msg(addr, data);
    queue(msg);
  }

  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  template <size_t DataBuffSize>
  inline void BufferedSender<MaxCapacity, TimeoutInMilliseconds>::queue(
   const net::Address& addr, std::array<uint8_t, DataBuffSize>& data, size_t len)
  {
    net::Message msg(addr, data, len);
    queue(msg);
  }

  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  inline void BufferedSender<MaxCapacity, TimeoutInMilliseconds>::queue(const net::Address& addr, uint8_t* data, size_t len)
  {
    net::Message msg(addr, data, len);
    queue(msg);
  }

  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  void BufferedSender<MaxCapacity, TimeoutInMilliseconds>::autoSend()
  {
    std::lock_guard<std::mutex> lk(mLock);
    if (mShouldAutoSend) {
      LogDebug("auto sending");
      sendAll();
    }
  }

  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  void BufferedSender<MaxCapacity, TimeoutInMilliseconds>::sendAll()
  {
    int messagesSent = 0;
    if (!mSocket.multisend(mBuffer, messagesSent)) {
      Log("failed to send buffered messages");
    }

    mNextIndex = 0;
  }
}  // namespace net
#endif