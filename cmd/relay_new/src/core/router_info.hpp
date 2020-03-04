#ifndef CORE_ROUTER_INFO_HPP
#define CORE_ROUTER_INFO_HPP
namespace core
{
  struct RouterInfo
  {
    uint64_t InitalizeTimeInSeconds = 0; // from relay_init, Unix time (since jan 1 1970)
  };
}  // namespace core
#endif