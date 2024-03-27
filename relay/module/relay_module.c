/*
    Network Next Relay Linux kernel module

    This module supports Linux Kernel 6.5.0 *ONLY*

    USAGE:

        sudo insmod relay_module.ko
        lsmod
        sudo dmesg --follow
        sudo rmmod relay_module

    BTF debugging:

        sudo bpftool btf show
        sudo bpftool btf dump id <id>
*/

#include <linux/init.h>
#include <linux/module.h>
#include <linux/bpf.h>
#include <linux/btf.h>
#include <crypto/hash.h>
#include <crypto/kpp.h>
#include <crypto/poly1305.h>
#include <crypto/chacha.h>
#include <crypto/utils.h>
#include <linux/scatterlist.h>

MODULE_VERSION( "1.0.0" );
MODULE_LICENSE( "GPL" ); 
MODULE_AUTHOR( "Glenn Fiedler" ); 
MODULE_DESCRIPTION( "Network Next relay kernel module" );

#define XCHACHA20POLY1305_NONCE_SIZE 24

#define CHACHA20POLY1305_KEY_SIZE 32

struct chacha20poly1305_crypto
{
    __u8 nonce[XCHACHA20POLY1305_NONCE_SIZE];
    __u8 key[CHACHA20POLY1305_KEY_SIZE];
};

__bpf_kfunc int bpf_relay_sha256( void * data, int data__sz, void * output, int output__sz );

__bpf_kfunc int bpf_relay_xchacha20poly1305_decrypt( void * data, int data__sz, struct chacha20poly1305_crypto * crypto );

// ----------------------------------------------------------------------------------------------------------------------

#define CHACHA_KEY_WORDS ( CHACHA_KEY_SIZE / sizeof(u32) )

static bool __chacha20poly1305_decrypt( u8 * dst, const u8 * src, const size_t src_len, const u8 * ad, const size_t ad_len, u32 * chacha_state )
{
    const u8 *pad0 = page_address(ZERO_PAGE(0));
    struct poly1305_desc_ctx poly1305_state;
    size_t dst_len;
    int ret;
    union {
        u8 block0[POLY1305_KEY_SIZE];
        u8 mac[POLY1305_DIGEST_SIZE];
        __le64 lens[2];
    } b;

    if (unlikely(src_len < POLY1305_DIGEST_SIZE))
        return false;

    chacha20_crypt(chacha_state, b.block0, pad0, sizeof(b.block0));
    poly1305_init(&poly1305_state, b.block0);

    poly1305_update(&poly1305_state, ad, ad_len);
    if (ad_len & 0xf)
        poly1305_update(&poly1305_state, pad0, 0x10 - (ad_len & 0xf));

    dst_len = src_len - POLY1305_DIGEST_SIZE;
    poly1305_update(&poly1305_state, src, dst_len);
    if (dst_len & 0xf)
        poly1305_update(&poly1305_state, pad0, 0x10 - (dst_len & 0xf));

    b.lens[0] = cpu_to_le64(ad_len);
    b.lens[1] = cpu_to_le64(dst_len);
    poly1305_update(&poly1305_state, (u8 *)b.lens, sizeof(b.lens));

    poly1305_final(&poly1305_state, b.mac);

    ret = crypto_memneq(b.mac, src + dst_len, POLY1305_DIGEST_SIZE);
    if (likely(!ret))
        chacha20_crypt(chacha_state, dst, src, dst_len);

    memzero_explicit(&b, sizeof(b));

    return !ret;
}

static void chacha_load_key(u32 *k, const u8 *in)
{
    k[0] = get_unaligned_le32(in);
    k[1] = get_unaligned_le32(in + 4);
    k[2] = get_unaligned_le32(in + 8);
    k[3] = get_unaligned_le32(in + 12);
    k[4] = get_unaligned_le32(in + 16);
    k[5] = get_unaligned_le32(in + 20);
    k[6] = get_unaligned_le32(in + 24);
    k[7] = get_unaligned_le32(in + 28);
}

static void xchacha_init( u32 * chacha_state, const u8 * key, const u8 * nonce )
{
    u32 k[CHACHA_KEY_WORDS];
    u8 iv[CHACHA_IV_SIZE];

    memset(iv, 0, 8);
    memcpy(iv + 8, nonce + 16, 8);

    chacha_load_key(k, key);

    chacha_init(chacha_state, k, nonce);
    hchacha_block(chacha_state, k, 20);

    chacha_init(chacha_state, k, iv);

    memzero_explicit(k, sizeof(k));
    memzero_explicit(iv, sizeof(iv));
}

