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
    template <typename IT, typename OT>
    size_t Encode(const IT& input, OT& output)
    {
      return legacy::base64_encode_data(input.data(), input.size(), output.data(), output.size());
    }

    template <typename IT, typename OT>
    size_t Decode(const IT& input, OT& output)
    {
      return legacy::base64_decode_data(input.data(), output.data(), output.size());
    }
  }  // namespace base64
}  // namespace encoding
#endif
