
#ifdef HAVE_AVX_ASM

#define IN_SANDY2X

#include "sandy2x_consts.S"
#include "sandy2x_fe51_mul.S"
#include "sandy2x_fe51_nsquare.S"
#include "sandy2x_fe51_pack.S"
#include "sandy2x_ladder.S"
#include "sandy2x_ladder_base.S"

#if defined(__linux__) && defined(__ELF__)
.section .note.GNU-stack,"",%progbits
#endif

#endif
