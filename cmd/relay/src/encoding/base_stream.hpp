#ifndef ENCODING_BASE_STREAM_HPP
#define ENCODING_BASE_STREAM_HPP

#include "net/address.hpp"

namespace encoding
{
  /**
      Reads bit packed integer values from a buffer.
      Relies on the user reconstructing the exact same set of bit reads as bit writes when the buffer was written. This is an
     unattributed bitpacked binary stream! Implementation: 32 bit dwords are read in from memory to the high bits of a scratch
     value as required. The user reads off bit values from the scratch value from the right, after which the scratch value is
     shifted by the same number of bits.
   */

  /**
      Functionality common to all stream classes.
   */

  class BaseStream
  {
   public:
    /**
        Base stream constructor.
     */
    explicit BaseStream();

    /**
        Set a context on the stream.
     */

    void SetContext(void* context);

    /**
        Get the context pointer set on the stream.
        @returns The context pointer. May be NULL.
     */

    void* GetContext() const;

   private:
    void* m_context;  ///< The context pointer set on the stream. May be NULL.
  };

  /**
      Serialize integer value (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param value The integer value to serialize in [min,max].
      @param min The minimum value.
      @param max The maximum value.
   */

#define serialize_int(stream, value, min, max)                              \
  do {                                                                      \
    assert(min < max);                                                      \
    int32_t int32_value = 0;                                                \
    if (Stream::IsWriting) {                                                \
      assert(int64_t(value) >= int64_t(min));                               \
      assert(int64_t(value) <= int64_t(max));                               \
      int32_value = (int32_t)value;                                         \
    }                                                                       \
    if (!stream.SerializeInteger(int32_value, min, max)) {                  \
      return false;                                                         \
    }                                                                       \
    if (Stream::IsReading) {                                                \
      value = int32_value;                                                  \
      if (int64_t(value) < int64_t(min) || int64_t(value) > int64_t(max)) { \
        return false;                                                       \
      }                                                                     \
    }                                                                       \
  } while (0)

  /**
      Serialize bits to the stream (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param value The unsigned integer value to serialize.
      @param bits The number of bits to serialize in [1,32].
   */

#define serialize_bits(stream, value, bits)          \
  do {                                               \
    assert(bits > 0);                                \
    assert(bits <= 32);                              \
    uint32_t uint32_value = 0;                       \
    if (Stream::IsWriting) {                         \
      uint32_value = (uint32_t)value;                \
    }                                                \
    if (!stream.SerializeBits(uint32_value, bits)) { \
      return false;                                  \
    }                                                \
    if (Stream::IsReading) {                         \
      value = uint32_value;                          \
    }                                                \
  } while (0)

  /**
      Serialize a boolean value to the stream (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param value The boolean value to serialize.
   */

#define serialize_bool(stream, value)             \
  do {                                            \
    uint32_t uint32_bool_value = 0;               \
    if (Stream::IsWriting) {                      \
      uint32_bool_value = value ? 1 : 0;          \
    }                                             \
    serialize_bits(stream, uint32_bool_value, 1); \
    if (Stream::IsReading) {                      \
      value = uint32_bool_value ? true : false;   \
    }                                             \
  } while (0)

  template <typename Stream>
  bool serialize_float_internal(Stream& stream, float& value)
  {
    uint32_t int_value;
    if (Stream::IsWriting) {
      memcpy(&int_value, &value, 4);
    }
    bool result = stream.SerializeBits(int_value, 32);
    if (Stream::IsReading && result) {
      memcpy(&value, &int_value, 4);
    }
    return result;
  }

  /**
      Serialize floating point value (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param value The float value to serialize.
   */

#define serialize_float(stream, value)              \
  do {                                              \
    if (!serialize_float_internal(stream, value)) { \
      return false;                                 \
    }                                               \
  } while (0)

  /**
      Serialize a 32 bit unsigned integer to the stream (read/write).
      This is a helper macro to make unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param value The unsigned 32 bit integer value to serialize.
   */

#define serialize_uint32(stream, value) serialize_bits(stream, value, 32);

  template <typename Stream>
  bool serialize_uint64_internal(Stream& stream, uint64_t& value)
  {
    uint32_t hi = 0, lo = 0;
    if (Stream::IsWriting) {
      lo = value & 0xFFFFFFFF;
      hi = value >> 32;
    }
    serialize_bits(stream, lo, 32);
    serialize_bits(stream, hi, 32);
    if (Stream::IsReading) {
      value = (uint64_t(hi) << 32) | lo;
    }
    return true;
  }

