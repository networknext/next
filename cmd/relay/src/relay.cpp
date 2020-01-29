/*
 * Network Next Relay.
 * Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
 */

#include "relay.h"

#include <cassert>
#include <string.h>
#include <stdio.h>
#include <cinttypes>
#include <stdarg.h>
#include <sodium.h>
#include <math.h>
#include <map>
#include <float.h>
#include <signal.h>
#include <curl/curl.h>

#include "config.hpp"
#include "relay_test.hpp"
#include "encoding/base64.hpp"
#include "relay/relay_replay_protection.hpp"
#include "relay/relay_ping_history.hpp"
#include "relay/relay_address.hpp"
#include "sysinfo.hpp"
#include "relay/relay_platform.hpp"


// -----------------------------------------------------------------------------

static int relay_debug = 0;

void relay_printf(const char* format, ...)
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

// -----------------------------------------------------------------------------

int relay_initialize()
{
    if (relay::relay_platform_init() != RELAY_OK) {
        relay_printf("failed to initialize platform");
        return RELAY_ERROR;
    }

    if (sodium_init() == -1) {
        relay_printf("failed to initialize sodium");
        return RELAY_ERROR;
    }

    const char* relay_debug_env = relay::relay_platform_getenv("RELAY_DEBUG");
    if (relay_debug_env) {
        relay_debug = atoi(relay_debug_env);
    }

    return RELAY_OK;
}

void relay_term()
{
    relay::relay_platform_term();
}

// -----------------------------------------------------------------------------

void relay_random_bytes(uint8_t* buffer, int bytes)
{
    randombytes_buf(buffer, bytes);
}

uint16_t relay_ntohs(uint16_t in)
{
    return (uint16_t)(((in << 8) & 0xFF00) | ((in >> 8) & 0x00FF));
}

uint16_t relay_htons(uint16_t in)
{
    return (uint16_t)(((in << 8) & 0xFF00) | ((in >> 8) & 0x00FF));
}

namespace relay
{
    /**
        Stream class for reading bitpacked data.
        This class is a wrapper around the bit reader class. Its purpose is to provide unified interface for reading and
       writing. You can determine if you are reading from a stream by calling Stream::IsReading inside your templated serialize
       method. This is evaluated at compile time, letting the compiler generate optimized serialize functions without the hassle
       of maintaining separate read and write functions. IMPORTANT: Generally, you don't call methods on this class directly.
       Use the serialize_* macros instead. See test/shared.h for some examples.
     */

    /**
        Serialize integer value (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param value The integer value to serialize in [min,max].
        @param min The minimum value.
        @param max The maximum value.
     */

#define serialize_int(stream, value, min, max)                                    \
    do {                                                                          \
        assert(min < max);                                                        \
        int32_t int32_value = 0;                                                  \
        if (Stream::IsWriting) {                                                  \
            assert(int64_t(value) >= int64_t(min));                               \
            assert(int64_t(value) <= int64_t(max));                               \
            int32_value = (int32_t)value;                                         \
        }                                                                         \
        if (!stream.SerializeInteger(int32_value, min, max)) {                    \
            return false;                                                         \
        }                                                                         \
        if (Stream::IsReading) {                                                  \
            value = int32_value;                                                  \
            if (int64_t(value) < int64_t(min) || int64_t(value) > int64_t(max)) { \
                return false;                                                     \
            }                                                                     \
        }                                                                         \
    } while (0)

    /**
        Serialize bits to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param value The unsigned integer value to serialize.
        @param bits The number of bits to serialize in [1,32].
     */

#define serialize_bits(stream, value, bits)              \
    do {                                                 \
        assert(bits > 0);                                \
        assert(bits <= 32);                              \
        uint32_t uint32_value = 0;                       \
        if (Stream::IsWriting) {                         \
            uint32_value = (uint32_t)value;              \
        }                                                \
        if (!stream.SerializeBits(uint32_value, bits)) { \
            return false;                                \
        }                                                \
        if (Stream::IsReading) {                         \
            value = uint32_value;                        \
        }                                                \
    } while (0)

    /**
        Serialize a boolean value to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param value The boolean value to serialize.
     */

#define serialize_bool(stream, value)                 \
    do {                                              \
        uint32_t uint32_bool_value = 0;               \
        if (Stream::IsWriting) {                      \
            uint32_bool_value = value ? 1 : 0;        \
        }                                             \
        serialize_bits(stream, uint32_bool_value, 1); \
        if (Stream::IsReading) {                      \
            value = uint32_bool_value ? true : false; \
        }                                             \
    } while (0)

    template <typename Stream>
    bool serialize_float_internal(Stream& stream, float& value)
    {
        uint32_t int_value;
        if (Stream::IsWriting) {
            memcpy(&int_value, &value, 4);
        }
        bool result = stream.SerializeBits(int_value, 32);
        if (Stream::IsReading && result) {
            memcpy(&value, &int_value, 4);
        }
        return result;
    }

    /**
        Serialize floating point value (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param value The float value to serialize.
     */

#define serialize_float(stream, value)                         \
    do {                                                       \
        if (!relay::serialize_float_internal(stream, value)) { \
            return false;                                      \
        }                                                      \
    } while (0)

    /**
        Serialize a 32 bit unsigned integer to the stream (read/write).
        This is a helper macro to make unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param value The unsigned 32 bit integer value to serialize.
     */

#define serialize_uint32(stream, value) serialize_bits(stream, value, 32);

    template <typename Stream>
    bool serialize_uint64_internal(Stream& stream, uint64_t& value)
    {
        uint32_t hi = 0, lo = 0;
        if (Stream::IsWriting) {
            lo = value & 0xFFFFFFFF;
            hi = value >> 32;
        }
        serialize_bits(stream, lo, 32);
        serialize_bits(stream, hi, 32);
        if (Stream::IsReading) {
            value = (uint64_t(hi) << 32) | lo;
        }
        return true;
    }

    /**
        Serialize a 64 bit unsigned integer to the stream (read/write).
        This is a helper macro to make unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param value The unsigned 64 bit integer value to serialize.
     */

#define serialize_uint64(stream, value)                       \
    do {                                                      \
        if (!relay::serialize_uint64_internal(stream, value)) \
            return false;                                     \
    } while (0)

    template <typename Stream>
    bool serialize_double_internal(Stream& stream, double& value)
    {
        union DoubleInt
        {
            double double_value;
            uint64_t int_value;
        };
        DoubleInt tmp = { 0 };
        if (Stream::IsWriting) {
            tmp.double_value = value;
        }
        serialize_uint64(stream, tmp.int_value);
        if (Stream::IsReading) {
            value = tmp.double_value;
        }
        return true;
    }

    /**
        Serialize double precision floating point value to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param value The double precision floating point value to serialize.
     */

#define serialize_double(stream, value)                         \
    do {                                                        \
        if (!relay::serialize_double_internal(stream, value)) { \
            return false;                                       \
        }                                                       \
    } while (0)

    template <typename Stream>
    bool serialize_bytes_internal(Stream& stream, uint8_t* data, int bytes)
    {
        return stream.SerializeBytes(data, bytes);
    }

    /**
        Serialize an array of bytes to the stream (read/write).
        This is a helper macro to make unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param data Pointer to the data to be serialized.
        @param bytes The number of bytes to serialize.
     */

#define serialize_bytes(stream, data, bytes)                         \
    do {                                                             \
        if (!relay::serialize_bytes_internal(stream, data, bytes)) { \
            return false;                                            \
        }                                                            \
    } while (0)

    template <typename Stream>
    bool serialize_string_internal(Stream& stream, char* string, int buffer_size)
    {
        int length = 0;
        if (Stream::IsWriting) {
            length = (int)strlen(string);
            assert(length < buffer_size);
        }
        serialize_int(stream, length, 0, buffer_size - 1);
        serialize_bytes(stream, (uint8_t*)string, length);
        if (Stream::IsReading) {
            string[length] = '\0';
        }
        return true;
    }

