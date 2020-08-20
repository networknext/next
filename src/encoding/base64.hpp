#pragma once

#include "util/logger.hpp"
#include "util/macros.hpp"

namespace encoding
{
  namespace base64
  {
    template <typename In, typename Out>
    INLINE auto encode(const In& input, Out& output) -> size_t
    {
      namespace b64 = boost::beast::detail::base64;
      return b64::encode(output.data(), input.data(), input.size());
    }

    template <typename In, typename Out>
    INLINE auto decode(const In& input, Out& output) -> size_t
    {
      namespace b64 = boost::beast::detail::base64;
      auto [written, read] = b64::decode(output.data(), input.data(), input.size());
      return written;
    }
  }  // namespace base64
}  // namespace encoding

/*
old encode func
      static const std::array<unsigned char, 65> base64_table_encode = {
       'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V',
       'W', 'X', 'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r',
       's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '/',
      };

      char* pos;
      const uint8_t* end;
      const uint8_t* in;

      size_t output_length = 4 * ((input.size() + 2) / 3);  // 3-byte blocks to 4-byte

      if (output.size() < input.size()) {
        LOG(ERROR, "could not encode base64 data, ouput buffer smaller than input");
        return 0;  // integer overflow
      }

      if (output_length >= output_size) {
        return 0;  // not enough room in output buffer
      }

      end = input + input_length;
      in = input;
      pos = output;
      while (end - in >= 3) {
        *pos++ = base64_table_encode[in[0] >> 2];
        *pos++ = base64_table_encode[((in[0] & 0x03) << 4) | (in[1] >> 4)];
        *pos++ = base64_table_encode[((in[1] & 0x0f) << 2) | (in[2] >> 6)];
        *pos++ = base64_table_encode[in[2] & 0x3f];
        in += 3;
      }

      if (end - in) {
        *pos++ = base64_table_encode[in[0] >> 2];
        if (end - in == 1) {
          *pos++ = base64_table_encode[(in[0] & 0x03) << 4];
          *pos++ = '=';
        } else {
          *pos++ = base64_table_encode[((in[0] & 0x03) << 4) | (in[1] >> 4)];
          *pos++ = base64_table_encode[(in[1] & 0x0f) << 2];
        }
        *pos++ = '=';
      }

      output[output_length] = '\0';

old decode func
      static const std::array<int, 256> base64_table_decode = {
       0, 0, 0,  0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
       0, 0, 0,  0, 0,  0,  0,  0,  0,  0,  0,  0,  62, 63, 62, 62, 63, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 0,  0,  0,  0,
       0, 0, 0,  0, 1,  2,  3,  4,  5,  6,  7,  8,  9,  10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 0,  0,
       0, 0, 63, 0, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
      };

      size_t input_length = strlen(input);
      int pad = input_length > 0 && (input_length % 4 || input[input_length - 1] == '=');
      size_t L = ((input_length + 3) / 4 - pad) * 4;
      size_t output_length = L / 4 * 3 + pad;

      if (output_length > output_size) {
        return 0;
      }

      for (size_t i = 0, j = 0; i < L; i += 4) {
        int n = base64_table_decode[int(input[i])] << 18 | base64_table_decode[int(input[i + 1])] << 12 |
                base64_table_decode[int(input[i + 2])] << 6 | base64_table_decode[int(input[i + 3])];
        output[j++] = uint8_t(n >> 16);
        output[j++] = uint8_t(n >> 8 & 0xFF);
        output[j++] = uint8_t(n & 0xFF);
      }

      if (pad) {
        int n = base64_table_decode[int(input[L])] << 18 | base64_table_decode[int(input[L + 1])] << 12;
        output[output_length - 1] = uint8_t(n >> 16);

        if (input_length > L + 2 && input[L + 2] != '=') {
          n |= base64_table_decode[int(input[L + 2])] << 6;
          output_length += 1;
          if (output_length > output_size) {
            return 0;
          }
          output[output_length - 1] = uint8_t(n >> 8 & 0xFF);
        }
      }

      return int(output_length);
*/
