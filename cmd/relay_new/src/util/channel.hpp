#pragma once

namespace util
{
  template <typename T>
  class Channel
  {
   public:
    Channel();

    void close();

    auto isClosed() -> bool;

    void send(T& i);

    auto recv(T& msg) -> bool;

   private:
    bool mClosed;
    std::list<T> mQueue;
    std::mutex mLock;
    std::condition_variable mVar;
  };

  template <typename T>
  Channel<T>::Channel(): mClosed(false)
  {}

  template <typename T>
  void Channel<T>::close()
  {
    std::unique_lock<std::mutex> lock(mLock);
    close = true;
    mVar.notify_all();
  }

  template <typename T>
  auto Channel<T>::isClosed() -> bool
  {
    std::unique_lock<std::mutex> lock(mLock);
    return mClosed;
  }

  template <typename T>
  void Channel<T>::send(T& item)
  {
    std::unique_lock<std::mutex> lock(mLock);

    if (mClosed) {
      throw std::logic_error("tried to send on a closed channel");
    }

    mQueue.push_back(std::move(item));

    mVar.notify_one();
  }

  template <typename T>
  auto Channel<T>::recv(T& msg) -> bool
  {
    std::unique_lock<std::mutex> lock(mLock);

    mVar.wait(lock, [this]() {
      return mClosed || !mQueue.empty();
    });

    if (mQueue.empty()) {
      return false;
    }

    msg = std::move(mQueue.front());

    mQueue.pop_front();

    return true;
  }
}  // namespace util