    /**
        Serialize a string to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param string The string to serialize write. Pointer to buffer to be filled on read.
        @param buffer_size The size of the string buffer. String with terminating null character must fit into this buffer.
     */

#define serialize_string(stream, string, buffer_size)                         \
    do {                                                                      \
        if (!relay::serialize_string_internal(stream, string, buffer_size)) { \
            return false;                                                     \
        }                                                                     \
    } while (0)

    /**
        Serialize an alignment to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
     */

#define serialize_align(stream)         \
    do {                                \
        if (!stream.SerializeAlign()) { \
            return false;               \
        }                               \
    } while (0)

    /**
        Serialize an object to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param object The object to serialize. Must have a serialize method on it.
     */

#define serialize_object(stream, object) \
    do {                                 \
        if (!object.Serialize(stream)) { \
            return false;                \
        }                                \
    } while (0)

    template <typename Stream, typename T>
    bool serialize_int_relative_internal(Stream& stream, T previous, T& current)
    {
        uint32_t difference = 0;
        if (Stream::IsWriting) {
            assert(previous < current);
            difference = current - previous;
        }

        bool oneBit = false;
        if (Stream::IsWriting) {
            oneBit = difference == 1;
        }
        serialize_bool(stream, oneBit);
        if (oneBit) {
            if (Stream::IsReading) {
                current = previous + 1;
            }
            return true;
        }

        bool twoBits = false;
        if (Stream::IsWriting) {
            twoBits = difference <= 6;
        }
        serialize_bool(stream, twoBits);
        if (twoBits) {
            serialize_int(stream, difference, 2, 6);
            if (Stream::IsReading) {
                current = previous + difference;
            }
            return true;
        }

        bool fourBits = false;
        if (Stream::IsWriting) {
            fourBits = difference <= 23;
        }
        serialize_bool(stream, fourBits);
        if (fourBits) {
            serialize_int(stream, difference, 7, 23);
            if (Stream::IsReading) {
                current = previous + difference;
            }
            return true;
        }

        bool eightBits = false;
        if (Stream::IsWriting) {
            eightBits = difference <= 280;
        }
        serialize_bool(stream, eightBits);
        if (eightBits) {
            serialize_int(stream, difference, 24, 280);
            if (Stream::IsReading) {
                current = previous + difference;
            }
            return true;
        }

        bool twelveBits = false;
        if (Stream::IsWriting) {
            twelveBits = difference <= 4377;
        }
        serialize_bool(stream, twelveBits);
        if (twelveBits) {
            serialize_int(stream, difference, 281, 4377);
            if (Stream::IsReading) {
                current = previous + difference;
            }
            return true;
        }

        bool sixteenBits = false;
        if (Stream::IsWriting) {
            sixteenBits = difference <= 69914;
        }
        serialize_bool(stream, sixteenBits);
        if (sixteenBits) {
            serialize_int(stream, difference, 4378, 69914);
            if (Stream::IsReading) {
                current = previous + difference;
            }
            return true;
        }

        uint32_t value = current;
        serialize_uint32(stream, value);
        if (Stream::IsReading) {
            current = value;
        }

        return true;
    }

    /**
        Serialize an integer value relative to another (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param previous The previous integer value.
        @param current The current integer value.
     */

#define serialize_int_relative(stream, previous, current)                         \
    do {                                                                          \
        if (!relay::serialize_int_relative_internal(stream, previous, current)) { \
            return false;                                                         \
        }                                                                         \
    } while (0)

    template <typename Stream>
    bool serialize_ack_relative_internal(Stream& stream, uint16_t sequence, uint16_t& ack)
    {
        int ack_delta = 0;
        bool ack_in_range = false;
        if (Stream::IsWriting) {
            if (ack < sequence) {
                ack_delta = sequence - ack;
            } else {
                ack_delta = (int)sequence + 65536 - ack;
            }
            assert(ack_delta > 0);
            assert(uint16_t(sequence - ack_delta) == ack);
            ack_in_range = ack_delta <= 64;
        }
        serialize_bool(stream, ack_in_range);
        if (ack_in_range) {
            serialize_int(stream, ack_delta, 1, 64);
            if (Stream::IsReading) {
                ack = sequence - ack_delta;
            }
        } else {
            serialize_bits(stream, ack, 16);
        }
        return true;
    }

    /**
        Serialize an ack relative to the current sequence number (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param sequence The current sequence number.
        @param ack The ack sequence number, which is typically near the current sequence number.
     */

#define serialize_ack_relative(stream, sequence, ack)                         \
    do {                                                                      \
        if (!relay::serialize_ack_relative_internal(stream, sequence, ack)) { \
            return false;                                                     \
        }                                                                     \
    } while (0)

    template <typename Stream>
    bool serialize_sequence_relative_internal(Stream& stream, uint16_t sequence1, uint16_t& sequence2)
    {
        if (Stream::IsWriting) {
            uint32_t a = sequence1;
            uint32_t b = sequence2 + ((sequence1 > sequence2) ? 65536 : 0);
            serialize_int_relative(stream, a, b);
        } else {
            uint32_t a = sequence1;
            uint32_t b = 0;
            serialize_int_relative(stream, a, b);
            if (b >= 65536) {
                b -= 65536;
            }
            sequence2 = uint16_t(b);
        }

        return true;
    }

    /**
        Serialize a sequence number relative to another (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an
       important safety measure because packet data comes from the network and may be malicious. IMPORTANT: This macro must be
       called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool
       return value.
        @param stream The stream object. May be a read or write stream.
        @param sequence1 The first sequence number to serialize relative to.
        @param sequence2 The second sequence number to be encoded relative to the first.
     */

#define serialize_sequence_relative(stream, sequence1, sequence2)                         \
    do {                                                                                  \
        if (!relay::serialize_sequence_relative_internal(stream, sequence1, sequence2)) { \
            return false;                                                                 \
        }                                                                                 \
    } while (0)

    template <typename Stream>
    bool serialize_address_internal(Stream& stream, relay_address_t& address)
    {
        serialize_bits(stream, address.type, 2);
        if (address.type == RELAY_ADDRESS_IPV4) {
            serialize_bytes(stream, address.data.ipv4, 4);
            serialize_bits(stream, address.port, 16);
        } else if (address.type == RELAY_ADDRESS_IPV6) {
            for (int i = 0; i < 8; ++i) {
                serialize_bits(stream, address.data.ipv6[i], 16);
            }
            serialize_bits(stream, address.port, 16);
        } else {
            if (Stream::IsReading) {
                memset(&address, 0, sizeof(relay_address_t));
            }
        }
        return true;
    }

#define serialize_address(stream, address)                         \
    do {                                                           \
        if (!relay::serialize_address_internal(stream, address)) { \
            return false;                                          \
        }                                                          \
    } while (0)
}  // namespace relay

// --------------------------------------------------------------------------

int relay_wire_packet_bits(int packet_bytes)
{
    return (14 + 20 + 8 + packet_bytes + 4) * 8;
}

struct relay_bandwidth_limiter_t
{
    uint64_t bits_sent;
    double last_check_time;
    double average_kbps;
};

void relay_bandwidth_limiter_reset(relay_bandwidth_limiter_t* bandwidth_limiter)
{
    assert(bandwidth_limiter);
    bandwidth_limiter->last_check_time = -100.0;
    bandwidth_limiter->bits_sent = 0;
    bandwidth_limiter->average_kbps = 0.0;
}

bool relay_bandwidth_limiter_add_packet(
    relay_bandwidth_limiter_t* bandwidth_limiter, double current_time, uint32_t kbps_allowed, uint32_t packet_bits)
{
    assert(bandwidth_limiter);
    const bool invalid = bandwidth_limiter->last_check_time < 0.0;
    if (invalid || current_time - bandwidth_limiter->last_check_time >= RELAY_BANDWIDTH_LIMITER_INTERVAL - 0.001f) {
        bandwidth_limiter->bits_sent = 0;
        bandwidth_limiter->last_check_time = current_time;
    }
    bandwidth_limiter->bits_sent += packet_bits;
    return bandwidth_limiter->bits_sent > (uint64_t)(kbps_allowed * 1000 * RELAY_BANDWIDTH_LIMITER_INTERVAL);
}

