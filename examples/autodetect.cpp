/*
    Network Next SDK. Copyright © 2017 - 2021 Network Next, Inc.

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

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

bool next_autodetect_google( char * output )
{
    FILE * file;
    char buffer[1024*10];

    // are we running in google cloud?

    file = popen( "/bin/ls /usr/bin | grep google_ 2>/dev/null", "r");
    if ( file == NULL ) 
    {
        printf( "autodetect: could not run ls\n" );
        return false;
    }

    bool in_gcp = false;
    while ( fgets(buffer, sizeof(buffer), file ) != NULL ) 
    {
        if ( strstr( buffer, "google_authorized_keys" ) != NULL )
        {
            printf( "autodetect: running in google cloud\n" );
            in_gcp = true;
            break;
        }
    }
    pclose( file );

    // we are not running in google cloud :(

    if ( !in_gcp )
    {
        printf( "autodetect: not in google cloud\n" );
        return false;
    }

    // we are running in google cloud, which zone are we in?

    char zone[256];
    zone[0] = '\0';
    file = popen( "curl \"http://metadata.google.internal/computeMetadata/v1/instance/zone\" -H \"Metadata-Flavor: Google\" --max-time 1 -vs 2>/dev/null", "r" );
    while ( fgets(buffer, sizeof(buffer), file ) != NULL ) 
    {
        int length = strlen( buffer );
        if ( length < 10 )
        {
            continue;
        }

        if ( buffer[0] != 'p' ||
             buffer[1] != 'r' || 
             buffer[2] != 'o' ||
             buffer[3] != 'j' || 
             buffer[4] != 'e' ||
             buffer[5] != 'c' ||
             buffer[6] != 't' ||
             buffer[7] != 's' ||
             buffer[8] != '/' )
        {
            continue;
        }

        bool found = false;
        int index = length - 1;
        while ( index > 10 && length  )
        {
            if ( buffer[index] == '/' )
            {
                found = true;
                break;
            }
            index--;
        }

        if ( !found )
        {
            continue;
        }

        strcpy( zone, buffer + index + 1 );

        int zone_length = strlen(zone);
        index = zone_length - 1;
        while ( index > 0 && ( zone[index] == '\n' || zone[index] == '\r' ) )
        {
            zone[index] = '\0';
            index--;
        }

        printf( "autodetect: google zone is \"%s\"\n", zone );

        break;
    }
    pclose( file );

    // we couldn't work out which zone we are in :(

    if ( zone[0] != '\0' )
    {
        printf( "autodetect: could not detect google zone\n" );
        return false;
    }

    // look up google zone -> network next datacenter via mapping in google cloud storage "google.txt" file

    bool found = false;
    file = popen( "curl https://storage.googleapis.com/network-next-sdk/google.txt --max-time 1 -vs 2>/dev/null", "r" );
    while ( fgets(buffer, sizeof(buffer), file ) != NULL ) 
    {
        const char * separators = ",\n\r";

        char * google_zone = strtok( buffer, separators );
        if ( google_zone == NULL )
        {
            continue;
        }

        char * google_datacenter = strtok( NULL, separators );
        if ( google_datacenter == NULL )
        {
            continue;
        }

        if ( strcmp( zone, google_zone ) == 0 )
        {
            printf( "autodetect: \"%s\" -> \"%s\"\n", zone, google_datacenter );
            strcpy( output, google_datacenter );
            found = true;
            break;
        }
    }
    pclose( file );

    return found;
}

bool next_autodetect_datacenter( char * output )
{
    // we need linux + curl to do any autodetect. bail if we don't have it

    printf( "\nautodetect: looking for curl\n" );

    int result = system( "curl >/dev/null 2>&1" );

    if ( result < 0 )
    {
        printf( "autodetect: curl not found\n" );
        return false;
    }

    printf( "autodetect: curl exists\n" );

    // google cloud

    bool google_result = next_autodetect_google( output );
    if ( google_result )
    {
        return true;
    }

    // todo: aws

    // ...

    // nope

    printf( "autodetect: could not autodetect datacenter\n" );

    return false;
}

int main()
{
    char datacenter[1024];
    bool result = next_autodetect_datacenter( datacenter );
    if ( result )
    {
        printf( "\ndatacenter is \"%s\"\n\n", datacenter );    
    }
    else
    {
        printf( "\ncould not autodetect datacenter\n\n" );
    }

    fflush( stdout );

    return 0;
}
