#include "includes.h"
#include "bit_writer.hpp"

#include "net/net.hpp"

namespace encoding
{
  BitWriter::BitWriter(void* data, int bytes): m_data((uint32_t*)data), m_numWords(bytes / 4)
  {
    assert(data);
    assert((bytes % 4) == 0);
    m_numBits = m_numWords * 32;
    m_bitsWritten = 0;
    m_wordIndex = 0;
    m_scratch = 0;
    m_scratchBits = 0;
  }

  void BitWriter::WriteBits(uint32_t value, int bits)
  {
    assert(bits > 0);
    assert(bits <= 32);
    assert(m_bitsWritten + bits <= m_numBits);
    assert(uint64_t(value) <= ((1ULL << bits) - 1));

    m_scratch |= uint64_t(value) << m_scratchBits;

    m_scratchBits += bits;

    if (m_scratchBits >= 32) {
      assert(m_wordIndex < m_numWords);
      m_data[m_wordIndex] = net::host_to_network(uint32_t(m_scratch & 0xFFFFFFFF));
      m_scratch >>= 32;
      m_scratchBits -= 32;
      m_wordIndex++;
    }

    m_bitsWritten += bits;
  }

  void BitWriter::WriteAlign()
  {
    const int remainderBits = m_bitsWritten % 8;

    if (remainderBits != 0) {
      uint32_t zero = 0;
      WriteBits(zero, 8 - remainderBits);
      assert((m_bitsWritten % 8) == 0);
    }
  }

  void BitWriter::WriteBytes(const uint8_t* data, int bytes)
  {
    assert(GetAlignBits() == 0);
    assert(m_bitsWritten + bytes * 8 <= m_numBits);
    assert((m_bitsWritten % 32) == 0 || (m_bitsWritten % 32) == 8 || (m_bitsWritten % 32) == 16 || (m_bitsWritten % 32) == 24);

    int headBytes = (4 - (m_bitsWritten % 32) / 8) % 4;
    if (headBytes > bytes)
      headBytes = bytes;
    for (int i = 0; i < headBytes; ++i)
      WriteBits(data[i], 8);
    if (headBytes == bytes)
      return;

    FlushBits();

    assert(GetAlignBits() == 0);

    int numWords = (bytes - headBytes) / 4;
    if (numWords > 0) {
      assert((m_bitsWritten % 32) == 0);
      memcpy(&m_data[m_wordIndex], data + headBytes, numWords * 4);
      m_bitsWritten += numWords * 32;
      m_wordIndex += numWords;
      m_scratch = 0;
    }

    assert(GetAlignBits() == 0);

    int tailStart = headBytes + numWords * 4;
    int tailBytes = bytes - tailStart;
    assert(tailBytes >= 0 && tailBytes < 4);
    for (int i = 0; i < tailBytes; ++i)
      WriteBits(data[tailStart + i], 8);

    assert(GetAlignBits() == 0);

    assert(headBytes + numWords * 4 + tailBytes == bytes);
  }

  void BitWriter::FlushBits()
  {
    if (m_scratchBits != 0) {
      assert(m_scratchBits <= 32);
      assert(m_wordIndex < m_numWords);
      m_data[m_wordIndex] = net::host_to_network(uint32_t(m_scratch & 0xFFFFFFFF));
      m_scratch >>= 32;
      m_scratchBits = 0;
      m_wordIndex++;
    }
  }

  int BitWriter::GetAlignBits() const
  {
    return (8 - (m_bitsWritten % 8)) % 8;
  }

  int BitWriter::GetBitsWritten() const
  {
    return m_bitsWritten;
  }

  int BitWriter::GetBitsAvailable() const
  {
    return m_numBits - m_bitsWritten;
  }

  const uint8_t* BitWriter::GetData() const
  {
    return (uint8_t*)m_data;
  }

  int BitWriter::GetBytesWritten() const
  {
    return (m_bitsWritten + 7) / 8;
  }
}  // namespace encoding
