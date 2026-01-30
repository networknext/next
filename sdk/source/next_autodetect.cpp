/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "next_autodetect.h"

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

#if defined(_WIN32) || defined(_WIN64)
#define strtok_r strtok_s
#endif

bool next_default_http_request_function( const char * url, const char * header, int timeout_seconds, char * output, size_t output_size )
{
    next_assert( url );
    next_assert( header );
    next_assert( output );
    next_assert( output_size > 0 );

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    char command_line[1024];
    if ( header[0] != '\0' )
    {
        snprintf( command_line, sizeof(command_line), "curl -s \"%s\" -H \"%s\" --max-time %d -s 2>/dev/null", url, header, timeout_seconds );
    }
    else
    {
        snprintf( command_line, sizeof(command_line), "curl -s \"%s\" --max-time %d -s 2>/dev/null", url, timeout_seconds );
    }

    next_printf( NEXT_LOG_LEVEL_SPAM, "command line: '%s'", command_line );

    FILE * file = popen( command_line, "r" );
    if ( !file )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "server autodetect datacenter: could not execute curl" );
        return false;
    }

    char * p = output;

    int num_lines = 0;

    next_printf( NEXT_LOG_LEVEL_SPAM, "---------------------------" );

    char line[1024];
    while ( fgets( line, sizeof(line), file ) != NULL )
    {
        int bytes_copied = next_copy_string( p, line, output_size - ( p - output ) );
        p += bytes_copied;
        num_lines++;

        for ( size_t i = 0; i < sizeof(line); i++ )
        {
            if ( line[i] == '\0' )
                break;

            if ( line[i] == '\n' || line[i] == '\r' )
            {
                line[i] = '\0';
            }
        }

        next_printf( NEXT_LOG_LEVEL_SPAM, "%s", line );
    }

    pclose( file );

    next_printf( NEXT_LOG_LEVEL_SPAM, "---------------------------" );

    return num_lines > 0;

#else // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    next_printf( NEXT_LOG_LEVEL_ERROR, "you must supply your own autodetect_http_request_function on windows" );

    return false;

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC
}

static bool (*next_http_request)( const char * url, const char * header, int timeout_seconds, char * output, size_t output_size ) = next_default_http_request_function;

void next_autodetect_http_request_function( bool (*function)( const char * url, const char * header, int timeout_seconds, char * output, size_t output_size ) )
{
    if ( function != NULL )
    {
        next_http_request = function;
    }
    else
    {
        next_http_request = next_default_http_request_function;
    }
}

bool next_autodetect_google( char * datacenter, size_t datacenter_size )
{
    // if we are in google cloud, we can hit the metadata server and ask it what google cloud zone we are in

    char buffer[32*1024];
    memset( buffer, 0, sizeof(buffer) );

    if ( !next_http_request( "http://metadata.google.internal/computeMetadata/v1/instance/zone", "Metadata-Flavor: Google", 2, buffer, sizeof(buffer) ) )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: not in google cloud" );
        return false;
    }

    next_printf( NEXT_LOG_LEVEL_DEBUG, "output: '%s'", buffer );

    char * p = strstr( buffer, "projects/" );
    if ( p == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "server autodetect datacenter: 'projects/' not found" );
        return false;
    }

    char * q = strstr( p + 9, "zones/" );
    if ( q == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "server autodetect datacenter: 'zone/' not found" );
        return false;
    }

    // we are in google cloud

    char zone[256];
    next_copy_string( zone, q + 6, sizeof(zone) );
    next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: google cloud zone is '%s'", zone );

    // download "google.txt" from google cloud storage. it contains lines mapping from google cloud zone to network next datacenter

    char url[1024];
    snprintf( url, sizeof(url), "https://storage.googleapis.com/%s/google.txt?ts=%x", NEXT_CONFIG_BUCKET_NAME, uint32_t(time(NULL)) );
    memset( buffer, 0, sizeof(buffer) );
    if ( !next_http_request( url, "", 5, buffer, sizeof(buffer) ) )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "server autodetect datacenter: could not download google.txt file" );
        return false;
    }

    // map our google cloud zone to a network next datacenter

    p = buffer;

    char * end = buffer + sizeof(buffer);

    int line = 0;
    while ( p < end && *p != '\0' )
    {
        q = p;
        while ( q < end && *q != '\n' && *q != '\r' && *q != '\0' )
        {
            q++;
        }
        if ( q >= end )
            break;
        *q = '\0';
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "line %d: '%s'", line, p );

            const char * separators = ",\n\r";
            char * rest;
            char * google_zone = strtok_r( p, separators, &rest );
            if ( google_zone && strcmp( zone, google_zone ) == 0 )
            {
                char * network_next_datacenter = strtok_r( NULL, separators, &rest );
                if ( network_next_datacenter )
                {
                    next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: '%s' -> '%s'", zone, network_next_datacenter );
                    next_copy_string( datacenter, network_next_datacenter, datacenter_size );
                    return true;
                }
            }
        }
        p = q + 1;
        line++;
    }

    next_printf( NEXT_LOG_LEVEL_WARN, "could not find network next datacenter for google cloud zone :(" );

    return false;
}

