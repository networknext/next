#pragma once

#include "session.hpp"
#include "router_info.hpp"

namespace core
{
  /*
   * Each method locks a mutex so that this map can be
   * used with multiple threads without worry
   */

  // TODO this needs to use sharding

  class SessionMap
  {
   public:
    SessionMap();

    /* Emplace a new entry into the map. Overwrites if it already exists */
    void set(uint64_t key, SessionPtr val);

    /* Get the specified entry */
    auto get(uint64_t key) -> SessionPtr;

    /* Erase the specified entry, returns true if it did, false otherwise */
    auto erase(uint64_t key) -> bool;

    /* Return the number of elements in the map */
    auto size() const -> size_t;

    /* Remove all entries past the given timestamp */
    void purge(const uint64_t backend_timestamp);

    auto envelope_up_total() const -> size_t;

    auto envelope_down_total() const -> size_t;

   private:
    // Using a map for now, it's a uint key so an unordered map might
    // not be any better considering the memory footprint
    std::map<uint64_t, SessionPtr> internal_map;
    mutable std::mutex mutex;

    std::atomic<size_t> envelope_bandwidth_kbps_up;
    std::atomic<size_t> envelope_bandwidth_kbps_down;
  };

  INLINE SessionMap::SessionMap(): envelope_bandwidth_kbps_up(0), envelope_bandwidth_kbps_down(0) {}

  INLINE void SessionMap::set(uint64_t key, SessionPtr val)
  {
    std::lock_guard<std::mutex> lk(this->mutex);
    this->internal_map[key] = val;
    this->envelope_bandwidth_kbps_up += val->kbps_up;
    this->envelope_bandwidth_kbps_down += val->kbps_down;
  }

  INLINE auto SessionMap::get(uint64_t key) -> SessionPtr
  {
    std::lock_guard<std::mutex> lk(this->mutex);
    // don't create an entry if it doesn't exist
    return this->internal_map.find(key) != this->internal_map.end() ? this->internal_map[key] : nullptr;
  }

  INLINE auto SessionMap::erase(uint64_t key) -> bool
  {
    std::lock_guard<std::mutex> lk(this->mutex);
    bool existed = false;
    if (this->internal_map.find(key) != this->internal_map.end()) {
      existed = true;
      auto v = this->internal_map[key];
      this->envelope_bandwidth_kbps_up -= v->kbps_up;
      this->envelope_bandwidth_kbps_down -= v->kbps_down;
      this->internal_map.erase(key);
    }
    return existed;
  }

  INLINE auto SessionMap::size() const -> size_t
  {
    std::lock_guard<std::mutex> lk(this->mutex);
    return this->internal_map.size();
  }

  INLINE void SessionMap::purge(const uint64_t backend_timestamp)
  {
    std::lock_guard<std::mutex> lk(this->mutex);
    auto iter = this->internal_map.begin();
    while (iter != this->internal_map.end()) {
      if (iter->second && iter->second->expired(backend_timestamp)) {
        this->envelope_bandwidth_kbps_up -= iter->second->kbps_up;
        this->envelope_bandwidth_kbps_down -= iter->second->kbps_down;
        iter = this->internal_map.erase(iter);
      } else {
        iter++;
      }
    }
  }

  INLINE auto SessionMap::envelope_up_total() const -> size_t
  {
    return this->envelope_bandwidth_kbps_up;
  }

  INLINE auto SessionMap::envelope_down_total() const -> size_t
  {
    return this->envelope_bandwidth_kbps_down;
  }
}  // namespace core
