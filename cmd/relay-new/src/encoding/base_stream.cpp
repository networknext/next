#include "includes.h"
#include "base_stream.hpp"

namespace encoding
{
  BaseStream::BaseStream(): m_context(nullptr) {}

  void BaseStream::SetContext(void* context)
  {
    m_context = context;
  }

  void* BaseStream::GetContext() const
  {
    return m_context;
  }
}  // namespace encoding
