#ifndef NET_BUFFERED_SENDER_HPP
#define NET_BUFFERED_SENDER_HPP

#include "net/message.hpp"
#include "os/platform.hpp"

namespace net
{
  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  class BufferedSender
  {
    static_assert(MaxCapacity > 0 && MaxCapacity <= 1024);  // 1024 is the hard limit for sendmmsg()

   public:
    BufferedSender(const os::Socket& socket);
    ~BufferedSender();

    /* Calls swap() on the msg, such that the passed in value should not be used after this */
    void queue(net::Message& msg);

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

    // for dev purposes, if max is 1 then it's the same as sending immediately
    if (MaxCapacity > 1) {
      mSendThread = std::make_unique<std::thread>([this] {
        while (mSocket.isOpen()) {
          std::this_thread::sleep_for(1ms * TimeoutInMilliseconds);
          autoSend();
          mShouldAutoSend = true;
        }
      });
    }
  }

  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  BufferedSender<MaxCapacity, TimeoutInMilliseconds>::~BufferedSender()
  {
    if (mSendThread) {
      mSendThread->join();
    }
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
    if (mNextIndex > 0) {
      int messagesSent = 0;

      // mNextIndex also keeps track of how many messages are to be sent
      if (!mSocket.multisend(mBuffer, mNextIndex, messagesSent)) {
        Log("failed to send buffered messages, sent ", messagesSent, " messages");
      }

      mNextIndex = 0;
    }
  }
}  // namespace net
#endif