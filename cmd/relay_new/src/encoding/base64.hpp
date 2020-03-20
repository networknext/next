#ifndef ENCODING_BASE64_HPP
#define ENCODING_BASE64_HPP

namespace legacy
{
  int base64_encode_data(const uint8_t* input, size_t input_length, char* output, size_t output_size);
  int base64_decode_data(const char* input, uint8_t* output, size_t output_size);
  int base64_encode_string(const char* input, char* output, size_t output_size);
  int base64_decode_string(const char* input, char* output, size_t output_size);
}  // namespace legacy

namespace encoding
{
  namespace base64
  {
    template <typename T>
    bool EncodeToString(const T& input, std::string& output)
    {
      // TODO make this retval actually matter
      return legacy::base64_encode_data(input.data(), input.size(), output.data(), output.size()) > 0;
    }

    template <typename T>
    bool DecodeString(const std::string& input, T& output)
    {
      // TODO and this
      return legacy::base64_decode_data(input.data(), output.data(), output.size()) > 0;
    }
  }  // namespace base64
}  // namespace encoding
#endif
