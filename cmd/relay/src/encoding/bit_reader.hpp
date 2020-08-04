#ifndef ENCODING_BIT_READER_HPP
#define ENCODING_BIT_READER_HPP

namespace encoding
{
  class BitReader
  {
   public:
    /**
        Bit reader constructor.
        Non-multiples of four buffer sizes are supported, as this naturally tends to occur when packets are read from the
        network. However, actual buffer allocated for the packet data must round up at least to the next 4 bytes in memory,
        because the bit reader reads dwords from memory not bytes.
        @param data Pointer to the bitpacked data to read.
        @param bytes The number of bytes of bitpacked data to read.
     */

    BitReader(const void* data, int bytes);

    /**
        Would the bit reader would read past the end of the buffer if it read this many bits?
        @param bits The number of bits that would be read.
        @returns True if reading the number of bits would read past the end of the buffer.
     */

    bool WouldReadPastEnd(int bits) const;

    /**
        Read bits from the bit buffer.
        This function will assert in debug builds if this read would read past the end of the buffer.
        In production situations, the higher level ReadStream takes care of checking all packet data and never calling this
        function if it would read past the end of the buffer.
        @param bits The number of bits to read in [1,32].
        @returns The integer value read in range [0,(1<<bits)-1].
     */

    uint32_t ReadBits(int bits);

    /**
        Read an align.
        Call this on read to correspond to a WriteAlign call when the bitpacked buffer was written.
        This makes sure we skip ahead to the next aligned byte index. As a safety check, we verify that the padding to next
       byte is zero bits and return false if that's not the case. This will typically abort packet read. Just another safety
       measure...
        @returns True if we successfully read an align and skipped ahead past zero pad, false otherwise (probably means, no
       align was written to the stream).
     */

    bool ReadAlign();

    /**
        Read bytes from the bitpacked data.
     */

    void ReadBytes(uint8_t* data, int bytes);

    /**
        How many align bits would be read, if we were to read an align right now?
        @returns Result in [0,7], where 0 is zero bits required to align (already aligned) and 7 is worst case.
     */

    int GetAlignBits() const;

    /**
        How many bits have we read so far?
        @returns The number of bits read from the bit buffer so far.
     */

    int GetBitsRead() const;

    /**
        How many bits are still available to read?
        For example, if the buffer size is 4, we have 32 bits available to read, if we have already written 10 bytes then 22
       are still available.
        @returns The number of bits available to read.
     */

    int GetBitsRemaining() const;

   private:
    const uint32_t* m_data;  ///< The bitpacked data we're reading as a dword array.
    uint64_t m_scratch;      ///< The scratch value. New data is read in 32 bits at a top to the left of this buffer, and data
                             ///< is read off to the right.
    int m_numBits;           ///< Number of bits to read in the buffer. Of course, we can't *really* know this so it's actually
                             ///< m_numBytes * 8.
    int m_numBytes;          ///< Number of bytes to read in the buffer. We know this, and this is the non-rounded up version.
#ifndef NDEBUG
    int m_numWords;     ///< Number of words to read in the buffer. This is rounded up to the next word if necessary.
#endif                  // #ifndef NDEBUG
    int m_bitsRead;     ///< Number of bits read from the buffer so far.
    int m_scratchBits;  ///< Number of bits currently in the scratch value. If the user wants to read more bits than this,
                        ///< we have to go fetch another dword from memory.
    int m_wordIndex;    ///< Index of the next word to read from memory.
  };
}  // namespace encoding
#endif
