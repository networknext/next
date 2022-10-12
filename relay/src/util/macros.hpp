#pragma once

// force any function with this to always be inlined
#define INLINE [[gnu::always_inline]] inline

// this will prevent GCC from optimizing out useless function calls, for benchmarking
#ifdef BENCH_BUILD
#define GCC_NO_OPT_OUT asm("")
#else
#define GCC_NO_OPT_OUT
#endif
