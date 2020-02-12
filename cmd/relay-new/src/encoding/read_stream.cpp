#include "includes.h"
#include "read_stream.hpp"

#include "encoding/binary.hpp"

namespace encoding
{
  ReadStream::ReadStream(const uint8_t* buffer, int bytes): BaseStream(), m_reader(buffer, bytes) {}

  bool ReadStream::SerializeInteger(int32_t& value, int32_t min, int32_t max)
  {
    assert(min < max);
    const int bits = encoding::bits_required(min, max);
    if (m_reader.WouldReadPastEnd(bits))
      return false;
    uint32_t unsigned_value = m_reader.ReadBits(bits);
    value = (int32_t)unsigned_value + min;
    return true;
  }

  bool ReadStream::SerializeBits(uint32_t& value, int bits)
  {
    assert(bits > 0);
    assert(bits <= 32);
    if (m_reader.WouldReadPastEnd(bits))
      return false;
    uint32_t read_value = m_reader.ReadBits(bits);
    value = read_value;
    return true;
  }

  bool ReadStream::SerializeBytes(uint8_t* data, int bytes)
  {
    if (!SerializeAlign())
      return false;
    if (m_reader.WouldReadPastEnd(bytes * 8))
      return false;
    m_reader.ReadBytes(data, bytes);
    return true;
  }

  bool ReadStream::SerializeAlign()
  {
    const int alignBits = m_reader.GetAlignBits();
    if (m_reader.WouldReadPastEnd(alignBits))
      return false;
    if (!m_reader.ReadAlign())
      return false;
    return true;
  }

  int ReadStream::GetAlignBits() const
  {
    return m_reader.GetAlignBits();
  }

  int ReadStream::GetBitsProcessed() const
  {
    return m_reader.GetBitsRead();
  }

  int ReadStream::GetBytesProcessed() const
  {
    return (m_reader.GetBitsRead() + 7) / 8;
  }
}  // namespace encoding
