#include "includes.h"
#include "bit_reader.hpp"

#include "net/net.hpp"

namespace encoding
{
#ifndef NDEBUG
  BitReader::BitReader(const void* data, int bytes)
   : m_data((const uint32_t*)data), m_numBytes(bytes), m_numWords((bytes + 3) / 4)
#else   // #ifndef NDEBUG
  BitReader::BitReader(const void* data, int bytes): m_data((const uint32_t*)data), m_numBytes(bytes)
#endif  // #ifndef NDEBUG
  {
    assert(data);
    m_numBits = m_numBytes * 8;
    m_bitsRead = 0;
    m_scratch = 0;
    m_scratchBits = 0;
    m_wordIndex = 0;
  }

  bool BitReader::WouldReadPastEnd(int bits) const
  {
    return m_bitsRead + bits > m_numBits;
  }

  uint32_t BitReader::ReadBits(int bits)
  {
    assert(bits > 0);
    assert(bits <= 32);
    assert(m_bitsRead + bits <= m_numBits);

    m_bitsRead += bits;

    assert(m_scratchBits >= 0 && m_scratchBits <= 64);

    if (m_scratchBits < bits) {
      assert(m_wordIndex < m_numWords);
      m_scratch |= uint64_t(net::network_to_host(m_data[m_wordIndex])) << m_scratchBits;
      m_scratchBits += 32;
      m_wordIndex++;
    }

    assert(m_scratchBits >= bits);

    const uint32_t output = m_scratch & ((uint64_t(1) << bits) - 1);

    m_scratch >>= bits;
    m_scratchBits -= bits;

    return output;
  }

  bool BitReader::ReadAlign()
  {
    const int remainderBits = m_bitsRead % 8;
    if (remainderBits != 0) {
      uint32_t value = ReadBits(8 - remainderBits);
      assert(m_bitsRead % 8 == 0);
      if (value != 0)
        return false;
    }
    return true;
  }

  void BitReader::ReadBytes(uint8_t* data, int bytes)
  {
    assert(GetAlignBits() == 0);
    assert(m_bitsRead + bytes * 8 <= m_numBits);
    assert((m_bitsRead % 32) == 0 || (m_bitsRead % 32) == 8 || (m_bitsRead % 32) == 16 || (m_bitsRead % 32) == 24);

    int headBytes = (4 - (m_bitsRead % 32) / 8) % 4;
    if (headBytes > bytes)
      headBytes = bytes;
    for (int i = 0; i < headBytes; ++i)
      data[i] = (uint8_t)ReadBits(8);
    if (headBytes == bytes)
      return;

    assert(GetAlignBits() == 0);

    int numWords = (bytes - headBytes) / 4;
    if (numWords > 0) {
      assert((m_bitsRead % 32) == 0);
      memcpy(data + headBytes, &m_data[m_wordIndex], numWords * 4);
      m_bitsRead += numWords * 32;
      m_wordIndex += numWords;
      m_scratchBits = 0;
    }

    assert(GetAlignBits() == 0);

    int tailStart = headBytes + numWords * 4;
    int tailBytes = bytes - tailStart;
    assert(tailBytes >= 0 && tailBytes < 4);
    for (int i = 0; i < tailBytes; ++i)
      data[tailStart + i] = (uint8_t)ReadBits(8);

    assert(GetAlignBits() == 0);

    assert(headBytes + numWords * 4 + tailBytes == bytes);
  }

  int BitReader::GetAlignBits() const
  {
    return (8 - m_bitsRead % 8) % 8;
  }

  int BitReader::GetBitsRead() const
  {
    return m_bitsRead;
  }

  int BitReader::GetBitsRemaining() const
  {
    return m_numBits - m_bitsRead;
  }
}  // namespace encoding
