#ifndef NET_BUFFERED_SENDER_HPP
#define NET_BUFFERED_SENDER_HPP

#include "net/message.hpp"
#include "os/platform.hpp"

namespace net
{
  template <size_t MaxCapacity, size_t TimeoutInMilliseconds>
  class BufferedSender
  {
   public:
    BufferedSender(const os::Socket& socket);
    ~BufferedSender();

    /* Calls swap() on the msg, such that the passed in value should not be used after this */
    void queue(net::Message& msg);

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

        if (mShouldAutoSend) {
          std::lock_guard<std::mutex> lk(mLock);
          LogDebug("auto sending");
          sendAll();
        }

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