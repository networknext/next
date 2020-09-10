#include "includes.h"
#include "testing/test.hpp"

#include "encoding/base64.hpp"

TEST(encoding_base64_general)
{
  std::string input = "a test string. let's see if it works properly";
  std::array<char, 1024> encoded = {};
  std::array<char, 1024> decoded = {};
  CHECK(encoding::base64::encode(input, encoded) > 0);
  size_t decoded_length = encoding::base64::decode(encoded, decoded);
  CHECK(decoded_length > 0);

  std::string output(decoded.begin(), decoded.begin() + decoded_length);
  CHECK(input == output).on_fail([&] {
    std::cout << "input = " << input;
    std::cout << "output = " << output;
  });
}