bool next_autodetect_amazon( char * datacenter, size_t datacenter_size )
{
    char buffer[32*1024];
    memset( buffer, 0, sizeof(buffer) );

    // Get the AZID from instance metadata
    // This is necessary because only AZ IDs are the same across different AWS accounts
    // See https://docs.aws.amazon.com/ram/latest/userguide/working-with-az-ids.html for details

    if ( !next_http_request( "http://169.254.169.254/latest/meta-data/placement/availability-zone-id", "", 2, buffer, sizeof(buffer) ) || strstr( buffer, "-az" ) == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: not in amazon cloud" );
        return false;
    }

    char azid[256];
    next_copy_string( azid, buffer, sizeof(azid) );

    size_t azid_length = strlen( azid );
    size_t index = azid_length - 1;
    while ( index > 0 && ( azid[index] == '\n' || azid[index] == '\r' ) )
    {
        azid[index] = '\0';
        index--;
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: amazon azid is '%s'", azid );

    // download "amazon.txt" from google cloud storage. it contains lines mapping from AZID to network next datacenter

    char url[1024];
    snprintf( url, sizeof(url), "https://storage.googleapis.com/%s/amazon.txt?ts=%x", NEXT_CONFIG_BUCKET_NAME, uint32_t(time(NULL)) );
    memset( buffer, 0, sizeof(buffer) );
    if ( !next_http_request( url, "", 5, buffer, sizeof(buffer) ) )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "server autodetect datacenter: could not download amazon.txt file" );
        return false;
    }

    // map our amazon azid to a network next datacenter

    char * p = buffer;

    char * end = buffer + sizeof(buffer);

    while ( p < end && *p != '\0' )
    {
        char * q = p;
        while ( q < end && *q != '\n' && *q != '\r' && *q != '\0' )
        {
            q++;
        }
        if ( q >= end )
            break;
        *q = '\0';
        {
            const char * separators = ",\n\r";
            char * rest;
            char * amazon_azid = strtok_r( p, separators, &rest );
            if ( amazon_azid && strcmp( azid, amazon_azid ) == 0 )
            {
                char * network_next_datacenter = strtok_r( NULL, separators, &rest );
                if ( network_next_datacenter )
                {
                    next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: '%s' -> '%s'", azid, network_next_datacenter );
                    next_copy_string( datacenter, network_next_datacenter, datacenter_size );
                    return true;
                }
            }
        }
        p = q + 1;
    }

    next_printf( NEXT_LOG_LEVEL_WARN, "server autodetect datacenter: could not find network next datacenter for azid :(" );

    return true;
}

bool next_autodetect_unity( const char * input_datacenter, const char * public_address, char * output_datacenter, size_t output_datacenter_size )
{
    if ( input_datacenter[0] != 'u' ||
         input_datacenter[1] != 'n' ||
         input_datacenter[2] != 'i' ||
         input_datacenter[3] != 't' ||
         input_datacenter[4] != 'y' ||
         input_datacenter[5] != '.' )
    {
        // not unity bare metal
        return false;
    }

    char url[1024];
    snprintf( url, sizeof(url), "%s/%s/%s", NEXT_AUTODETECT_URL, input_datacenter, public_address );
    char unity_data[1024];
    memset( unity_data, 0, sizeof(unity_data) );
    if ( !next_http_request( url, "", 10, unity_data, sizeof(unity_data) ) )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "server autodetect datacenter: no match found" );
        return false;
    }

    next_copy_string( output_datacenter, unity_data, output_datacenter_size );

    next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: '%s' -> '%s'", input_datacenter, output_datacenter );

    return true;
}

bool next_autodetect_datacenter( const char * input_datacenter, const char * public_address, char * output_datacenter, size_t output_datacenter_size )
{
    next_assert( input_datacenter );
    next_assert( public_address );
    next_assert( output_datacenter );
    next_assert( output_datacenter_size > 0 );

    if ( input_datacenter[0] == 'c' && 
         input_datacenter[1] == 'l' &&
         input_datacenter[2] == 'o' &&
         input_datacenter[3] == 'u' &&
         input_datacenter[4] == 'd' &&
         input_datacenter[5] == '\0' )
    {
        if ( next_autodetect_google( output_datacenter, output_datacenter_size ) )
            return true;

        if ( next_autodetect_amazon( output_datacenter, output_datacenter_size ) )
            return true;
    }

    if ( next_autodetect_unity( input_datacenter, public_address, output_datacenter, output_datacenter_size ) )
        return true;

    return false;
}

#else // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

void next_autodetect_http_request_function( bool (*function)( const char * url, const char * header, char * output, size_t output_size ) )
{
    (void) function;
}

bool next_autodetect_google( char * output_datacenter, size_t output_datacenter_size )
{
    next_printf( NEXT_LOG_LEVEL_WARN, "autodetect google datacenter is not available on this platform" );
    (void) output_datacenter;
    (void) output_datacenter_size;
    return false;
}

bool next_autodetect_amazon( char * output_datacenter, size_t output_datacenter_size )
{
    next_printf( NEXT_LOG_LEVEL_WARN, "autodetect amazon datacenter is not available on this platform" );
    (void) output_datacenter;
    (void) output_datacenter_size;
    return false;   
}

bool next_autodetect_unity( const char * input_datacenter, const char * public_address, char * output_datacenter, size_t output_datacenter_size )
{
    next_printf( NEXT_LOG_LEVEL_WARN, "autodetect unity datacenter is not available on this platform" );
    (void) input_datacenter;
    (void) public_address;
    (void) output_datacenter;
    (void) output_datacenter_size;
    return false;
}

bool next_autodetect_datacenter( const char * input_datacenter, const char * public_address, char * output_datacenter, size_t output_datacenter_size )
{
    next_printf( NEXT_LOG_LEVEL_WARN, "autodetect datacenter is not available on this platform" );
    (void) input_datacenter;
    (void) public_address;
    (void) output_datacenter;
    (void) output_datacenter_size;
    return false;
}

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS
