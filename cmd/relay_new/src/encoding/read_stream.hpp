#ifndef ENCODING_READ_STREAM
#define ENCODING_READ_STREAM

#include "base_stream.hpp"
#include "bit_reader.hpp"

namespace encoding
{
  /**
      Stream class for reading bitpacked data.
      This class is a wrapper around the bit reader class. Its purpose is to provide unified interface for reading and
     writing. You can determine if you are reading from a stream by calling Stream::IsReading inside your templated serialize
     method. This is evaluated at compile time, letting the compiler generate optimized serialize functions without the hassle
     of maintaining separate read and write functions. IMPORTANT: Generally, you don't call methods on this class directly.
     Use the serialize_* macros instead. See test/shared.h for some examples.
   */

  class ReadStream: public BaseStream
  {
   public:
    enum
    {
      IsWriting = 0
    };
    enum
    {
      IsReading = 1
    };

    /**
        Read stream constructor.
        @param buffer The buffer to read from.
        @param bytes The number of bytes in the buffer. May be a non-multiple of four, however if it is, the underlying
       buffer allocated should be large enough to read the any remainder bytes as a dword.
        @param allocator The allocator to use for stream allocations. This lets you dynamically allocate memory as you read
       and write packets.
     */

    ReadStream(const uint8_t* buffer, int bytes);

    /**
        Serialize an integer (read).
        @param value The integer value read is stored here. It is guaranteed to be in [min,max] if this function succeeds.
        @param min The minimum allowed value.
        @param max The maximum allowed value.
        @returns Returns true if the serialize succeeded and the value is in the correct range. False otherwise.
     */

    bool SerializeInteger(int32_t& value, int32_t min, int32_t max);

    /**
        Serialize a number of bits (read).
        @param value The integer value read is stored here. Will be in range [0,(1<<bits)-1].
        @param bits The number of bits to read in [1,32].
        @returns Returns true if the serialize read succeeded, false otherwise.
     */

    bool SerializeBits(uint32_t& value, int bits);

    /**
        Serialize an array of bytes (read).
        @param data Array of bytes to read.
        @param bytes The number of bytes to read.
        @returns Returns true if the serialize read succeeded. False otherwise.
     */

    bool SerializeBytes(uint8_t* data, int bytes);

    /**
        Serialize an align (read).
        @returns Returns true if the serialize read succeeded. False otherwise.
     */

    bool SerializeAlign();

    /**
        If we were to read an align right now, how many bits would we need to read?
        @returns The number of zero pad bits required to achieve byte alignment in [0,7].
     */

    int GetAlignBits() const;

    /**
        Get number of bits read so far.
        @returns Number of bits read.
     */

    int GetBitsProcessed() const;

    /**
        How many bytes have been read so far?
        @returns Number of bytes read. Effectively this is the number of bits read, rounded up to the next byte where
       necessary.
     */

    int GetBytesProcessed() const;

   private:
    BitReader m_reader;  ///< The bit reader used for all bitpacked read operations.
  };
}  // namespace encoding
#endif
