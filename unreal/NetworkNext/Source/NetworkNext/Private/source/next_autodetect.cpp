/*
	Network Next. Copyright 2017 - 2025 Network Next, Inc.  
	Licensed under the Network Next Source Available License 1.0
*/

#include "next_autodetect.h"

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

#include "next_config.h"

#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

#if defined(_WIN32) || defined(_WIN64)
#define strtok_r strtok_s
#endif

bool next_autodetect_google( char * output, size_t output_size )
{
	FILE * file;
	char buffer[1024*10];

	// are we running in google cloud?
#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	file = popen( "/bin/ls /usr/bin | grep google_ 2>/dev/null", "r" );
	if ( file == NULL )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run ls" );
		return false;
	}

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	file = _popen( "dir \"C:\\Program Files (x86)\\Google\\Cloud SDK\\google-cloud-sdk\\bin\" | findstr gcloud", "r" );
	if ( file == NULL )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run dir" );
		return false;
	}

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	bool in_gcp = false;
	while ( fgets( buffer, sizeof(buffer), file ) != NULL )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: running in google cloud" );
		in_gcp = true;
		break;
	}

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	pclose( file );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	_pclose( file );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	// we are not running in google cloud :(

	if ( !in_gcp )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: not in google cloud" );
		return false;
	}

	// we are running in google cloud, which zone are we in?

	char zone[256];
	zone[0] = '\0';

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	file = popen( "curl -s \"http://metadata.google.internal/computeMetadata/v1/instance/zone\" -H \"Metadata-Flavor: Google\" --max-time 10 -vs 2>/dev/null", "r" );
	if ( !file )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run curl" );
		return false;
	}

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	file = _popen( "powershell Invoke-RestMethod -Uri http://metadata.google.internal/computeMetadata/v1/instance/zone -Headers @{'Metadata-Flavor' = 'Google'} -TimeoutSec 10", "r" );
	if ( !file )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run powershell Invoke-RestMethod" );
		return false;
	}

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	while ( fgets( buffer, sizeof(buffer), file ) != NULL )
	{
		size_t length = strlen( buffer );
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
		size_t index = length - 1;
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

		next_copy_string( zone, buffer + index + 1, sizeof(zone) );

		size_t zone_length = strlen(zone);
		index = zone_length - 1;
		while ( index > 0 && ( zone[index] == '\n' || zone[index] == '\r' ) )
		{
			zone[index] = '\0';
			index--;
		}

		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: google zone is \"%s\"", zone );

		break;
	}

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	pclose( file );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	_pclose( file );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	// we couldn't work out which zone we are in :(

	if ( zone[0] == '\0' )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not detect google zone" );
		return false;
	}

	// look up google zone -> network next datacenter via mapping in google cloud storage "google.txt" file

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	char cmd[1024];
	snprintf( cmd, sizeof(cmd), "curl -s \"https://storage.googleapis.com/%s/google.txt?ts=%x\" --max-time 10 -vs 2>/dev/null", NEXT_CONFIG_BUCKET_NAME, uint32_t(time(NULL)) );
	file = popen( cmd, "r" );
	if ( !file )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run curl" );
		return false;
	}

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	char cmd[1024];
	snprintf( cmd, sizeof(cmd), "powershell Invoke-RestMethod -Uri \"https://storage.googleapis.com/%s/google.txt?ts=%x\" -TimeoutSec 10", NEXT_CONFIG_BUCKET_NAME, uint32_t(time(NULL)) );
	file = _popen( cmd, "r" );
	if ( !file )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run powershell Invoke-RestMethod" );
		return false;
	}

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	bool found = false;

	while ( fgets( buffer, sizeof(buffer), file ) != NULL )
	{
		const char * separators = ",\n\r";
		char * rest = buffer;
		char * google_zone = strtok_r( buffer, separators, &rest );
		if ( google_zone == NULL )
		{
			continue;
		}

		char * google_datacenter = strtok_r( NULL, separators, &rest );
		if ( google_datacenter == NULL )
		{
			continue;
		}

		if ( strcmp( zone, google_zone ) == 0 )
		{
			next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: \"%s\" -> \"%s\"", zone, google_datacenter );
			next_copy_string( output, google_datacenter, output_size );
			found = true;
			break;
		}
	}

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	pclose( file );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	_pclose( file );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	return found;
}