void relay_bandwidth_limiter_add_sample(relay_bandwidth_limiter_t* bandwidth_limiter, double kbps)
{
    if (bandwidth_limiter->average_kbps == 0.0 && kbps != 0.0) {
        bandwidth_limiter->average_kbps = kbps;
        return;
    }

    if (bandwidth_limiter->average_kbps != 0.0 && kbps == 0.0) {
        bandwidth_limiter->average_kbps = 0.0;
        return;
    }

    const double delta = kbps - bandwidth_limiter->average_kbps;

    if (delta < 0.000001f) {
        bandwidth_limiter->average_kbps = kbps;
        return;
    }

    bandwidth_limiter->average_kbps += delta * 0.1f;
}

double relay_bandwidth_limiter_usage_kbps(relay_bandwidth_limiter_t* bandwidth_limiter, double current_time)
{
    assert(bandwidth_limiter);
    const bool invalid = bandwidth_limiter->last_check_time < 0.0;
    if (!invalid) {
        const double delta_time = current_time - bandwidth_limiter->last_check_time;
        if (delta_time > 0.1f) {
            const double kbps = bandwidth_limiter->bits_sent / delta_time / 1000.0;
            relay_bandwidth_limiter_add_sample(bandwidth_limiter, kbps);
        }
    }
    return bandwidth_limiter->average_kbps;
}

// --------------------------------------------------------------------------

struct relay_route_token_t
{
    uint64_t expire_timestamp;
    uint64_t session_id;
    uint8_t session_version;
    uint8_t session_flags;
    int kbps_up;
    int kbps_down;
    relay::relay_address_t next_address;
    uint8_t private_key[crypto_box_SECRETKEYBYTES];
};

void relay_write_route_token(relay_route_token_t* token, uint8_t* buffer, int buffer_length)
{
    (void)buffer_length;

    assert(token);
    assert(buffer);
    assert(buffer_length >= RELAY_ROUTE_TOKEN_BYTES);

    uint8_t* start = buffer;

    (void)start;

    relay_write_uint64(&buffer, token->expire_timestamp);
    relay_write_uint64(&buffer, token->session_id);
    relay_write_uint8(&buffer, token->session_version);
    relay_write_uint8(&buffer, token->session_flags);
    relay_write_uint32(&buffer, token->kbps_up);
    relay_write_uint32(&buffer, token->kbps_down);
    relay_write_address(&buffer, &token->next_address);
    relay_write_bytes(&buffer, token->private_key, crypto_box_SECRETKEYBYTES);

    assert(buffer - start == RELAY_ROUTE_TOKEN_BYTES);
}

void relay_read_route_token(relay_route_token_t* token, const uint8_t* buffer)
{
    assert(token);
    assert(buffer);

    const uint8_t* start = buffer;

    (void)start;

    token->expire_timestamp = relay_read_uint64(&buffer);
    token->session_id = relay_read_uint64(&buffer);
    token->session_version = relay_read_uint8(&buffer);
    token->session_flags = relay_read_uint8(&buffer);
    token->kbps_up = relay_read_uint32(&buffer);
    token->kbps_down = relay_read_uint32(&buffer);
    relay_read_address(&buffer, &token->next_address);
    relay_read_bytes(&buffer, token->private_key, crypto_box_SECRETKEYBYTES);
    assert(buffer - start == RELAY_ROUTE_TOKEN_BYTES);
}

int relay_encrypt_route_token(
    uint8_t* sender_private_key, uint8_t* receiver_public_key, uint8_t* nonce, uint8_t* buffer, int buffer_length)
{
    assert(sender_private_key);
    assert(receiver_public_key);
    assert(buffer);
    assert(buffer_length >= (int)(RELAY_ROUTE_TOKEN_BYTES + crypto_box_MACBYTES));

    (void)buffer_length;

    if (crypto_box_easy(buffer, buffer, RELAY_ROUTE_TOKEN_BYTES, nonce, receiver_public_key, sender_private_key) != 0) {
        return RELAY_ERROR;
    }

    return RELAY_OK;
}

int relay_decrypt_route_token(
    const uint8_t* sender_public_key, const uint8_t* receiver_private_key, const uint8_t* nonce, uint8_t* buffer)
{
    assert(sender_public_key);
    assert(receiver_private_key);
    assert(buffer);

    if (crypto_box_open_easy(
            buffer, buffer, RELAY_ROUTE_TOKEN_BYTES + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key) !=
        0) {
        return RELAY_ERROR;
    }

    return RELAY_OK;
}

int relay_write_encrypted_route_token(
    uint8_t** buffer, relay_route_token_t* token, uint8_t* sender_private_key, uint8_t* receiver_public_key)
{
    assert(buffer);
    assert(token);
    assert(sender_private_key);
    assert(receiver_public_key);

    unsigned char nonce[crypto_box_NONCEBYTES];
    relay_random_bytes(nonce, crypto_box_NONCEBYTES);

    uint8_t* start = *buffer;

    (void)start;

    relay_write_bytes(buffer, nonce, crypto_box_NONCEBYTES);

    relay_write_route_token(token, *buffer, RELAY_ROUTE_TOKEN_BYTES);

    if (relay_encrypt_route_token(
            sender_private_key, receiver_public_key, nonce, *buffer, RELAY_ROUTE_TOKEN_BYTES + crypto_box_NONCEBYTES) !=
        RELAY_OK)
        return RELAY_ERROR;

    *buffer += RELAY_ROUTE_TOKEN_BYTES + crypto_box_MACBYTES;

    assert((*buffer - start) == RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES);

    return RELAY_OK;
}

int relay_read_encrypted_route_token(
    uint8_t** buffer, relay_route_token_t* token, const uint8_t* sender_public_key, const uint8_t* receiver_private_key)
{
    assert(buffer);
    assert(token);
    assert(sender_public_key);
    assert(receiver_private_key);

    const uint8_t* nonce = *buffer;

    *buffer += crypto_box_NONCEBYTES;

    if (relay_decrypt_route_token(sender_public_key, receiver_private_key, nonce, *buffer) != RELAY_OK) {
        return RELAY_ERROR;
    }

    relay_read_route_token(token, *buffer);

    *buffer += RELAY_ROUTE_TOKEN_BYTES + crypto_box_MACBYTES;

    return RELAY_OK;
}

// --------------------------------------------------------------------------

struct relay_continue_token_t
{
    uint64_t expire_timestamp;
    uint64_t session_id;
    uint8_t session_version;
    uint8_t session_flags;
};

void relay_write_continue_token(relay_continue_token_t* token, uint8_t* buffer, int buffer_length)
{
    (void)buffer_length;

    assert(token);
    assert(buffer);
    assert(buffer_length >= RELAY_CONTINUE_TOKEN_BYTES);

    uint8_t* start = buffer;

    (void)start;

    relay_write_uint64(&buffer, token->expire_timestamp);
    relay_write_uint64(&buffer, token->session_id);
    relay_write_uint8(&buffer, token->session_version);
    relay_write_uint8(&buffer, token->session_flags);

    assert(buffer - start == RELAY_CONTINUE_TOKEN_BYTES);
}

void relay_read_continue_token(relay_continue_token_t* token, const uint8_t* buffer)
{
    assert(token);
    assert(buffer);

    const uint8_t* start = buffer;

    (void)start;

    token->expire_timestamp = relay_read_uint64(&buffer);
    token->session_id = relay_read_uint64(&buffer);
    token->session_version = relay_read_uint8(&buffer);
    token->session_flags = relay_read_uint8(&buffer);

    assert(buffer - start == RELAY_CONTINUE_TOKEN_BYTES);
}

int relay_encrypt_continue_token(
    uint8_t* sender_private_key, uint8_t* receiver_public_key, uint8_t* nonce, uint8_t* buffer, int buffer_length)
{
    assert(sender_private_key);
    assert(receiver_public_key);
    assert(buffer);
    assert(buffer_length >= (int)(RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES));

    (void)buffer_length;

    if (crypto_box_easy(buffer, buffer, RELAY_CONTINUE_TOKEN_BYTES, nonce, receiver_public_key, sender_private_key) != 0) {
        return RELAY_ERROR;
    }

    return RELAY_OK;
}

