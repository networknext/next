#include "includes.h"
#include "write_stream.hpp"

#include "encoding/binary.hpp"

namespace encoding
{
  WriteStream::WriteStream(uint8_t* buffer, int bytes): m_writer(buffer, bytes) {}

  bool WriteStream::SerializeInteger(int32_t value, int32_t min, int32_t max)
  {
    assert(min < max);
    assert(value >= min);
    assert(value <= max);
    const int bits = encoding::bits_required(min, max);
    uint32_t unsigned_value = value - min;
    m_writer.WriteBits(unsigned_value, bits);
    return true;
  }

  bool WriteStream::SerializeBits(uint32_t value, int bits)
  {
    assert(bits > 0);
    assert(bits <= 32);
    m_writer.WriteBits(value, bits);
    return true;
  }

  bool WriteStream::SerializeBytes(const uint8_t* data, int bytes)
  {
    assert(data);
    assert(bytes >= 0);
    SerializeAlign();
    m_writer.WriteBytes(data, bytes);
    return true;
  }

  bool WriteStream::SerializeAlign()
  {
    m_writer.WriteAlign();
    return true;
  }

  int WriteStream::GetAlignBits() const
  {
    return m_writer.GetAlignBits();
  }

  void WriteStream::Flush()
  {
    m_writer.FlushBits();
  }

  const uint8_t* WriteStream::GetData() const
  {
    return m_writer.GetData();
  }

  int WriteStream::GetBytesProcessed() const
  {
    return m_writer.GetBytesWritten();
  }

  int WriteStream::GetBitsProcessed() const
  {
    return m_writer.GetBitsWritten();
  }
}  // namespace encoding