bool next_autodetect_amazon( char * output, size_t output_size )
{
	FILE * file;
	char buffer[1024*10];

	// Get the AZID from instance metadata
	// This is necessary because only AZ IDs are the same across different AWS accounts
	// See https://docs.aws.amazon.com/ram/latest/userguide/working-with-az-ids.html for details

	char azid[256];
	azid[0] = '\0';

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	file = popen( "curl -s \"http://169.254.169.254/latest/meta-data/placement/availability-zone-id\" --max-time 2 -vs 2>/dev/null", "r" );
	if ( !file )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run curl" );
		return false;
	}

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	file = _popen ( "powershell Invoke-RestMethod -Uri http://169.254.169.254/latest/meta-data/placement/availability-zone-id -TimeoutSec 2", "r" );
	if ( !file )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run powershell Invoke-RestMethod" );
		return false;
	}

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	while ( fgets( buffer, sizeof(buffer), file ) != NULL )
	{
		if ( strstr( buffer, "-az" ) == NULL )
		{
			continue;
		}

		next_copy_string( azid, buffer, sizeof(azid) );

		size_t azid_length = strlen( azid );
		size_t index = azid_length - 1;
		while ( index > 0 && ( azid[index] == '\n' || azid[index] == '\r' ) )
		{
			azid[index] = '\0';
			index--;
		}

		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: azid is \"%s\"", azid );

		break;
	}

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	pclose( file );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	_pclose( file );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	// we are probably not in AWS :(

	if ( azid[0] == '\0' )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: not in AWS" );
		return false;
	}

	// look up AZID -> network next datacenter via mapping in google cloud storage "amazon.txt" file

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	char cmd[1024];
	snprintf( cmd, sizeof(cmd), "curl -s \"https://storage.googleapis.com/%s/amazon.txt?ts=%x\" --max-time 10 -vs 2>/dev/null", NEXT_CONFIG_BUCKET_NAME, uint32_t(time(NULL)) );
	file = popen( cmd, "r" );
	if ( !file )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run curl" );
		return false;
	}

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	char cmd[1024];
	snprintf( cmd, sizeof(cmd), "powershell Invoke-RestMethod -Uri \"https://storage.googleapis.com/%s/amazon.txt?ts=%x\" -TimeoutSec 10", NEXT_CONFIG_BUCKET_NAME, uint32_t(time(NULL)) );
	file = _popen ( cmd, "r" );
	if ( !file )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run powershell Invoke-RestMethod" );
		return false;
	}

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	bool found = false;

	while ( fgets( buffer, sizeof(buffer), file ) != NULL )
	{
		const char * separators = ",\n\r";
		char* rest = buffer;
		char * amazon_zone = strtok_r( buffer, separators, &rest );
		if ( amazon_zone == NULL )
		{
			continue;
		}

		char * amazon_datacenter = strtok_r( NULL, separators, &rest );
		if ( amazon_datacenter == NULL )
		{
			continue;
		}

		if ( strcmp( azid, amazon_zone ) == 0 )
		{
			next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: \"%s\" -> \"%s\"", azid, amazon_datacenter );
			next_copy_string( output, amazon_datacenter, output_size );
			found = true;
			break;
		}
	}

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	pclose( file );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	_pclose( file );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	return found;
}

// --------------------------------------------------------------------------------------------------------------

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

#include <sys/cdefs.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <ctype.h>
#include <err.h>
#include <netdb.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sysexits.h>
#include <unistd.h>

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

bool next_autodetect_datacenter( const char * input_datacenter, const char * public_address, char * output, size_t output_size )
{
#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC
	
	// we need curl to do any autodetect. bail if we don't have it

	next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: looking for curl" );

	int result = system( "curl -s >/dev/null 2>&1" );

	if ( result < 0 )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: curl not found" );
		return false;
	}

	next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: curl exists" );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

	// we need access to powershell and Invoke-RestMethod to do any autodetect. bail if we don't have it

	next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: looking for powershell Invoke-RestMethod" );

	int result = system( "powershell Invoke-RestMethod -? > NUL 2>&1" );

	if ( result > 0 )
	{
		next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: powershell Invoke-RestMethod not found" );
		return false;
	}

	next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: powershell Invoke-RestMethod exists" );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

	// google cloud

	bool google_result = next_autodetect_google( output, output_size );
	if ( google_result )
	{
		return true;
	}

	// amazon

	bool amazon_result = next_autodetect_amazon( output, output_size );
	if ( amazon_result )
	{
		return true;
	}

	(void) input_datacenter;
	(void) public_address;
	(void) output_size;

	return false;
}

#else // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

int next_autodetect_dummy;

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS
