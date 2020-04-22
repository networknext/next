#pragma once

namespace util
{
  template <typename T>
  class Channel
  {
   public:
    virtual ~Channel();

   protected:
    Channel(
     std::shared_ptr<std::atomic<bool>> closed,
     std::shared_ptr<std::list<T>> queue,
     std::shared_ptr<std::mutex> lock,
     std::shared_ptr<std::condition_variable> var);

    std::shared_ptr<std::atomic<bool>> mClosed;
    std::shared_ptr<std::list<T>> mQueue;
    std::shared_ptr<std::mutex> mLock;
    std::shared_ptr<std::condition_variable> mVar;
  };

  template <typename T>
  class Sender: public Channel<T>
  {
   public:
    Sender(
     std::shared_ptr<std::atomic<bool>> closed,
     std::shared_ptr<std::list<T>> queue,
     std::shared_ptr<std::mutex> lock,
     std::shared_ptr<std::condition_variable> var);
    ~Sender() override = default;

    void send(T& msg);
  };

  template <typename T>
  class Receiver: public Channel<T>
  {
   public:
    Receiver(
     std::shared_ptr<std::atomic<bool>> closed,
     std::shared_ptr<std::list<T>> queue,
     std::shared_ptr<std::mutex> lock,
     std::shared_ptr<std::condition_variable> var);
    ~Receiver() override = default;

    auto recv(T& msg) -> bool;

    auto hasItems() -> bool;

   private:
    std::mutex mWaitLock;
  };

  template <typename T>
  inline std::tuple<Sender<T>, Receiver<T>> makeChannel()
  {
    auto closed = std::make_shared<std::atomic<bool>>();
    auto lock = std::make_shared<std::mutex>();
    auto queue = std::make_shared<std::list<T>>();
    auto var = std::make_shared<std::condition_variable>();

    Sender<T> sender(closed, lock, queue, var);
    Receiver<T> receiver(closed, lock, queue, var);

    return {sender, receiver};
  }

  template <typename T>
  Channel<T>::Channel(
   std::shared_ptr<std::atomic<bool>> closed,
   std::shared_ptr<std::list<T>> queue,
   std::shared_ptr<std::mutex> lock,
   std::shared_ptr<std::condition_variable> var)
   : mClosed(closed), mQueue(queue), mLock(lock), mVar(var)
  {}

  template <typename T>
  Channel<T>::~Channel()
  {
    std::unique_lock<std::mutex> lock(*mLock);
    close = true;
    mVar.notify_all();
  }

  template <typename T>
  Sender<T>::Sender(
   std::shared_ptr<std::atomic<bool>> closed,
   std::shared_ptr<std::list<T>> queue,
   std::shared_ptr<std::mutex> lock,
   std::shared_ptr<std::condition_variable> var)
   : Channel(closed, queue, lock, var)
  {}

  template <typename T>
  Receiver<T>::Receiver(
   std::shared_ptr<std::atomic<bool>> closed,
   std::shared_ptr<std::list<T>> queue,
   std::shared_ptr<std::mutex> lock,
   std::shared_ptr<std::condition_variable> var)
   : Channel(closed, queue, lock, var)
  {}

  template <typename T>
  void Sender<T>::send(T& item)
  {
    std::unique_lock<std::mutex> lock(*mLock);

    if (mClosed) {
      throw std::logic_error("tried to send on a closed channel");
    }

    mQueue->push_back(std::move(item));

    mVar->notify_one();
  }

  template <typename T>
  auto Receiver<T>::recv(T& msg) -> bool
  {
    std::unique_lock<std::mutex> lock(mWaitLock);

    mVar.wait(lock, [this] {
      return *mClosed || !mQueue->empty();
    });

    // queue modification takes place within this block
    {
      std::unique_lock<std::mutex>(*mLock);

      if (mQueue->empty()) {
        return false;
      }

      msg = std::move(mQueue->front());

      mQueue->pop_front();

      return true;
    }
  }

  template <typename T>
  auto Receiver<T>::hasItems() -> bool
  {
    std::unique_lock<std::mutex>(*mLock);
    return mQueue->size() > 0;
  }
}  // namespace util
