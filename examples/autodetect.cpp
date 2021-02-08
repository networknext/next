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

#include "next.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

const char * next_autodetect_gcp( int num_zones, const char ** zones, const char ** datacenters )
{
    FILE * file;
    char buffer[1024*10];

    // are we running in google cloud?

    file = popen( "/bin/ls /usr/bin | grep google_ 2>/dev/null", "r");
    if ( file == NULL ) 
    {
        printf( "autodetect: could not run ls\n" );
        return NULL;
    }

    bool in_gcp = false;
    while ( fgets(buffer, sizeof(buffer), file ) != NULL ) 
    {
        printf( "%s", buffer );
        if ( strcmp( buffer, "google_authorized_keys\n" ) == 0 )
        {
            printf( "autodetect: running in google cloud\n" );
            break;
        }
    }
    pclose( file );

    if ( !in_gcp )
    {
        printf( "autodetect: not in google cloud\n" );
        return NULL;
    }

    // we are running in google cloud, which zone are we in?

    char * zone = NULL;
    file = popen( "curl \"http://metadata.google.internal/computeMetadata/v1/instance/zone\" -H \"Metadata-Flavor: Google\" --max-time 1 -vs 2>/dev/null", "r" );
    while ( fgets(buffer, sizeof(buffer), file ) != NULL ) 
    {
        int length = strlen( buffer );
        if ( length < 10 )
            continue;

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
            continue;

        zone = (char*) &buffer[index+1];

        int zone_length = strlen(zone);
        index = zone_length - 1;
        while ( index > 0 && ( zone[index] == '\n' || zone[index] == '\r' ) )
        {
            zone[index] = '\0';
            index--;
        }

        printf( "google zone = \"%s\"\n", zone );

        break;
    }
    pclose( file );

    // if we couldn't look up which zone, we cannot autodetect gcp

    if ( zone == NULL )
        return NULL;

    // look up network next datacenter from google zone string

    for ( int i = 0; i < num_zones; ++i )
    {
        if ( strcmp( zone, zones[i] ) == 0 )
        {
            return datacenters[i];
        }
    }

    return NULL;
}

const char * next_autodetect_datacenter()
{
    // we need linux + curl to do any autodetect. bail if we don't have it

    printf( "\nautodetect: looking for curl\n" );

    int result = system( "curl >/dev/null 2>&1" );

    if ( result < 0 )
    {
        printf( "autodetect: curl not found\n" );
        return NULL;
    }

    printf( "autodetect: curl exists\n" );

    // google cloud

    const char * google_zones[] = 
    { 
        "northamerica-northeast1-a",
        "northamerica-northeast1-b",
        "northamerica-northeast1-c",
        "southamerica-east1-a",
        "southamerica-east1-b",
        "southamerica-east1-c",
        "us-central1-a",
        "us-central1-b",
        "us-central1-c",
        "us-central1-f",
        "us-east1-b",
        "us-east1-c",
        "us-east1-d",
        "us-east4-a",
        "us-east4-b",
        "us-east4-c",
        "us-west2-a",
        "us-west2-b",
        "us-west2-c",
        "us-west3-a",
        "us-west3-b",
        "us-west3-c",
        "us-west4-a",
        "us-west4-b",
        "us-west4-c",
        "europe-north1-a",
        "europe-north1-b",
        "europe-north1-c",
        "europe-west1-b",
        "europe-west1-c",
        "europe-west1-d",
        "europe-west2-a",
        "europe-west2-b",
        "europe-west2-c",
        "europe-west3-a",
        "europe-west3-b",
        "europe-west3-c",
        "europe-west4-a",
        "europe-west4-b",
        "europe-west4-c",
        "europe-west6-a",
        "europe-west6-b",
        "europe-west6-c",
        "asia-east1-a",
        "asia-east1-b",
        "asia-east1-c",
        "asia-east2-a",
        "asia-east2-b",
        "asia-east2-c",
        "asia-northeast1-a",
        "asia-northeast1-b",
        "asia-northeast1-c",
        "asia-northeast2-a",
        "asia-northeast2-b",
        "asia-northeast2-c",
        "asia-northeast3-a",
        "asia-northeast3-b",
        "asia-northeast3-c",
        "asia-south1-a",
        "asia-south1-b",
        "asia-south1-c",
        "asia-southeast1-a",
        "asia-southeast1-b",
        "asia-southeast1-c",
        "asia-southeast2-a",
        "asia-southeast2-b",
        "asia-southeast2-c",
        "australia-southeast1-a",
        "australia-southeast1-b",
        "australia-southeast1-c",
    };
    const char * google_datacenters[] = 
    { 
        "google.montreal.1",
        "google.montreal.2",
        "google.montreal.3",
        "google.saopaulo.1",
        "google.saopaulo.2",
        "google.saopaulo.3",
        "google.iowa.1",
        "google.iowa.2",
        "google.iowa.3",
        "google.iowa.6",
        "google.southcarolina.2",
        "google.southcarolina.3",
        "google.southcarolina.4",
        "google.nothernvirginia.1",
        "google.nothernvirginia.2",
        "google.nothernvirginia.3",
        "google.losangeles.1",
        "google.losangeles.2",
        "google.losangeles.3",
        "google.saltlakecity.1",
        "google.saltlakecity.2",
        "google.saltlakecity.3",
        "google.lasvegas.1",
        "google.lasvegas.2",
        "google.lasvegas.3",
        "google.finland.1",
        "google.finland.2",
        "google.finland.3",
        "google.belgium.2",
        "google.belgium.3",
        "google.belgium.4",
        "google.london.1",
        "google.london.2",
        "google.london.3",
        "google.frankfurt.1",
        "google.frankfurt.2",
        "google.frankfurt.3",
        "google.netherlands.1",
        "google.netherlands.2",
        "google.netherlands.3",
        "google.zurich.1",
        "google.zurich.2",
        "google.zurich.3",
        "google.taiwan.1",
        "google.taiwan.2",
        "google.taiwan.3",
        "google.hongkong.1",
        "google.hongkong.2",
        "google.hongkong.3",
        "google.tokyo.1",
        "google.tokyo.2",
        "google.tokyo.3",
        "google.osaka.1",
        "google.osaka.2",
        "google.osaka.3",
        "google.seoul.1",
        "google.seoul.2",
        "google.seoul.3",
        "google.mumbai.1",
        "google.mumbai.2",
        "google.mumbai.3",
        "google.singapore.1",
        "google.singapore.2",
        "google.singapore.3",
        "google.jakarta.1",
        "google.jakarta.2",
        "google.jakarta.3",
        "google.sydney.1",
        "google.sydney.2",
        "google.sydney.3"
    };

    const int num_google_zones = sizeof(google_zones) / sizeof(char*);

    for ( int i = 0; i < num_google_zones; ++i )
    {
        printf( "%s,%s\n", google_zones[i], google_datacenters[i] );
    }

    printf( "autodetect: %d google zones\n", num_google_zones );

    const char * datacenter_gcp = next_autodetect_gcp( num_google_zones, google_zones, google_datacenters );

    if ( datacenter_gcp != NULL )
    {
        return datacenter_gcp;
    }

    // aws

    // ...

    // could not autodetect

    printf( "autodetect: could not autodetect datacenter\n" );

    return NULL;
}

int main()
{
    if ( next_init( NULL, NULL ) != NEXT_OK )
    {
        printf( "error: failed to initialize network next\n" );
    }

    const char * datacenter = next_autodetect_datacenter();

    printf( "\ndatacenter is \"%s\"\n\n", datacenter );

    next_term();

    fflush( stdout );

    return 0;
}