int relay_decrypt_continue_token(
    const uint8_t* sender_public_key, const uint8_t* receiver_private_key, const uint8_t* nonce, uint8_t* buffer)
{
    assert(sender_public_key);
    assert(receiver_private_key);
    assert(buffer);

    if (crypto_box_open_easy(
            buffer, buffer, RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key) !=
        0) {
        return RELAY_ERROR;
    }

    return RELAY_OK;
}

int relay_write_encrypted_continue_token(
    uint8_t** buffer, relay_continue_token_t* token, uint8_t* sender_private_key, uint8_t* receiver_public_key)
{
    assert(buffer);
    assert(token);
    assert(sender_private_key);
    assert(receiver_public_key);

    unsigned char nonce[crypto_box_NONCEBYTES];
    relay_random_bytes(nonce, crypto_box_NONCEBYTES);

    uint8_t* start = *buffer;

    relay_write_bytes(buffer, nonce, crypto_box_NONCEBYTES);

    relay_write_continue_token(token, *buffer, RELAY_CONTINUE_TOKEN_BYTES);

    if (relay_encrypt_continue_token(
            sender_private_key, receiver_public_key, nonce, *buffer, RELAY_CONTINUE_TOKEN_BYTES + crypto_box_NONCEBYTES) !=
        RELAY_OK)
        return RELAY_ERROR;

    *buffer += RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES;

    (void)start;

    assert((*buffer - start) == RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES);

    return RELAY_OK;
}

int relay_read_encrypted_continue_token(
    uint8_t** buffer, relay_continue_token_t* token, const uint8_t* sender_public_key, const uint8_t* receiver_private_key)
{
    assert(buffer);
    assert(token);
    assert(sender_public_key);
    assert(receiver_private_key);

    const uint8_t* nonce = *buffer;

    *buffer += crypto_box_NONCEBYTES;

    if (relay_decrypt_continue_token(sender_public_key, receiver_private_key, nonce, *buffer) != RELAY_OK) {
        return RELAY_ERROR;
    }

    relay_read_continue_token(token, *buffer);

    *buffer += RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES;

    return RELAY_OK;
}

// --------------------------------------------------------------------------

int relay_write_header(int direction,
    uint8_t type,
    uint64_t sequence,
    uint64_t session_id,
    uint8_t session_version,
    const uint8_t* private_key,
    uint8_t* buffer,
    int buffer_length)
{
    assert(private_key);
    assert(buffer);
    assert(RELAY_HEADER_BYTES <= buffer_length);

    (void)buffer_length;

    uint8_t* start = buffer;

    (void)start;

    if (direction == RELAY_DIRECTION_SERVER_TO_CLIENT) {
        // high bit must be set
        assert(sequence & (1ULL << 63));
    } else {
        // high bit must be clear
        assert((sequence & (1ULL << 63)) == 0);
    }

    if (type == RELAY_SESSION_PING_PACKET || type == RELAY_SESSION_PONG_PACKET || type == RELAY_ROUTE_RESPONSE_PACKET ||
        type == RELAY_CONTINUE_RESPONSE_PACKET) {
        // second highest bit must be set
        assert(sequence & (1ULL << 62));
    } else {
        // second highest bit must be clear
        assert((sequence & (1ULL << 62)) == 0);
    }

    relay_write_uint8(&buffer, type);

    relay_write_uint64(&buffer, sequence);

    uint8_t* additional = buffer;
    const int additional_length = 8 + 2;

    relay_write_uint64(&buffer, session_id);
    relay_write_uint8(&buffer, session_version);
    relay_write_uint8(&buffer, 0);  // todo: remove this once we fully switch to new relay

    uint8_t nonce[12];
    {
        uint8_t* p = nonce;
        relay_write_uint32(&p, 0);
        relay_write_uint64(&p, sequence);
    }

    unsigned long long encrypted_length = 0;

    int result = crypto_aead_chacha20poly1305_ietf_encrypt(
        buffer, &encrypted_length, buffer, 0, additional, (unsigned long long)additional_length, NULL, nonce, private_key);

    if (result != 0)
        return RELAY_ERROR;

    buffer += encrypted_length;

    assert(int(buffer - start) == RELAY_HEADER_BYTES);

    return RELAY_OK;
}

int relay_peek_header(int direction,
    uint8_t* type,
    uint64_t* sequence,
    uint64_t* session_id,
    uint8_t* session_version,
    const uint8_t* buffer,
    int buffer_length)
{
    uint8_t packet_type;
    uint64_t packet_sequence;

    assert(buffer);

    if (buffer_length < RELAY_HEADER_BYTES)
        return RELAY_ERROR;

    packet_type = relay_read_uint8(&buffer);

    packet_sequence = relay_read_uint64(&buffer);

    if (direction == RELAY_DIRECTION_SERVER_TO_CLIENT) {
        // high bit must be set
        if (!(packet_sequence & (1ULL << 63)))
            return RELAY_ERROR;
    } else {
        // high bit must be clear
        if (packet_sequence & (1ULL << 63))
            return RELAY_ERROR;
    }

    *type = packet_type;

    if (*type == RELAY_SESSION_PING_PACKET || *type == RELAY_SESSION_PONG_PACKET || *type == RELAY_ROUTE_RESPONSE_PACKET ||
        *type == RELAY_CONTINUE_RESPONSE_PACKET) {
        // second highest bit must be set
        assert(packet_sequence & (1ULL << 62));
    } else {
        // second highest bit must be clear
        assert((packet_sequence & (1ULL << 62)) == 0);
    }

    *sequence = packet_sequence;
    *session_id = relay_read_uint64(&buffer);
    *session_version = relay_read_uint8(&buffer);

    return RELAY_OK;
}

int relay_verify_header(int direction, const uint8_t* private_key, uint8_t* buffer, int buffer_length)
{
    assert(private_key);
    assert(buffer);

    if (buffer_length < RELAY_HEADER_BYTES) {
        return RELAY_ERROR;
    }

    const uint8_t* p = buffer;

    uint8_t packet_type = relay_read_uint8(&p);

    uint64_t packet_sequence = relay_read_uint64(&p);

    if (direction == RELAY_DIRECTION_SERVER_TO_CLIENT) {
        // high bit must be set
        if (!(packet_sequence & (1ULL << 63))) {
            return RELAY_ERROR;
        }
    } else {
        // high bit must be clear
        if (packet_sequence & (1ULL << 63)) {
            return RELAY_ERROR;
        }
    }

    if (packet_type == RELAY_SESSION_PING_PACKET || packet_type == RELAY_SESSION_PONG_PACKET ||
        packet_type == RELAY_ROUTE_RESPONSE_PACKET || packet_type == RELAY_CONTINUE_RESPONSE_PACKET) {
        // second highest bit must be set
        assert(packet_sequence & (1ULL << 62));
    } else {
        // second highest bit must be clear
        assert((packet_sequence & (1ULL << 62)) == 0);
    }

    const uint8_t* additional = p;

    const int additional_length = 8 + 2;

    uint64_t packet_session_id = relay_read_uint64(&p);
    uint8_t packet_session_version = relay_read_uint8(&p);
    uint8_t packet_session_flags = relay_read_uint8(&p);  // todo: remove once we fully switch over to new relay

    (void)packet_session_id;
    (void)packet_session_version;
    (void)packet_session_flags;

    uint8_t nonce[12];
    {
        uint8_t* q = nonce;
        relay_write_uint32(&q, 0);
        relay_write_uint64(&q, packet_sequence);
    }

    unsigned long long decrypted_length;

    int result = crypto_aead_chacha20poly1305_ietf_decrypt(buffer + 19,
        &decrypted_length,
        NULL,
        buffer + 19,
        (unsigned long long)crypto_aead_chacha20poly1305_IETF_ABYTES,
        additional,
        (unsigned long long)additional_length,
        nonce,
        private_key);

    if (result != 0) {
        return RELAY_ERROR;
    }

    return RELAY_OK;
}

// -------------------------------------------------------------

#define RELAY_TOKEN_BYTES 32
#define RESPONSE_MAX_BYTES 1024 * 1024

#define NEAR_PING_PACKET 73
#define NEAR_PONG_PACKET 74
#define RELAY_PING_PACKET 75
#define RELAY_PONG_PACKET 76

struct relay_session_t
{
    uint64_t expire_timestamp;
    uint64_t session_id;
    uint8_t session_version;
    uint64_t client_to_server_sequence;
    uint64_t server_to_client_sequence;
    int kbps_up;
    int kbps_down;
    relay::relay_address_t prev_address;
    relay::relay_address_t next_address;
    uint8_t private_key[crypto_box_SECRETKEYBYTES];
    relay::relay_replay_protection_t replay_protection_server_to_client;
    relay::relay_replay_protection_t replay_protection_client_to_server;
};

struct relay_t
{
    relay_manager_t* relay_manager;
    relay_platform_socket_t* socket;
    relay_platform_mutex_t* mutex;
    double initialize_time;
    uint64_t initialize_router_timestamp;
    uint8_t relay_public_key[RELAY_PUBLIC_KEY_BYTES];
    uint8_t relay_private_key[RELAY_PRIVATE_KEY_BYTES];
    uint8_t router_public_key[RELAY_PUBLIC_KEY_BYTES];
    std::map<uint64_t, relay_session_t*>* sessions;
    bool relays_dirty;
    int num_relays;
    uint64_t relay_ids[MAX_RELAYS];
    relay::relay_address_t relay_addresses[MAX_RELAYS];
};

struct curl_buffer_t
{
    int size;
    int max_size;
    uint8_t* data;
};

size_t curl_buffer_write_function(char* ptr, size_t size, size_t nmemb, void* userdata)
{
    curl_buffer_t* buffer = (curl_buffer_t*)userdata;
    assert(buffer);
    assert(size == 1);
    if (int(buffer->size + size * nmemb) > buffer->max_size)
        return 0;
    memcpy(buffer->data + buffer->size, ptr, size * nmemb);
    buffer->size += size * nmemb;
    return size * nmemb;
}

int relay_init(CURL* curl,
    const char* hostname,
    uint8_t* relay_token,
    const char* relay_address,
    const uint8_t* router_public_key,
    const uint8_t* relay_private_key,
    uint64_t* router_timestamp)
{
    const uint32_t init_request_magic = 0x9083708f;

    uint32_t init_request_version = 0;

    uint8_t init_data[1024];
    memset(init_data, 0, sizeof(init_data));

    unsigned char nonce[crypto_box_NONCEBYTES];
    relay_random_bytes(nonce, crypto_box_NONCEBYTES);

    uint8_t* p = init_data;

    relay_write_uint32(&p, init_request_magic);
    relay_write_uint32(&p, init_request_version);
    relay_write_bytes(&p, nonce, crypto_box_NONCEBYTES);
    relay_write_string(&p, relay_address, RELAY_MAX_ADDRESS_STRING_LENGTH);

    uint8_t* q = p;

    relay_write_bytes(&p, relay_token, RELAY_TOKEN_BYTES);

    int encrypt_length = int(p - q);

    if (crypto_box_easy(q, q, encrypt_length, nonce, router_public_key, relay_private_key) != 0) {
        return RELAY_ERROR;
    }

    int init_length = (int)(p - init_data) + encrypt_length + crypto_box_MACBYTES;

    struct curl_slist* slist = curl_slist_append(NULL, "Content-Type:application/octet-stream");

    curl_buffer_t init_response_buffer;
    init_response_buffer.size = 0;
    init_response_buffer.max_size = 1024;
    init_response_buffer.data = (uint8_t*)alloca(init_response_buffer.max_size);

    char init_url[1024];
    sprintf(init_url, "%s/relay_init", hostname);

    curl_easy_setopt(curl, CURLOPT_BUFFERSIZE, 102400L);
    curl_easy_setopt(curl, CURLOPT_URL, init_url);
    curl_easy_setopt(curl, CURLOPT_NOPROGRESS, 1L);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, init_data);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)init_length);
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, slist);
    curl_easy_setopt(curl, CURLOPT_USERAGENT, "network next relay");
    curl_easy_setopt(curl, CURLOPT_MAXREDIRS, 50L);
    curl_easy_setopt(curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS);
    curl_easy_setopt(curl, CURLOPT_TCP_KEEPALIVE, 1L);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT_MS, long(1000));
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &init_response_buffer);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, &curl_buffer_write_function);

    CURLcode ret = curl_easy_perform(curl);

    curl_slist_free_all(slist);
    slist = NULL;

    if (ret != 0) {
        return RELAY_ERROR;
    }

    long code;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &code);
    if (code != 200) {
        return RELAY_ERROR;
    }

    if (init_response_buffer.size < 4) {
        relay_printf("\nerror: bad relay init response size. too small to have valid data (%d)\n\n", init_response_buffer.size);
        return RELAY_ERROR;
    }

    const uint8_t* r = init_response_buffer.data;

    uint32_t version = relay_read_uint32(&r);

    const uint32_t init_response_version = 0;

    if (version != init_response_version) {
        relay_printf("\nerror: bad relay init response version. expected %d, got %d\n\n", init_response_version, version);
        return RELAY_ERROR;
    }

    if (init_response_buffer.size != 4 + 8 + RELAY_TOKEN_BYTES) {
        relay_printf("\nerror: bad relay init response size. expected %d bytes, got %d\n\n",
            RELAY_TOKEN_BYTES,
            init_response_buffer.size);
        return RELAY_ERROR;
    }

    *router_timestamp = relay_read_uint64(&r);

    memcpy(relay_token, init_response_buffer.data + 4 + 8, RELAY_TOKEN_BYTES);

    return RELAY_OK;
}

int relay_update(CURL* curl,
    const char* hostname,
    const uint8_t* relay_token,
    const char* relay_address,
    uint8_t* update_response_memory,
    relay_t* relay)
{
    // build update data

    uint32_t update_version = 0;

    uint8_t update_data[10 * 1024];

    uint8_t* p = update_data;
    relay_write_uint32(&p, update_version);
    relay_write_string(&p, relay_address, 256);
    relay_write_bytes(&p, relay_token, RELAY_TOKEN_BYTES);

    relay_platform_mutex_acquire(relay->mutex);
    relay_stats_t stats;
    relay_manager_get_stats(relay->relay_manager, &stats);
    relay_platform_mutex_release(relay->mutex);

    relay_write_uint32(&p, stats.num_relays);
    for (int i = 0; i < stats.num_relays; ++i) {
        relay_write_uint64(&p, stats.relay_ids[i]);
        relay_write_float32(&p, stats.relay_rtt[i]);
        relay_write_float32(&p, stats.relay_jitter[i]);
        relay_write_float32(&p, stats.relay_packet_loss[i]);
    }

    int update_data_length = (int)(p - update_data);

    // post it to backend

    struct curl_slist* slist = curl_slist_append(NULL, "Content-Type:application/octet-stream");

    curl_buffer_t update_response_buffer;
    update_response_buffer.size = 0;
    update_response_buffer.max_size = RESPONSE_MAX_BYTES;
    update_response_buffer.data = (uint8_t*)update_response_memory;

    char update_url[1024];
    sprintf(update_url, "%s/relay_update", hostname);

    curl_easy_setopt(curl, CURLOPT_BUFFERSIZE, 102400L);
    curl_easy_setopt(curl, CURLOPT_URL, update_url);
    curl_easy_setopt(curl, CURLOPT_NOPROGRESS, 1L);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, update_data);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)update_data_length);
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, slist);
    curl_easy_setopt(curl, CURLOPT_USERAGENT, "network next relay");
    curl_easy_setopt(curl, CURLOPT_MAXREDIRS, 50L);
    curl_easy_setopt(curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS);
    curl_easy_setopt(curl, CURLOPT_TCP_KEEPALIVE, 1L);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT_MS, long(1000));
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &update_response_buffer);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, &curl_buffer_write_function);

    CURLcode ret = curl_easy_perform(curl);

    curl_slist_free_all(slist);
    slist = NULL;

    if (ret != 0) {
        relay_printf("\nerror: could not post relay update\n\n");
        return RELAY_ERROR;
    }

    long code;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &code);
    if (code != 200) {
        relay_printf("\nerror: relay update response was %d, expected 200\n\n", int(code));
        return RELAY_ERROR;
    }

    // parse update response

    const uint8_t* q = update_response_buffer.data;

    uint32_t version = relay_read_uint32(&q);

    const uint32_t update_response_version = 0;

    if (version != update_response_version) {
        relay_printf("\nerror: bad relay update response version. expected %d, got %d\n\n", update_response_version, version);
        return RELAY_ERROR;
    }

    uint32_t num_relays = relay_read_uint32(&q);

    if (num_relays > MAX_RELAYS) {
        relay_printf("\nerror: too many relays to ping. max is %d, got %d\n\n", MAX_RELAYS, version);
        return RELAY_ERROR;
    }

    bool error = false;

    struct relay_ping_data_t
    {
        uint64_t id;
        relay_address_t address;
    };

    relay_ping_data_t relay_ping_data[MAX_RELAYS];

    for (uint32_t i = 0; i < num_relays; ++i) {
        char address_string[RELAY_MAX_ADDRESS_STRING_LENGTH];
        relay_ping_data[i].id = relay_read_uint64(&q);
        relay_read_string(&q, address_string, RELAY_MAX_ADDRESS_STRING_LENGTH);
        if (relay_address_parse(&relay_ping_data[i].address, address_string) != RELAY_OK) {
            error = true;
            break;
        }
    }

    if (error) {
        relay_printf("\nerror: error while reading set of relays to ping in update response\n\n");
        return RELAY_ERROR;
    }

    relay_platform_mutex_acquire(relay->mutex);
    relay->num_relays = num_relays;
    for (int i = 0; i < int(num_relays); ++i) {
        relay->relay_ids[i] = relay_ping_data[i].id;
        relay->relay_addresses[i] = relay_ping_data[i].address;
    }
    relay->relays_dirty = true;
    relay_platform_mutex_release(relay->mutex);

    return RELAY_OK;
}

static volatile uint64_t quit = 0;

void interrupt_handler(int signal)
{
    (void)signal;
    quit = 1;
}

uint64_t relay_timestamp(relay_t* relay)
{
    assert(relay);
    double current_time = relay_platform_time();
    uint64_t seconds_since_initialize = uint64_t(current_time - relay->initialize_time);
    return relay->initialize_router_timestamp + seconds_since_initialize;
}

uint64_t relay_clean_sequence(uint64_t sequence)
{
    uint64_t mask = ~((1ULL << 63) | (1ULL << 62));
    return sequence & mask;
}

static relay_platform_thread_return_t RELAY_PLATFORM_THREAD_FUNC receive_thread_function(void* context)
{
    relay_t* relay = (relay_t*)context;

    uint8_t packet_data[RELAY_MAX_PACKET_BYTES];

    while (!quit) {
        relay_address_t from;
        const int packet_bytes = relay_platform_socket_receive_packet(relay->socket, &from, packet_data, sizeof(packet_data));
        if (packet_bytes == 0)
            continue;
        if (packet_data[0] == RELAY_PING_PACKET && packet_bytes == 9) {
            packet_data[0] = RELAY_PONG_PACKET;
            relay_platform_socket_send_packet(relay->socket, &from, packet_data, 9);
        } else if (packet_data[0] == RELAY_PONG_PACKET && packet_bytes == 9) {
            relay_platform_mutex_acquire(relay->mutex);
            const uint8_t* p = packet_data + 1;
            uint64_t sequence = relay_read_uint64(&p);
            relay_manager_process_pong(relay->relay_manager, &from, sequence);
            relay_platform_mutex_release(relay->mutex);
        } else if (packet_data[0] == RELAY_ROUTE_REQUEST_PACKET) {
            if (packet_bytes < int(1 + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES * 2)) {
                relay_printf("ignoring route request. bad packet size (%d)", packet_bytes);
                continue;
            }
            uint8_t* p = &packet_data[1];
            relay_route_token_t token;
            if (relay_read_encrypted_route_token(&p, &token, relay->router_public_key, relay->relay_private_key) != RELAY_OK) {
                relay_printf("ignoring route request. could not read route token");
                continue;
            }
            if (token.expire_timestamp < relay_timestamp(relay)) {
                continue;
            }
            uint64_t hash = token.session_id ^ token.session_version;
            if (relay->sessions->find(hash) == relay->sessions->end()) {
                relay_session_t* session = (relay_session_t*)malloc(sizeof(relay_session_t));
                assert(session);
                session->expire_timestamp = token.expire_timestamp;
                session->session_id = token.session_id;
                session->session_version = token.session_version;
                session->client_to_server_sequence = 0;
                session->server_to_client_sequence = 0;
                session->kbps_up = token.kbps_up;
                session->kbps_down = token.kbps_down;
                session->prev_address = from;
                session->next_address = token.next_address;
                memcpy(session->private_key, token.private_key, crypto_box_SECRETKEYBYTES);
                relay_replay_protection_reset(&session->replay_protection_client_to_server);
                relay_replay_protection_reset(&session->replay_protection_server_to_client);
                relay->sessions->insert(std::make_pair(hash, session));
                printf("session created: %" PRIx64 ".%d\n", token.session_id, token.session_version);
            }
            packet_data[RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES] = RELAY_ROUTE_REQUEST_PACKET;
            relay_platform_socket_send_packet(relay->socket,
                &token.next_address,
                packet_data + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES,
                packet_bytes - RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES);
        } else if (packet_data[0] == RELAY_ROUTE_RESPONSE_PACKET) {
            if (packet_bytes != RELAY_HEADER_BYTES) {
                continue;
            }
            uint8_t type;
            uint64_t sequence;
            uint64_t session_id;
            uint8_t session_version;
            if (relay_peek_header(RELAY_DIRECTION_SERVER_TO_CLIENT,
                    &type,
                    &sequence,
                    &session_id,
                    &session_version,
                    packet_data,
                    packet_bytes) != RELAY_OK) {
                continue;
            }
            uint64_t hash = session_id ^ session_version;
            relay_session_t* session = (*(relay->sessions))[hash];
            if (!session) {
                continue;
            }
            if (session->expire_timestamp < relay_timestamp(relay)) {
                continue;
            }
            uint64_t clean_sequence = relay_clean_sequence(sequence);
            if (clean_sequence <= session->server_to_client_sequence) {
                continue;
            }
            session->server_to_client_sequence = clean_sequence;
            if (relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet_data, packet_bytes) !=
                RELAY_OK) {
                continue;
            }
            relay_platform_socket_send_packet(relay->socket, &session->prev_address, packet_data, packet_bytes);
        } else if (packet_data[0] == RELAY_CONTINUE_REQUEST_PACKET) {
            if (packet_bytes < int(1 + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES * 2)) {
                relay_printf("ignoring continue request. bad packet size (%d)", packet_bytes);
                continue;
            }
            uint8_t* p = &packet_data[1];
            relay_continue_token_t token;
            if (relay_read_encrypted_continue_token(&p, &token, relay->router_public_key, relay->relay_private_key) !=
                RELAY_OK) {
                relay_printf("ignoring continue request. could not read continue token");
                continue;
            }
            if (token.expire_timestamp < relay_timestamp(relay)) {
                continue;
            }
            uint64_t hash = token.session_id ^ token.session_version;
            relay_session_t* session = (*(relay->sessions))[hash];
            if (!session) {
                continue;
            }
            if (session->expire_timestamp < relay_timestamp(relay)) {
                continue;
            }
            if (session->expire_timestamp != token.expire_timestamp) {
                printf("session continued: %" PRIx64 ".%d\n", token.session_id, token.session_version);
            }
            session->expire_timestamp = token.expire_timestamp;
            packet_data[RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES] = RELAY_CONTINUE_REQUEST_PACKET;
            relay_platform_socket_send_packet(relay->socket,
                &session->next_address,
                packet_data + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES,
                packet_bytes - RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES);
        } else if (packet_data[0] == RELAY_CONTINUE_RESPONSE_PACKET) {
            if (packet_bytes != RELAY_HEADER_BYTES) {
                continue;
            }
            uint8_t type;
            uint64_t sequence;
            uint64_t session_id;
            uint8_t session_version;
            if (relay_peek_header(RELAY_DIRECTION_SERVER_TO_CLIENT,
                    &type,
                    &sequence,
                    &session_id,
                    &session_version,
                    packet_data,
                    packet_bytes) != RELAY_OK) {
                continue;
            }
            uint64_t hash = session_id ^ session_version;
            relay_session_t* session = (*(relay->sessions))[hash];
            if (!session) {
                continue;
            }
            if (session->expire_timestamp < relay_timestamp(relay)) {
                continue;
            }
            uint64_t clean_sequence = relay_clean_sequence(sequence);
            if (clean_sequence <= session->server_to_client_sequence) {
                continue;
            }
            session->server_to_client_sequence = clean_sequence;
            if (relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet_data, packet_bytes) !=
                RELAY_OK) {
                continue;
            }
            relay_platform_socket_send_packet(relay->socket, &session->prev_address, packet_data, packet_bytes);
        } else if (packet_data[0] == RELAY_CLIENT_TO_SERVER_PACKET) {
            if (packet_bytes <= RELAY_HEADER_BYTES || packet_bytes > RELAY_HEADER_BYTES + RELAY_MTU) {
                continue;
            }
            uint8_t type;
            uint64_t sequence;
            uint64_t session_id;
            uint8_t session_version;
            if (relay_peek_header(RELAY_DIRECTION_CLIENT_TO_SERVER,
                    &type,
                    &sequence,
                    &session_id,
                    &session_version,
                    packet_data,
                    packet_bytes) != RELAY_OK) {
                continue;
            }
            uint64_t hash = session_id ^ session_version;
            relay_session_t* session = (*(relay->sessions))[hash];
            if (!session) {
                continue;
            }
            if (session->expire_timestamp < relay_timestamp(relay)) {
                continue;
            }
            uint64_t clean_sequence = relay_clean_sequence(sequence);
            if (relay_replay_protection_already_received(&session->replay_protection_client_to_server, clean_sequence)) {
                continue;
            }
            relay_replay_protection_advance_sequence(&session->replay_protection_client_to_server, clean_sequence);
            if (relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, session->private_key, packet_data, packet_bytes) !=
                RELAY_OK) {
                continue;
            }
            relay_platform_socket_send_packet(relay->socket, &session->next_address, packet_data, packet_bytes);
        } else if (packet_data[0] == RELAY_SERVER_TO_CLIENT_PACKET) {
            if (packet_bytes <= RELAY_HEADER_BYTES || packet_bytes > RELAY_HEADER_BYTES + RELAY_MTU) {
                continue;
            }
            uint8_t type;
            uint64_t sequence;
            uint64_t session_id;
            uint8_t session_version;
            if (relay_peek_header(RELAY_DIRECTION_SERVER_TO_CLIENT,
                    &type,
                    &sequence,
                    &session_id,
                    &session_version,
                    packet_data,
                    packet_bytes) != RELAY_OK) {
                continue;
            }
            uint64_t hash = session_id ^ session_version;
            relay_session_t* session = (*(relay->sessions))[hash];
            if (!session) {
                continue;
            }
            if (session->expire_timestamp < relay_timestamp(relay)) {
                continue;
            }
            uint64_t clean_sequence = relay_clean_sequence(sequence);
            if (relay_replay_protection_already_received(&session->replay_protection_server_to_client, clean_sequence)) {
                continue;
            }
            relay_replay_protection_advance_sequence(&session->replay_protection_server_to_client, clean_sequence);
            if (relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet_data, packet_bytes) !=
                RELAY_OK) {
                continue;
            }
            relay_platform_socket_send_packet(relay->socket, &session->prev_address, packet_data, packet_bytes);
        } else if (packet_data[0] == RELAY_SESSION_PING_PACKET) {
            if (packet_bytes > RELAY_HEADER_BYTES + 32) {
                continue;
            }
            uint8_t type;
            uint64_t sequence;
            uint64_t session_id;
            uint8_t session_version;
            if (relay_peek_header(RELAY_DIRECTION_CLIENT_TO_SERVER,
                    &type,
                    &sequence,
                    &session_id,
                    &session_version,
                    packet_data,
                    packet_bytes) != RELAY_OK) {
                continue;
            }
            uint64_t hash = session_id ^ session_version;
            relay_session_t* session = (*(relay->sessions))[hash];
            if (!session) {
                continue;
            }
            if (session->expire_timestamp < relay_timestamp(relay)) {
                continue;
            }
            uint64_t clean_sequence = relay_clean_sequence(sequence);
            if (clean_sequence <= session->client_to_server_sequence) {
                continue;
            }
            session->client_to_server_sequence = clean_sequence;
            if (relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, session->private_key, packet_data, packet_bytes) !=
                RELAY_OK) {
                continue;
            }
            relay_platform_socket_send_packet(relay->socket, &session->next_address, packet_data, packet_bytes);
        } else if (packet_data[0] == RELAY_SESSION_PONG_PACKET) {
            if (packet_bytes > RELAY_HEADER_BYTES + 32) {
                continue;
            }
            uint8_t type;
            uint64_t sequence;
            uint64_t session_id;
            uint8_t session_version;
            if (relay_peek_header(RELAY_DIRECTION_SERVER_TO_CLIENT,
                    &type,
                    &sequence,
                    &session_id,
                    &session_version,
                    packet_data,
                    packet_bytes) != RELAY_OK) {
                continue;
            }
            uint64_t hash = session_id ^ session_version;
            relay_session_t* session = (*(relay->sessions))[hash];
            if (!session) {
                continue;
            }
            if (session->expire_timestamp < relay_timestamp(relay)) {
                continue;
            }
            uint64_t clean_sequence = relay_clean_sequence(sequence);
            if (clean_sequence <= session->server_to_client_sequence) {
                continue;
            }
            session->server_to_client_sequence = clean_sequence;
            if (relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet_data, packet_bytes) !=
                RELAY_OK) {
                continue;
            }
            relay_platform_socket_send_packet(relay->socket, &session->prev_address, packet_data, packet_bytes);
        } else if (packet_data[0] == RELAY_NEAR_PING_PACKET) {
            if (packet_bytes != 1 + 8 + 8 + 8 + 8) {
                continue;
            }
            packet_data[0] = RELAY_NEAR_PONG_PACKET;
            relay_platform_socket_send_packet(relay->socket, &from, packet_data, packet_bytes - 16);
        }
    }

    RELAY_PLATFORM_THREAD_RETURN();
}

