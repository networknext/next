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

    /* Checks if the entry exists, however it does not check if it's a nullptr */
    auto exists(uint64_t key) const -> bool;

    /* Erase the specified entry, returns true if it did, false otherwise */
    auto erase(uint64_t key) -> bool;

    /* Return the number of elements in the map */
    auto size() const -> size_t;

    /* Remove all expired entries */
    void purge();

   private:
    // Using a map for now, it's a uint key so an unordered map might
    // not be any better considering the memory footprint
    std::map<uint64_t, SessionPtr> mInternal;
    mutable std::mutex mLock;
  };

  inline void SessionMap::set(uint64_t key, SessionPtr val)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mInternal[key] = val;
  }

  inline auto SessionMap::get(uint64_t key) -> SessionPtr
  {
    std::lock_guard<std::mutex> lk(mLock);
    return mInternal[key];
  }

  inline auto SessionMap::exists(uint64_t key) const -> bool
  {
    std::lock_guard<std::mutex> lk(mLock);
    return mInternal.find(key) != mInternal.end();
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

  inline void SessionMap::purge()
  {
    std::lock_guard<std::mutex> lk(mLock);
    auto iter = mInternal.begin();
    while (iter != mInternal.end()) {
      if (!iter->second || iter->second->expired()) {
        iter = mInternal.erase(iter);
      } else {
        iter++;
      }
    }
  }
}  // namespace core
#endif