#pragma once

namespace util
{
  INLINE void DumpHex(void const* const data, size_t size, std::ostream& out = std::cout)
  {
    char ascii[17];
    size_t i, j;
    ascii[16] = '\0';
    for (i = 0; i < size; ++i) {
      // printf("%02X ", ((unsigned char*)data)[i]);
      out << std::setw(2) << std::setfill('0') << std::hex << (size_t)((unsigned char*)data)[i] << " " << std::dec;
      if (((unsigned char*)data)[i] >= ' ' && ((unsigned char*)data)[i] <= '~') {
        ascii[i % 16] = ((unsigned char*)data)[i];
      } else {
        ascii[i % 16] = '.';
      }
      if ((i + 1) % 8 == 0 || i + 1 == size) {
        out << " ";
        if ((i + 1) % 16 == 0) {
          out << "|  " << ascii << " \n";
        } else if (i + 1 == size) {
          ascii[(i + 1) % 16] = '\0';
          if ((i + 1) % 16 <= 8) {
            out << " ";
          }
          for (j = (i + 1) % 16; j < 16; ++j) {
            out << "   ";
          }
          out << "|  " << ascii << " \n";
        }
      }
    }
  }

  template <typename T>
  INLINE void DumpHex(const T& buff, std::ostream& out = std::cout)
  {
    DumpHex(buff.data(), buff.size(), out);
  }
}  // namespace util
