#ifndef RELAY_RELAY_HPP
#define RELAY_RELAY_HPP

#include "net/address.hpp"
#include "core/relay_manager.hpp"

#include "core/replay_protection.hpp"
#include "core/packets/types.hpp"

namespace relay
{
  uint64_t relay_clean_sequence(uint64_t sequence);
}  // namespace relay
#endif