static bool xchacha20poly1305_decrypt( u8 * dst, const u8 * src, const size_t src_len,
                                       const u8 * ad, const size_t ad_len,
                                       const u8 nonce[XCHACHA20POLY1305_NONCE_SIZE],
                                       const u8 key[CHACHA20POLY1305_KEY_SIZE] )
{
    u32 chacha_state[CHACHA_STATE_WORDS];
    xchacha_init( chacha_state, key, nonce );
    return __chacha20poly1305_decrypt( dst, src, src_len, ad, ad_len, chacha_state );
}

// ----------------------------------------------------------------------------------------------------------------------

struct crypto_shash * sha256;

static int sha256_hash( const __u8 * data, __u32 data_len, __u8 * out_digest )
{
    SHASH_DESC_ON_STACK( shash, tfm );
    shash->tfm = sha256;
    crypto_shash_digest( shash, data, data_len, out_digest );
    return 0;
}

// ----------------------------------------------------------------------------------------------------------------------

__bpf_kfunc int bpf_relay_sha256( void * data, int data__sz, void * output, int output__sz )
{
    sha256_hash( data, data__sz, output );
    return 0;
}

__bpf_kfunc int bpf_relay_xchacha20poly1305_decrypt( void * data, int data__sz, struct chacha20poly1305_crypto * crypto )
{
    return xchacha20poly1305_decrypt( data, data, data__sz, NULL, 0, crypto->nonce, crypto->key ) == true;
}

BTF_SET8_START( bpf_task_set )
BTF_ID_FLAGS( func, bpf_relay_sha256 )
BTF_ID_FLAGS( func, bpf_relay_xchacha20poly1305_decrypt )
BTF_SET8_END( bpf_task_set )

static const struct btf_kfunc_id_set bpf_task_kfunc_set = {
    .owner = THIS_MODULE,
    .set   = &bpf_task_set,
};

// ----------------------------------------------------------------------------------------------------------------------

static int __init relay_init( void ) 
{
    pr_info( "Network Next relay module initializing...\n" );

    sha256 = crypto_alloc_shash( "sha256", 0, 0 );
    if ( IS_ERR( sha256 ) )
    {
        pr_err( "can't create sha256 crypto hash algorithm\n" );
        return PTR_ERR( sha256 );
    }

    __u8 digest[32];
    sha256_hash( "test", 4, digest );
    if ( digest[0]  != 0x9f || 
         digest[1]  != 0x86 ||
         digest[2]  != 0xd0 ||
         digest[3]  != 0x81 ||
         digest[4]  != 0x88 ||
         digest[5]  != 0x4c ||
         digest[6]  != 0x7d ||
         digest[7]  != 0x65 ||
         digest[8]  != 0x9a ||
         digest[9]  != 0x2f ||
         digest[10] != 0xea ||
         digest[11] != 0xa0 ||
         digest[12] != 0xc5 ||
         digest[13] != 0x5a ||
         digest[14] != 0xd0 ||
         digest[15] != 0x15 ||
         digest[16] != 0xa3 ||
         digest[17] != 0xbf ||
         digest[18] != 0x4f ||
         digest[19] != 0x1b ||
         digest[20] != 0x2b ||
         digest[21] != 0x0b ||
         digest[22] != 0x82 ||
         digest[23] != 0x2c ||
         digest[24] != 0xd1 ||
         digest[25] != 0x5d ||
         digest[26] != 0x6c ||
         digest[27] != 0x15 ||
         digest[28] != 0xb0 ||
         digest[29] != 0xf0 ||
         digest[30] != 0x0a ||
         digest[31] != 0x08 )
    {
        pr_err( "sha256 is broken\n" );
        return -1;
    }

    int result = register_btf_kfunc_id_set( BPF_PROG_TYPE_XDP, &bpf_task_kfunc_set );
    if ( result != 0 )
    {
        pr_err( "failed to register relay module kfuncs\n" );
        return -1;
    }

    pr_info( "Network Next relay module initialized successfully\n" );

    return result;
}

static void __exit relay_exit( void ) 
{
    pr_info( "Network Next relay module shutting down...\n" );

    if ( !IS_ERR( sha256 ) )
    {
        crypto_free_shash( sha256 );
    }

    pr_info( "Network Next relay module shut down successfully\n" );
}

module_init( relay_init );
module_exit( relay_exit );
