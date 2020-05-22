#ifndef CORE_SESSION_MAP_HPP
#define CORE_SESSION_MAP_HPP

#include "session.hpp"

namespace core
{
  /*
   * Each method locks a mutex so that this map can be
   * used with multiple threads without worry
   */

  class SessionMap
  {
   public:
    SessionMap() = default;

    /* Emplace a new entry into the map */
    void set(uint64_t key, SessionPtr val);

    /* Get the specified entry */
    auto get(uint64_t key) -> SessionPtr;

    /* Erase the specified entry, returns true if it did, false otherwise */
    auto erase(uint64_t key) -> bool;

    /* Return the number of elements in the map */
    auto size() const -> size_t;

    /* Remove all entries past the given timestamp */
    void purge(double seconds);

   private:
    // Using a map for now, it's a uint key so an unordered map might
    // not be any better considering the memory footprint
    std::map<uint64_t, SessionPtr> mInternal;
    mutable std::mutex mLock;

    /* Checks if the entry exists, however it does not check if it's a nullptr */
    auto exists(uint64_t key) const -> bool;
  };

  inline void SessionMap::set(uint64_t key, SessionPtr val)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mInternal[key] = val;
  }

  inline auto SessionMap::get(uint64_t key) -> SessionPtr
  {
    std::lock_guard<std::mutex> lk(mLock);
    return exists(key) ? mInternal[key] : nullptr;
  }

  inline auto SessionMap::erase(uint64_t key) -> bool
  {
    std::lock_guard<std::mutex> lk(mLock);
    return mInternal.erase(key) > 0;
  }

  inline auto SessionMap::size() const -> size_t
  {
    std::lock_guard<std::mutex> lk(mLock);
    return mInternal.size();
  }

  inline void SessionMap::purge(double seconds)
  {
    std::lock_guard<std::mutex> lk(mLock);
    auto iter = mInternal.begin();
    while (iter != mInternal.end()) {
      if (iter->second && iter->second->expired(seconds)) {
        iter = mInternal.erase(iter);
      } else {
        iter++;
      }
    }
  }

  /* Don't use a mutex, locking here will create a deadlock */
  inline auto SessionMap::exists(uint64_t key) const -> bool
  {
    return mInternal.find(key) != mInternal.end();
  }
}  // namespace core
#endif