  /**
      Serialize a 64 bit unsigned integer to the stream (read/write).
      This is a helper macro to make unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param value The unsigned 64 bit integer value to serialize.
   */

#define serialize_uint64(stream, value)            \
  do {                                             \
    if (!serialize_uint64_internal(stream, value)) \
      return false;                                \
  } while (0)

  template <typename Stream>
  bool serialize_double_internal(Stream& stream, double& value)
  {
    union DoubleInt
    {
      double double_value;
      uint64_t int_value;
    };
    DoubleInt tmp = {0};
    if (Stream::IsWriting) {
      tmp.double_value = value;
    }
    serialize_uint64(stream, tmp.int_value);
    if (Stream::IsReading) {
      value = tmp.double_value;
    }
    return true;
  }

  /**
      Serialize double precision floating point value to the stream (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param value The double precision floating point value to serialize.
   */

#define serialize_double(stream, value)              \
  do {                                               \
    if (!serialize_double_internal(stream, value)) { \
      return false;                                  \
    }                                                \
  } while (0)

  template <typename Stream>
  bool serialize_bytes_internal(Stream& stream, uint8_t* data, int bytes)
  {
    return stream.SerializeBytes(data, bytes);
  }

  /**
      Serialize an array of bytes to the stream (read/write).
      This is a helper macro to make unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param data Pointer to the data to be serialized.
      @param bytes The number of bytes to serialize.
   */

#define serialize_bytes(stream, data, bytes)              \
  do {                                                    \
    if (!serialize_bytes_internal(stream, data, bytes)) { \
      return false;                                       \
    }                                                     \
  } while (0)

  template <typename Stream>
  bool serialize_string_internal(Stream& stream, char* string, int buffer_size)
  {
    int length = 0;
    if (Stream::IsWriting) {
      length = (int)strlen(string);
      assert(length < buffer_size);
    }
    serialize_int(stream, length, 0, buffer_size - 1);
    serialize_bytes(stream, (uint8_t*)string, length);
    if (Stream::IsReading) {
      string[length] = '\0';
    }
    return true;
  }

  /**
      Serialize a string to the stream (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param string The string to serialize write. Pointer to buffer to be filled on read.
      @param buffer_size The size of the string buffer. String with terminating null character must fit into this buffer.
   */

#define serialize_string(stream, string, buffer_size)              \
  do {                                                             \
    if (!serialize_string_internal(stream, string, buffer_size)) { \
      return false;                                                \
    }                                                              \
  } while (0)

  /**
      Serialize an alignment to the stream (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
   */

#define serialize_align(stream)     \
  do {                              \
    if (!stream.SerializeAlign()) { \
      return false;                 \
    }                               \
  } while (0)

  /**
      Serialize an object to the stream (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param object The object to serialize. Must have a serialize method on it.
   */

#define serialize_object(stream, object) \
  do {                                   \
    if (!object.Serialize(stream)) {     \
      return false;                      \
    }                                    \
  } while (0)

  template <typename Stream, typename T>
  bool serialize_int_relative_internal(Stream& stream, T previous, T& current)
  {
    uint32_t difference = 0;
    if (Stream::IsWriting) {
      assert(previous < current);
      difference = current - previous;
    }

    bool oneBit = false;
    if (Stream::IsWriting) {
      oneBit = difference == 1;
    }
    serialize_bool(stream, oneBit);
    if (oneBit) {
      if (Stream::IsReading) {
        current = previous + 1;
      }
      return true;
    }

    bool twoBits = false;
    if (Stream::IsWriting) {
      twoBits = difference <= 6;
    }
    serialize_bool(stream, twoBits);
    if (twoBits) {
      serialize_int(stream, difference, 2, 6);
      if (Stream::IsReading) {
        current = previous + difference;
      }
      return true;
    }

    bool fourBits = false;
    if (Stream::IsWriting) {
      fourBits = difference <= 23;
    }
    serialize_bool(stream, fourBits);
    if (fourBits) {
      serialize_int(stream, difference, 7, 23);
      if (Stream::IsReading) {
        current = previous + difference;
      }
      return true;
    }

    bool eightBits = false;
    if (Stream::IsWriting) {
      eightBits = difference <= 280;
    }
    serialize_bool(stream, eightBits);
    if (eightBits) {
      serialize_int(stream, difference, 24, 280);
      if (Stream::IsReading) {
        current = previous + difference;
      }
      return true;
    }

    bool twelveBits = false;
    if (Stream::IsWriting) {
      twelveBits = difference <= 4377;
    }
    serialize_bool(stream, twelveBits);
    if (twelveBits) {
      serialize_int(stream, difference, 281, 4377);
      if (Stream::IsReading) {
        current = previous + difference;
      }
      return true;
    }

    bool sixteenBits = false;
    if (Stream::IsWriting) {
      sixteenBits = difference <= 69914;
    }
    serialize_bool(stream, sixteenBits);
    if (sixteenBits) {
      serialize_int(stream, difference, 4378, 69914);
      if (Stream::IsReading) {
        current = previous + difference;
      }
      return true;
    }

    uint32_t value = current;
    serialize_uint32(stream, value);
    if (Stream::IsReading) {
      current = value;
    }

    return true;
  }

