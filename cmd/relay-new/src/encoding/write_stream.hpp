#ifndef ENCODING_WRITE_STREAM_HPP
#define ENCODING_WRITE_STREAM_HPP

#include "base_stream.hpp"
#include "bit_writer.hpp"

namespace encoding
{
  /**
      Stream class for writing bitpacked data.
      This class is a wrapper around the bit writer class. Its purpose is to provide unified interface for reading and
     writing. You can determine if you are writing to a stream by calling Stream::IsWriting inside your templated serialize
     method. This is evaluated at compile time, letting the compiler generate optimized serialize functions without the hassle
     of maintaining separate read and write functions. IMPORTANT: Generally, you don't call methods on this class directly.
     Use the serialize_* macros instead. See test/shared.h for some examples.
   */

  class WriteStream: public BaseStream
  {
   public:
    enum
    {
      IsWriting = 1
    };
    enum
    {
      IsReading = 0
    };

    /**
        Write stream constructor.
        @param buffer The buffer to write to.
        @param bytes The number of bytes in the buffer. Must be a multiple of four.
        @param allocator The allocator to use for stream allocations. This lets you dynamically allocate memory as you read
        and write packets.
     */

    WriteStream(uint8_t* buffer, int bytes);

    /**
        Serialize an integer (write).
        @param value The integer value in [min,max].
        @param min The minimum value.
        @param max The maximum value.
        @returns Always returns true. All checking is performed by debug asserts only on write.
     */

    bool SerializeInteger(int32_t value, int32_t min, int32_t max);

    /**
        Serialize a number of bits (write).
        @param value The unsigned integer value to serialize. Must be in range [0,(1<<bits)-1].
        @param bits The number of bits to write in [1,32].
        @returns Always returns true. All checking is performed by debug asserts on write.
     */

    bool SerializeBits(uint32_t value, int bits);

    /**
        Serialize an array of bytes (write).
        @param data Array of bytes to be written.
        @param bytes The number of bytes to write.
        @returns Always returns true. All checking is performed by debug asserts on write.
     */

    bool SerializeBytes(const uint8_t* data, int bytes);

    /**
        Serialize an align (write).
        @returns Always returns true. All checking is performed by debug asserts on write.
     */

    bool SerializeAlign();

    /**
        If we were to write an align right now, how many bits would be required?
        @returns The number of zero pad bits required to achieve byte alignment in [0,7].
     */

    int GetAlignBits() const;

    /**
        Flush the stream to memory after you finish writing.
        Always call this after you finish writing and before you call WriteStream::GetData, or you'll potentially truncate
       the last dword of data you wrote.
     */

    void Flush();

    /**
        Get a pointer to the data written by the stream.
        IMPORTANT: Call WriteStream::Flush before you call this function!
        @returns A pointer to the data written by the stream
     */

    const uint8_t* GetData() const;

    /**
        How many bytes have been written so far?
        @returns Number of bytes written. This is effectively the packet size.
     */

    int GetBytesProcessed() const;

    /**
        Get number of bits written so far.
        @returns Number of bits written.
     */

    int GetBitsProcessed() const;

   private:
    BitWriter m_writer;  ///< The bit writer used for all bitpacked write operations.
  };
}  // namespace encoding
#endif
