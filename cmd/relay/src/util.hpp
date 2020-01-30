#ifndef UTIL_HPP
#define UTIL_HPP

#include <cstdarg>
#include <cstdio>

// TODO replace this with macro based loggging and then delete this file
extern int relay_debug;
inline void relay_printf(const char* format, ...)
{
    if (relay_debug)
        return;
    va_list args;
    va_start(args, format);
    char buffer[1024];
    vsnprintf(buffer, sizeof(buffer), format, args);
    printf("%s\n", buffer);
    va_end(args);
}
#endif