static relay_platform_thread_return_t RELAY_PLATFORM_THREAD_FUNC ping_thread_function(void* context)
{
    relay_t* relay = (relay_t*)context;

    while (!quit) {
        relay_platform_mutex_acquire(relay->mutex);

        if (relay->relays_dirty) {
            relay_manager_update(relay->relay_manager, relay->num_relays, relay->relay_ids, relay->relay_addresses);
            relay->relays_dirty = false;
        }

        double current_time = relay_platform_time();

        struct ping_data_t
        {
            uint64_t sequence;
            relay_address_t address;
        };

        int num_pings = 0;
        ping_data_t pings[MAX_RELAYS];

        for (int i = 0; i < relay->relay_manager->num_relays; ++i) {
            if (relay->relay_manager->relay_last_ping_time[i] + RELAY_PING_TIME <= current_time) {
                pings[num_pings].sequence =
                    relay_ping_history_ping_sent(relay->relay_manager->relay_ping_history[i], current_time);
                pings[num_pings].address = relay->relay_manager->relay_addresses[i];
                relay->relay_manager->relay_last_ping_time[i] = current_time;
                num_pings++;
            }
        }

        relay_platform_mutex_release(relay->mutex);

        for (int i = 0; i < num_pings; ++i) {
            uint8_t packet_data[9];
            packet_data[0] = RELAY_PING_PACKET;
            uint8_t* p = packet_data + 1;
            relay_write_uint64(&p, pings[i].sequence);
            relay_platform_socket_send_packet(relay->socket, &pings[i].address, packet_data, 9);
        }

        relay_platform_sleep(1.0 / 100.0);
    }

    RELAY_PLATFORM_THREAD_RETURN();
}

