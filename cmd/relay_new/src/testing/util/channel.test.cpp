#include "includes.h"
#include "testing/test.hpp"
#include "util/channel.hpp"

using namespace std::chrono_literals;

Test(util_Channel_general, true)
{
  auto chan = util::makeChannel<int>();
  auto sender = std::get<0>(chan);
  auto receiver = std::get<1>(chan);

  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(1s);
    sender.send(1);
  });

  int result = 0;
  check(receiver.recv(result));
  check(result == 1);
}