  /**
      Serialize an integer value relative to another (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param previous The previous integer value.
      @param current The current integer value.
   */

#define serialize_int_relative(stream, previous, current)              \
  do {                                                                 \
    if (!serialize_int_relative_internal(stream, previous, current)) { \
      return false;                                                    \
    }                                                                  \
  } while (0)

  template <typename Stream>
  bool serialize_ack_relative_internal(Stream& stream, uint16_t sequence, uint16_t& ack)
  {
    int ack_delta = 0;
    bool ack_in_range = false;
    if (Stream::IsWriting) {
      if (ack < sequence) {
        ack_delta = sequence - ack;
      } else {
        ack_delta = (int)sequence + 65536 - ack;
      }
      assert(ack_delta > 0);
      assert(uint16_t(sequence - ack_delta) == ack);
      ack_in_range = ack_delta <= 64;
    }
    serialize_bool(stream, ack_in_range);
    if (ack_in_range) {
      serialize_int(stream, ack_delta, 1, 64);
      if (Stream::IsReading) {
        ack = sequence - ack_delta;
      }
    } else {
      serialize_bits(stream, ack, 16);
    }
    return true;
  }

  /**
      Serialize an ack relative to the current sequence number (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param sequence The current sequence number.
      @param ack The ack sequence number, which is typically near the current sequence number.
   */

#define serialize_ack_relative(stream, sequence, ack)              \
  do {                                                             \
    if (!serialize_ack_relative_internal(stream, sequence, ack)) { \
      return false;                                                \
    }                                                              \
  } while (0)

  template <typename Stream>
  bool serialize_sequence_relative_internal(Stream& stream, uint16_t sequence1, uint16_t& sequence2)
  {
    if (Stream::IsWriting) {
      uint32_t a = sequence1;
      uint32_t b = sequence2 + ((sequence1 > sequence2) ? 65536 : 0);
      serialize_int_relative(stream, a, b);
    } else {
      uint32_t a = sequence1;
      uint32_t b = 0;
      serialize_int_relative(stream, a, b);
      if (b >= 65536) {
        b -= 65536;
      }
      sequence2 = uint16_t(b);
    }

    return true;
  }

  /**
      Serialize a sequence number relative to another (read/write).
      This is a helper macro to make writing unified serialize functions easier.
      Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
     important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
     called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
     return value.
      @param stream The stream object. May be a read or write stream.
      @param sequence1 The first sequence number to serialize relative to.
      @param sequence2 The second sequence number to be encoded relative to the first.
   */

#define serialize_sequence_relative(stream, sequence1, sequence2)              \
  do {                                                                         \
    if (!serialize_sequence_relative_internal(stream, sequence1, sequence2)) { \
      return false;                                                            \
    }                                                                          \
  } while (0)

  template <typename Stream>
  bool serialize_address_internal(Stream& stream, legacy::relay_address_t& address)
  {
    serialize_bits(stream, address.type, 2);
    if (address.type == net::AddressType::IPv4) {
      serialize_bytes(stream, address.data.ipv4, 4);
      serialize_bits(stream, address.port, 16);
    } else if (address.type == net::AddressType::IPv6) {
      for (int i = 0; i < 8; ++i) {
        serialize_bits(stream, address.data.ipv6[i], 16);
      }
      serialize_bits(stream, address.port, 16);
    } else {
      if (Stream::IsReading) {
        memset(&address, 0, sizeof(legacy::relay_address_t));
      }
    }
    return true;
  }

#define serialize_address(stream, address)              \
  do {                                                  \
    if (!serialize_address_internal(stream, address)) { \
      return false;                                     \
    }                                                   \
  } while (0)

}  // namespace encoding
#endif
