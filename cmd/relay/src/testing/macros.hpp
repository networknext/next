#ifndef TESTING_MACROS_HPP
#define TESTING_MACROS_HPP
#include <cstdio>
#include <cstdlib>

#define RUN_TEST(test_function)             \
    do {                                    \
        printf("    " #test_function "\n"); \
        fflush(stdout);                     \
        test_function();                    \
    } while (0)

static void check_handler(const char* condition, const char* function, const char* file, int line)
{
    printf("check failed: ( %s ), function %s, file %s, line %d\n", condition, function, file, line);
    fflush(stdout);
#ifndef NDEBUG
#if defined(__GNUC__)
    __builtin_trap();
#elif defined(_MSC_VER)
    __debugbreak();
#endif
#endif
    exit(1);
}

#define check(condition)                                                                           \
    do {                                                                                           \
        if (!(condition)) {                                                                        \
            check_handler(#condition, (const char*)__FUNCTION__, (const char*)__FILE__, __LINE__); \
        }                                                                                          \
    } while (0)
#endif