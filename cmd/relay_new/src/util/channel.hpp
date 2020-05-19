#pragma once

namespace util
{
  template <typename T>
  class Channel
  {
   public:
    virtual ~Channel() = default;

    void close();

    auto closed() -> bool;

    auto size() -> size_t;

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
    std::shared_ptr<std::mutex> mWaitLock;
  };

  template <typename T>
  inline std::tuple<Sender<T>, Receiver<T>> makeChannel()
  {
    auto closed = std::make_shared<std::atomic<bool>>(false);
    auto lock = std::make_shared<std::mutex>();
    auto queue = std::make_shared<std::list<T>>();
    auto var = std::make_shared<std::condition_variable>();

    Sender<T> sender(closed, queue, lock, var);
    Receiver<T> receiver(closed, queue, lock, var);

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
  Sender<T>::Sender(
   std::shared_ptr<std::atomic<bool>> closed,
   std::shared_ptr<std::list<T>> queue,
   std::shared_ptr<std::mutex> lock,
   std::shared_ptr<std::condition_variable> var)
   : Channel<T>(closed, queue, lock, var)
  {}

  template <typename T>
  Receiver<T>::Receiver(
   std::shared_ptr<std::atomic<bool>> closed,
   std::shared_ptr<std::list<T>> queue,
   std::shared_ptr<std::mutex> lock,
   std::shared_ptr<std::condition_variable> var)
   : Channel<T>(closed, queue, lock, var), mWaitLock(std::make_shared<std::mutex>())
  {}

  template <typename T>
  void Channel<T>::close()
  {
    std::unique_lock<std::mutex> lock(*mLock);
    *mClosed = true;
    mVar->notify_all();
  }

  template <typename T>
  auto Channel<T>::closed() -> bool
  {
    return *mClosed;
  }

  template <typename T>
  auto Channel<T>::size() -> size_t
  {
    return this->mQueue->size();
  }

  template <typename T>
  void Sender<T>::send(T& item)
  {
    std::unique_lock<std::mutex> lock(*this->mLock);

    if (*this->mClosed) {
      LogDebug("tried to send on a closed channel");
      return;
    }

    this->mQueue->push_back(std::move(item));

    this->mVar->notify_one();
  }

  template <typename T>
  auto Receiver<T>::recv(T& msg) -> bool
  {
    std::unique_lock<std::mutex> lock(*this->mWaitLock);

    this->mVar->wait(lock, [this] {
      return *this->mClosed || !this->mQueue->empty();
    });

    // queue modification takes place within this block
    {
      std::unique_lock<std::mutex> lock(*this->mLock);

      if (this->mQueue->empty()) {
        return false;
      }

      msg = std::move(this->mQueue->front());

      this->mQueue->pop_front();

      return true;
    }
  }

  template <typename T>
  auto Receiver<T>::hasItems() -> bool
  {
    std::unique_lock<std::mutex>(*this->mLock);
    return this->mQueue->size() > 0;
  }
}  // namespace util
