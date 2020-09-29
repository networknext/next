/*
    Network Next SDK. Copyright © 2017 - 2020 Network Next, Inc.

    Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following 
    conditions are met:

    1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

    2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions 
       and the following disclaimer in the documentation and/or other materials provided with the distribution.

    3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote 
       products derived from this software without specific prior written permission.

    THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, 
    INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. 
    IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR 
    CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; 
    OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING 
    NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

#include "next.h"

#ifndef NEXT_CRYPTO_H
#define NEXT_CRYPTO_H

#define NEXT_CRYPTO_GENERICHASH_KEYBYTES                    32

#define NEXT_CRYPTO_SECRETBOX_KEYBYTES                      32
#define NEXT_CRYPTO_SECRETBOX_MACBYTES                      16
#define NEXT_CRYPTO_SECRETBOX_NONCEBYTES                    24

#define NEXT_CRYPTO_KX_PUBLICKEYBYTES                       32
#define NEXT_CRYPTO_KX_SECRETKEYBYTES                       32
#define NEXT_CRYPTO_KX_SESSIONKEYBYTES                      32

#define NEXT_CRYPTO_BOX_MACBYTES                            16
#define NEXT_CRYPTO_BOX_NONCEBYTES                          24
#define NEXT_CRYPTO_BOX_PUBLICKEYBYTES                      32
#define NEXT_CRYPTO_BOX_SECRETKEYBYTES                      32

#define NEXT_CRYPTO_SIGN_BYTES                              64
#define NEXT_CRYPTO_SIGN_PUBLICKEYBYTES                     32
#define NEXT_CRYPTO_SIGN_SECRETKEYBYTES                     64

#define NEXT_CRYPTO_AEAD_CHACHA20POLY1305_ABYTES            16
#define NEXT_CRYPTO_AEAD_CHACHA20POLY1305_KEYBYTES          32
#define NEXT_CRYPTO_AEAD_CHACHA20POLY1305_NPUBBYTES          8

#define NEXT_CRYPTO_AEAD_CHACHA20POLY1305_IETF_ABYTES       16
#define NEXT_CRYPTO_AEAD_CHACHA20POLY1305_IETF_KEYBYTES     32
#define NEXT_CRYPTO_AEAD_CHACHA20POLY1305_IETF_NPUBBYTES    12

#endif // #ifndef NEXT_CRYPTO_H
