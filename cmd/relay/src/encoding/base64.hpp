#ifndef ENCODING_BASE64_HPP
#define ENCODING_BASE64_HPP

#include <cstddef>
#include <cinttypes>

namespace encoding
{
    int base64_encode_data(const uint8_t* input, size_t input_length, char* output, size_t output_size);
    int base64_decode_data(const char* input, uint8_t* output, size_t output_size);
    int base64_encode_string(const char* input, char* output, size_t output_size);
    int base64_decode_string(const char* input, char* output, size_t output_size);
}  // namespace encoding
#endif