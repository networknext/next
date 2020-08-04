#include "includes.h"
#include "testing/test.hpp"
#include "util/channel.hpp"

using namespace std::chrono_literals;

Test(util_Channel_general)
{
  auto chan = util::makeChannel<std::string>();
  auto sender = std::get<0>(chan);
  auto receiver = std::get<1>(chan);
  std::string msg = "test string";
  std::string result;

  sender.send(msg);

  check(receiver.recv(result));

  check(msg == "");
  check(result == "test string").onFail([&] {
    std::cout << "result is " << result << std::endl;
  });
}
