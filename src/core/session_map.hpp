#pragma once

#include "session.hpp"
#include "router_info.hpp"

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
    void purge(const RouterInfo& router_info);

   private:
    // Using a map for now, it's a uint key so an unordered map might
    // not be any better considering the memory footprint
    std::map<uint64_t, SessionPtr> internal_map;
    mutable std::mutex mutex;

    /* Checks if the entry exists, however it does not check if it's a nullptr */
    auto exists(uint64_t key) const -> bool;
  };

  INLINE void SessionMap::set(uint64_t key, SessionPtr val)
  {
    std::lock_guard<std::mutex> lk(this->mutex);
    this->internal_map[key] = val;
  }

  INLINE auto SessionMap::get(uint64_t key) -> SessionPtr
  {
    std::lock_guard<std::mutex> lk(this->mutex);
    return exists(key) ? this->internal_map[key] : nullptr;
  }

  INLINE auto SessionMap::erase(uint64_t key) -> bool
  {
    std::lock_guard<std::mutex> lk(this->mutex);
    return this->internal_map.erase(key) > 0;
  }

  INLINE auto SessionMap::size() const -> size_t
  {
    std::lock_guard<std::mutex> lk(this->mutex);
    return this->internal_map.size();
  }

  INLINE void SessionMap::purge(const RouterInfo& router_info)
  {
    std::lock_guard<std::mutex> lk(this->mutex);
    auto iter = this->internal_map.begin();
    while (iter != this->internal_map.end()) {
      if (iter->second && iter->second->expired(router_info)) {
        iter = this->internal_map.erase(iter);
      } else {
        iter++;
      }
    }
  }

  /* Don't use a mutex, locking here will create a deadlock */
  INLINE auto SessionMap::exists(uint64_t key) const -> bool
  {
    return this->internal_map.find(key) != this->internal_map.end();
  }
}  // namespace core