int main(int argc, const char** argv)
{
    if (argc == 2 && strcmp(argv[1], "test") == 0) {
        testing::relay_test();
        return 0;
    }

    printf("\nNetwork Next Relay\n");

    printf("\nEnvironment:\n\n");

    const char* relay_address_env = relay_platform_getenv("RELAY_ADDRESS");
    if (!relay_address_env) {
        printf("\nerror: RELAY_ADDRESS not set\n\n");
        return 1;
    }

    relay::relay_address_t relay_address;
    if (relay_address_parse(&relay_address, relay_address_env) != RELAY_OK) {
        printf("\nerror: invalid relay address '%s'\n\n", relay_address_env);
        return 1;
    }

    {
        relay::relay_address_t address_without_port = relay_address;
        address_without_port.port = 0;
        char address_buffer[RELAY_MAX_ADDRESS_STRING_LENGTH];
        printf("    relay address is '%s'\n", relay_address_to_string(&address_without_port, address_buffer));
    }

    uint16_t relay_bind_port = relay_address.port;

    printf("    relay bind port is %d\n", relay_bind_port);

    const char* relay_private_key_env = relay_platform_getenv("RELAY_PRIVATE_KEY");
    if (!relay_private_key_env) {
        printf("\nerror: RELAY_PRIVATE_KEY not set\n\n");
        return 1;
    }

    uint8_t relay_private_key[RELAY_PRIVATE_KEY_BYTES];
    if (encoding::relay_base64_decode_data(relay_private_key_env, relay_private_key, RELAY_PRIVATE_KEY_BYTES) !=
        RELAY_PRIVATE_KEY_BYTES) {
        printf("\nerror: invalid relay private key\n\n");
        return 1;
    }

    printf("    relay private key is '%s'\n", relay_private_key_env);

    const char* relay_public_key_env = relay_platform_getenv("RELAY_PUBLIC_KEY");
    if (!relay_public_key_env) {
        printf("\nerror: RELAY_PUBLIC_KEY not set\n\n");
        return 1;
    }

    uint8_t relay_public_key[RELAY_PUBLIC_KEY_BYTES];
    if (encoding::relay_base64_decode_data(relay_public_key_env, relay_public_key, RELAY_PUBLIC_KEY_BYTES) !=
        RELAY_PUBLIC_KEY_BYTES) {
        printf("\nerror: invalid relay public key\n\n");
        return 1;
    }

    printf("    relay public key is '%s'\n", relay_public_key_env);

    const char* router_public_key_env = relay_platform_getenv("RELAY_ROUTER_PUBLIC_KEY");
    if (!router_public_key_env) {
        printf("\nerror: RELAY_ROUTER_PUBLIC_KEY not set\n\n");
        return 1;
    }

    uint8_t router_public_key[crypto_sign_PUBLICKEYBYTES];
    if (encoding::relay_base64_decode_data(router_public_key_env, router_public_key, crypto_sign_PUBLICKEYBYTES) !=
        crypto_sign_PUBLICKEYBYTES) {
        printf("\nerror: invalid router public key\n\n");
        return 1;
    }

    printf("    router public key is '%s'\n", router_public_key_env);

    const char* backend_hostname = relay_platform_getenv("RELAY_BACKEND_HOSTNAME");
    if (!backend_hostname) {
        printf("\nerror: RELAY_BACKEND_HOSTNAME not set\n\n");
        return 1;
    }

    printf("    backend hostname is '%s'\n", backend_hostname);

    if (relay_initialize() != RELAY_OK) {
        printf("\nerror: failed to initialize relay\n\n");
        return 1;
    }

    relay_platform_socket_t* socket =
        relay_platform_socket_create(&relay_address, RELAY_PLATFORM_SOCKET_BLOCKING, 0.1f, 100 * 1024, 100 * 1024);
    if (socket == NULL) {
        printf("\ncould not create socket\n\n");
        relay_term();
        return 1;
    }

    printf("\nRelay socket opened on port %d\n", relay_address.port);
    char relay_address_buffer[RELAY_MAX_ADDRESS_STRING_LENGTH];
    const char* relay_address_string = relay_address_to_string(&relay_address, relay_address_buffer);

    CURL* curl = curl_easy_init();
    if (!curl) {
        printf("\nerror: could not initialize curl\n\n");
        relay_platform_socket_destroy(socket);
        curl_easy_cleanup(curl);
        relay_term();
        return 1;
    }

    uint8_t relay_token[RELAY_TOKEN_BYTES];

    printf("\nInitializing relay\n");

    bool relay_initialized = false;

    uint64_t router_timestamp = 0;

    for (int i = 0; i < 60; ++i) {
        if (relay_init(curl,
                backend_hostname,
                relay_token,
                relay_address_string,
                router_public_key,
                relay_private_key,
                &router_timestamp) == RELAY_OK) {
            printf("\n");
            relay_initialized = true;
            break;
        }

        printf(".");
        fflush(stdout);

        relay_platform_sleep(1.0);
    }

    if (!relay_initialized) {
        printf("\nerror: could not initialize relay\n\n");
        relay_platform_socket_destroy(socket);
        curl_easy_cleanup(curl);
        relay_term();
        return 1;
    }

    relay_t relay;
    memset(&relay, 0, sizeof(relay));
    relay.initialize_time = relay_platform_time();
    relay.initialize_router_timestamp = router_timestamp;
    relay.sessions = new std::map<uint64_t, relay_session_t*>();
    memcpy(relay.relay_public_key, relay_public_key, RELAY_PUBLIC_KEY_BYTES);
    memcpy(relay.relay_private_key, relay_private_key, RELAY_PRIVATE_KEY_BYTES);
    memcpy(relay.router_public_key, router_public_key, crypto_sign_PUBLICKEYBYTES);

    relay.socket = socket;
    relay.mutex = relay_platform_mutex_create();
    if (!relay.mutex) {
        printf("\nerror: could not create ping thread\n\n");
        quit = 1;
    }

    relay.relay_manager = relay_manager_create();
    if (!relay.relay_manager) {
        printf("\nerror: could not create relay manager\n\n");
        quit = 1;
    }

    relay_platform_thread_t* receive_thread = relay_platform_thread_create(receive_thread_function, &relay);
    if (!receive_thread) {
        printf("\nerror: could not create receive thread\n\n");
        quit = 1;
    }

    relay_platform_thread_t* ping_thread = relay_platform_thread_create(ping_thread_function, &relay);
    if (!ping_thread) {
        printf("\nerror: could not create ping thread\n\n");
        quit = 1;
    }

    printf("Relay initialized\n\n");

    signal(SIGINT, interrupt_handler);

    uint8_t* update_response_memory = (uint8_t*)malloc(RESPONSE_MAX_BYTES);

    while (!quit) {
        bool updated = false;

        for (int i = 0; i < 10; ++i) {
            if (relay_update(curl, backend_hostname, relay_token, relay_address_string, update_response_memory, &relay) ==
                RELAY_OK) {
                updated = true;
                break;
            }
        }

        if (!updated) {
            printf("error: could not update relay\n\n");
            quit = 1;
            break;
        }

        relay_platform_sleep(1.0);
    }

    printf("Cleaning up\n");

    if (receive_thread) {
        relay_platform_thread_join(receive_thread);
        relay_platform_thread_destroy(receive_thread);
    }

    if (ping_thread) {
        relay_platform_thread_join(ping_thread);
        relay_platform_thread_destroy(ping_thread);
    }

    free(update_response_memory);

    for (std::map<uint64_t, relay_session_t*>::iterator itor = relay.sessions->begin(); itor != relay.sessions->end(); ++itor) {
        delete itor->second;
    }

    delete relay.sessions;

    relay_manager_destroy(relay.relay_manager);

    relay_platform_mutex_destroy(relay.mutex);

    relay_platform_socket_destroy(socket);

    curl_easy_cleanup(curl);

    relay_term();

    return 0;
}
