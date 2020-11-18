package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/networknext/backend/modules/core"
	"math"
	"os"
)

const InitRequestMagic = uint32(0x9083708f)
const InitRequestVersion = 0
const NonceBytes = 24
const InitResponseVersion = 0
const UpdateRequestVersion = 0
const UpdateResponseVersion = 0
const MaxRelayAddressLength = 256
const RelayTokenBytes = 32
const MaxRelays = 5

func ReadUint32(data []byte, index *int, value *uint32) bool {
	if *index+4 > len(data) {
		return false
	}
	*value = binary.LittleEndian.Uint32(data[*index:])
	*index += 4
	return true
}

func ReadUint64(data []byte, index *int, value *uint64) bool {
	if *index+8 > len(data) {
		return false
	}
	*value = binary.LittleEndian.Uint64(data[*index:])
	*index += 8
	return true
}

func ReadFloat32(data []byte, index *int, value *float32) bool {
	var int_value uint32
	if !ReadUint32(data, index, &int_value) {
		return false
	}
	*value = math.Float32frombits(int_value)
	return true
}

func ReadString(data []byte, index *int, value *string, maxStringLength uint32) bool {
	var stringLength uint32
	if !ReadUint32(data, index, &stringLength) {
		return false
	}
	if stringLength > maxStringLength {
		return false
	}
	if *index+int(stringLength) > len(data) {
		return false
	}
	stringData := make([]byte, stringLength)
	for i := uint32(0); i < stringLength; i++ {
		stringData[i] = data[*index]
		*index += 1
	}
	*value = string(stringData)
	return true
}

func ReadBytes(data []byte, index *int, value *[]byte, bytes uint32) bool {
	if *index+int(bytes) > len(data) {
		return false
	}
	*value = make([]byte, bytes)
	for i := uint32(0); i < bytes; i++ {
		(*value)[i] = data[*index]
		*index += 1
	}
	return true
}

func WriteUint32(data []byte, index *int, value uint32) {
	binary.LittleEndian.PutUint32(data[*index:], value)
	*index += 4
}

func WriteUint64(data []byte, index *int, value uint64) {
	binary.LittleEndian.PutUint64(data[*index:], value)
	*index += 8
}

func WriteString(data []byte, index *int, value string, maxStringLength uint32) {
	stringLength := uint32(len(value))
	if stringLength > maxStringLength {
		panic("string is too long!\n")
	}
	binary.LittleEndian.PutUint32(data[*index:], stringLength)
	*index += 4
	for i := 0; i < int(stringLength); i++ {
		data[*index] = value[i]
		*index++
	}
}

func WriteBytes(data []byte, index *int, value []byte, numBytes int) {
	for i := 0; i < numBytes; i++ {
		data[*index] = value[i]
		*index++
	}
}

func main() {

	fmt.Printf("\nNetwork Next Mock Relay\n")

	fmt.Printf("\nEnvironment:\n\n")

	relayAddressEnv := os.Getenv("RELAY_ADDRESS")
	if relayAddressEnv == "" {
		fmt.Printf("error: RELAY_ADDRESS is not set\n\n")
		os.Exit(1)
	}

	fmt.Printf("    relay address is \"%s\"\n", relayAddressEnv)

	relayAddress := core.ParseAddress(relayAddressEnv)
	if relayAddress == nil {
		fmt.Printf("error: failed to parse RELAY_ADDRESS\n\n")
		os.Exit(1)
	}

	relayPort := relayAddress.Port

	fmt.Printf("    relay bind port is %d\n", relayPort)

	relayPrivateKeyEnv := os.Getenv("RELAY_PRIVATE_KEY")
	if relayPrivateKeyEnv == "" {
		fmt.Printf("error: RELAY_PRIVATE_KEY is not set\n\n")
		os.Exit(1)
	}

	relayPrivateKey, err := base64.StdEncoding.DecodeString(relayPrivateKeyEnv)
	if err != nil {
		fmt.Printf("error: could not parse RELAY_PRIVATE_KEY as base64\n\n")
		os.Exit(1)
	}

	_ = relayPrivateKey

	fmt.Printf("    relay private key is \"%s\"\n", relayPrivateKeyEnv)

	relayPublicKeyEnv := os.Getenv("RELAY_PUBLIC_KEY")
	if relayPublicKeyEnv == "" {
		fmt.Printf("error: RELAY_PUBLIC_KEY is not set\n\n")
		os.Exit(1)
	}

	relayPublicKey, err := base64.StdEncoding.DecodeString(relayPublicKeyEnv)
	if err != nil {
		fmt.Printf("error: could not parse RELAY_PUBLIC_KEY as base64\n\n")
		os.Exit(1)
	}

	_ = relayPublicKey

	fmt.Printf("    relay public key is \"%s\"\n", relayPublicKeyEnv)

	relayRouterPublicKeyEnv := os.Getenv("RELAY_ROUTER_PUBLIC_KEY")
	if relayRouterPublicKeyEnv == "" {
		fmt.Printf("error: RELAY_ROUTER_PUBLIC_KEY is not set\n\n")
		os.Exit(1)
	}

	relayRouterPublicKey, err := base64.StdEncoding.DecodeString(relayRouterPublicKeyEnv)
	if err != nil {
		fmt.Printf("error: could not parse RELAY_ROUTER_PUBLIC_KEY as base64\n\n")
		os.Exit(1)
	}

	_ = relayRouterPublicKey

	fmt.Printf("    relay router public key is \"%s\"\n", relayRouterPublicKeyEnv)

	relayBackendHostnameEnv := os.Getenv("RELAY_BACKEND_HOSTNAME")
	if relayBackendHostnameEnv == "" {
		fmt.Printf("error: RELAY_BACKEND_HOSTNAME is not set\n\n")
		os.Exit(1)
	}

	fmt.Printf("    relay backend hostname is \"%s\"\n", relayBackendHostnameEnv)

	// write init data

	initData := make([]byte, 1024)

	index := 0

	WriteUint32(initData, &index, InitRequestMagic)
	WriteUint32(initData, &index, InitRequestVersion)

	nonce := make([]byte, NonceBytes)
	core.RandomBytes(nonce)
	WriteBytes(initData, &index, nonce, NonceBytes)

	WriteString(initData, &index, relayAddressEnv, MaxRelayAddressLength)

	fmt.Printf("\n(wrote %d bytes)\n", index)

	// todo: write relay token to init data
	/*
	   uint8_t * q = p;

	   relay_write_bytes( &p, relay_token, RELAY_TOKEN_BYTES );
	*/

	// encrypt init data with relay private key (what part is being encrypted exactly, and why? I forget...)
	/*
	   int encrypt_length = int( p - q );

	   if ( crypto_box_easy( q, q, encrypt_length, nonce, router_public_key, relay_private_key ) != 0 )
	   {
	       return RELAY_ERROR;
	   }

	   int init_length = (int) ( p - init_data ) + encrypt_length + crypto_box_MACBYTES;
	*/

	// todo: send the init request to the backend

	/*
	   struct curl_slist * slist = curl_slist_append( NULL, "Content-Type:application/octet-stream" );

	   curl_buffer_t init_response_buffer;
	   init_response_buffer.size = 0;
	   init_response_buffer.max_size = 1024;
	   init_response_buffer.data = (uint8_t*) alloca( init_response_buffer.max_size );

	   char init_url[1024];
	   sprintf( init_url, "%s/relay_init", hostname );

	   curl_easy_setopt( curl, CURLOPT_BUFFERSIZE, 102400L );
	   curl_easy_setopt( curl, CURLOPT_URL, init_url );
	   curl_easy_setopt( curl, CURLOPT_NOPROGRESS, 1L );
	   curl_easy_setopt( curl, CURLOPT_POSTFIELDS, init_data );
	   curl_easy_setopt( curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)init_length );
	   curl_easy_setopt( curl, CURLOPT_HTTPHEADER, slist );
	   curl_easy_setopt( curl, CURLOPT_USERAGENT, "network next relay" );
	   curl_easy_setopt( curl, CURLOPT_MAXREDIRS, 50L );
	   curl_easy_setopt( curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS );
	   curl_easy_setopt( curl, CURLOPT_TCP_KEEPALIVE, 1L );
	   curl_easy_setopt( curl, CURLOPT_TIMEOUT_MS, long( 1000 ) );
	   curl_easy_setopt( curl, CURLOPT_WRITEDATA, &init_response_buffer );
	   curl_easy_setopt( curl, CURLOPT_WRITEFUNCTION, &curl_buffer_write_function );

	   CURLcode ret = curl_easy_perform( curl );

	   curl_slist_free_all( slist );
	   slist = NULL;
	*/

	// todo: check the http request response

	/*
	   if ( ret != 0 )
	   {
	       return RELAY_ERROR;
	   }

	   long code;
	   curl_easy_getinfo( curl, CURLINFO_RESPONSE_CODE, &code );
	   if ( code != 200 )
	   {
	       return RELAY_ERROR;
	   }
	*/

	/*
	   if ( init_response_buffer.size < 4 )
	   {
	       relay_printf( "\nerror: bad relay init response size. too small to have valid data (%d)\n\n", init_response_buffer.size );
	       return RELAY_ERROR;
	   }

	   const uint8_t * r = init_response_buffer.data;

	   uint32_t version = relay_read_uint32( &r );

	   const uint32_t init_response_version = 0;

	   if ( version != init_response_version )
	   {
	       relay_printf( "\nerror: bad relay init response version. expected %d, got %d\n\n", init_response_version, version );
	       return RELAY_ERROR;
	   }

	   if ( init_response_buffer.size != 4 + 8 + RELAY_TOKEN_BYTES )
	   {
	       relay_printf( "\nerror: bad relay init response size. expected %d bytes, got %d\n\n", RELAY_TOKEN_BYTES, init_response_buffer.size );
	       return RELAY_ERROR;
	   }

	   *router_timestamp = relay_read_uint64( &r );

	   memcpy( relay_token, init_response_buffer.data + 4 + 8, RELAY_TOKEN_BYTES );

	   return RELAY_OK;
	*/

	// ---------------------------------------------------------

	// loop and update the relay

	fmt.Printf("\n")
}

/*
int relay_update( CURL * curl, const char * hostname, const uint8_t * relay_token, const char * relay_address, uint8_t * update_response_memory, relay_t * relay, bool shutdown )
{
    // build update data

    uint32_t update_version = 0;

    uint8_t update_data[10*1024 + 8 + 8 + 8 + 1]; // + 8 for the session count, + 8 for the bytes sent counter, + 8 for the bytes received counter, + 1 for the shutdown flag

    uint8_t * p = update_data;
    relay_write_uint32( &p, update_version );
    relay_write_string( &p, relay_address, 256 );
    relay_write_bytes( &p, relay_token, RELAY_TOKEN_BYTES );

    relay_platform_mutex_acquire( relay->mutex );
    relay_stats_t stats;
    relay_manager_get_stats( relay->relay_manager, &stats );
    relay_platform_mutex_release( relay->mutex );

    relay_write_uint32( &p, stats.num_relays );
    for ( int i = 0; i < stats.num_relays; ++i )
    {
        relay_write_uint64( &p, stats.relay_ids[i] );
        relay_write_float32( &p, stats.relay_rtt[i] );
        relay_write_float32( &p, stats.relay_jitter[i] );
        relay_write_float32( &p, stats.relay_packet_loss[i] );
    }

    relay_write_uint64(&p, relay->sessions->size());
    relay_write_uint64(&p, relay->bytes_sent.load());
    relay->bytes_sent.store(0);
    relay_write_uint64(&p, relay->bytes_received.load());
    relay->bytes_received.store(0);
    relay_write_uint8(&p, shutdown);
    relay_write_float64(&p, 0.00); // cpu usage
    relay_write_float64(&p, 0.00); // memory usage
    relay_write_string(&p, "1.0.0", sizeof("1.0.0")); // relay version

    int update_data_length = (int) ( p - update_data );

    // post it to backend

    struct curl_slist * slist = curl_slist_append( NULL, "Content-Type:application/octet-stream" );

    curl_buffer_t update_response_buffer;
    update_response_buffer.size = 0;
    update_response_buffer.max_size = RESPONSE_MAX_BYTES;
    update_response_buffer.data = (uint8_t*) update_response_memory;

    char update_url[1024];
    sprintf( update_url, "%s/relay_update", hostname );

    curl_easy_setopt( curl, CURLOPT_BUFFERSIZE, 102400L );
    curl_easy_setopt( curl, CURLOPT_URL, update_url );
    curl_easy_setopt( curl, CURLOPT_NOPROGRESS, 1L );
    curl_easy_setopt( curl, CURLOPT_POSTFIELDS, update_data );
    curl_easy_setopt( curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)update_data_length );
    curl_easy_setopt( curl, CURLOPT_HTTPHEADER, slist );
    curl_easy_setopt( curl, CURLOPT_USERAGENT, "network next relay" );
    curl_easy_setopt( curl, CURLOPT_MAXREDIRS, 50L );
    curl_easy_setopt( curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS );
    curl_easy_setopt( curl, CURLOPT_TCP_KEEPALIVE, 1L );
    curl_easy_setopt( curl, CURLOPT_TIMEOUT_MS, long( 1000 ) );
    curl_easy_setopt( curl, CURLOPT_WRITEDATA, &update_response_buffer );
    curl_easy_setopt( curl, CURLOPT_WRITEFUNCTION, &curl_buffer_write_function );

    CURLcode ret = curl_easy_perform( curl );

    curl_slist_free_all( slist );
    slist = NULL;

    if ( ret != 0 )
    {
        // relay_printf( "\nerror: could not post relay update\n\n" );
        return RELAY_ERROR;
    }

    long code;
    curl_easy_getinfo( curl, CURLINFO_RESPONSE_CODE, &code );
    if ( code != 200 )
    {
        // relay_printf( "\nerror: relay update response was %d, expected 200\n\n", int(code) );
        return RELAY_ERROR;
    }

    // parse update response

    const uint8_t * q = update_response_buffer.data;

    uint32_t version = relay_read_uint32( &q );

    const uint32_t update_response_version = 0;

    if ( version != update_response_version )
    {
        // relay_printf( "\nerror: bad relay update response version. expected %d, got %d\n\n", update_response_version, version );
        return RELAY_ERROR;
    }

    uint64_t timestamp = relay_read_uint64( &q );
    (void) timestamp;

    uint32_t num_relays = relay_read_uint32( &q );

    if ( num_relays > MAX_RELAYS )
    {
        // relay_printf( "\nerror: too many relays to ping. max is %d, got %d\n\n", MAX_RELAYS, num_relays );
        return RELAY_ERROR;
    }

    bool error = false;

    struct relay_ping_data_t
    {
        uint64_t id;
        relay_address_t address;
    };

    relay_ping_data_t relay_ping_data[MAX_RELAYS];

    for ( uint32_t i = 0; i < num_relays; ++i )
    {
        char address_string[RELAY_MAX_ADDRESS_STRING_LENGTH];
        relay_ping_data[i].id = relay_read_uint64( &q );
        relay_read_string( &q, address_string, RELAY_MAX_ADDRESS_STRING_LENGTH );
        if ( relay_address_parse( &relay_ping_data[i].address, address_string ) != RELAY_OK )
        {
            error = true;
            break;
        }
    }

    if ( error )
    {
        // relay_printf( "\nerror: error while reading set of relays to ping in update response\n\n" );
        return RELAY_ERROR;
    }

    relay_platform_mutex_acquire( relay->mutex );
    relay->num_relays = num_relays;
    for ( int i = 0; i < int(num_relays); ++i )
    {
        relay->relay_ids[i] = relay_ping_data[i].id;
        relay->relay_addresses[i] = relay_ping_data[i].address;
    }
    relay->relays_dirty = true;
    relay_platform_mutex_release( relay->mutex );

    return RELAY_OK;
}
*/
