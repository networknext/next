#ifndef ENCODING_BIT_WRITER_HPP
#define ENCODING_BIT_WRITER_HPP

namespace encoding
{
  /**
      Bitpacks unsigned integer values to a buffer.
      Integer bit values are written to a 64 bit scratch value from right to left.
      Once the low 32 bits of the scratch is filled with bits it is flushed to memory as a dword and the scratch value is
     shifted right by 32. The bit stream is written to memory in little endian order, which is considered network byte order
     for this library.
   */

  class BitWriter
  {
   public:
    /**
        Bit writer constructor.
        Creates a bit writer object to write to the specified buffer.
        @param data The pointer to the buffer to fill with bitpacked data.
        @param bytes The size of the buffer in bytes. Must be a multiple of 4, because the bitpacker reads and writes memory
       as dwords, not bytes.
     */

    BitWriter(void* data, int bytes);

    /**
        Write bits to the buffer.
        Bits are written to the buffer as-is, without padding to nearest byte. Will assert if you try to write past the end
       of the buffer. A boolean value writes just 1 bit to the buffer, a value in range [0,31] can be written with just 5
       bits and so on. IMPORTANT: When you have finished writing to your buffer, take care to call BitWrite::FlushBits,
       otherwise the last dword of data will not get flushed to memory!
        @param value The integer value to write to the buffer. Must be in [0,(1<<bits)-1].
        @param bits The number of bits to encode in [1,32].
     */

    void WriteBits(uint32_t value, int bits);

    /**
        Write an alignment to the bit stream, padding zeros so the bit index becomes is a multiple of 8.
        This is useful if you want to write some data to a packet that should be byte aligned. For example, an array of
        bytes, or a string. IMPORTANT: If the current bit index is already a multiple of 8, nothing is written.
     */

    void WriteAlign();

    /**
        Write an array of bytes to the bit stream.
        Use this when you have to copy a large block of data into your bitstream.
        Faster than just writing each byte to the bit stream via BitWriter::WriteBits( value, 8 ), because it aligns to byte
        index and copies into the buffer without bitpacking.
        @param data The byte array data to write to the bit stream.
        @param bytes The number of bytes to write.
     */

    void WriteBytes(const uint8_t* data, int bytes);

    /**
        Flush any remaining bits to memory.
        Call this once after you've finished writing bits to flush the last dword of scratch to memory!
     */

    void FlushBits();

    /**
        How many align bits would be written, if we were to write an align right now?
        @returns Result in [0,7], where 0 is zero bits required to align (already aligned) and 7 is worst case.
     */

    int GetAlignBits() const;

    /**
        How many bits have we written so far?
        @returns The number of bits written to the bit buffer.
     */

    int GetBitsWritten() const;

    /**
        How many bits are still available to write?
        For example, if the buffer size is 4, we have 32 bits available to write, if we have already written 10 bytes then
       22 are still available to write.
        @returns The number of bits available to write.
     */

    int GetBitsAvailable() const;

    /**
        Get a pointer to the data written by the bit writer.
        Corresponds to the data block passed in to the constructor.
        @returns Pointer to the data written by the bit writer.
     */

    const uint8_t* GetData() const;

    /**
        The number of bytes flushed to memory.
        This is effectively the size of the packet that you should send after you have finished bitpacking values with this
       class. The returned value is not always a multiple of 4, even though we flush dwords to memory. You won't miss any
       data in this case because the order of bits written is designed to work with the little endian memory layout.
        IMPORTANT: Make sure you call BitWriter::FlushBits before calling this method, otherwise you risk missing the last
       dword of data.
     */

    int GetBytesWritten() const;

   private:
    uint32_t* m_data;    ///< The buffer we are writing to, as a uint32_t * because we're writing dwords at a time.
    uint64_t m_scratch;  ///< The scratch value where we write bits to (right to left). 64 bit for overflow. Once # of bits
                         ///< in scratch is >= 32, the low 32 bits are flushed to memory.
    int m_numBits;       ///< The number of bits in the buffer. This is equivalent to the size of the buffer in bytes multiplied
                         ///< by 8. Note that the buffer size must always be a multiple of 4.
    int m_numWords;      ///< The number of words in the buffer. This is equivalent to the size of the buffer in bytes divided
                         ///< by 4. Note that the buffer size must always be a multiple of 4.
    int m_bitsWritten;   ///< The number of bits written so far.
    int m_wordIndex;     ///< The current word index. The next word flushed to memory will be at this index in m_data.
    int m_scratchBits;   ///< The number of bits in scratch. When this is >= 32, the low 32 bits of scratch is flushed to
                         ///< memory as a dword and scratch is shifted right by 32.
  };
}  // namespace encoding
#endif
