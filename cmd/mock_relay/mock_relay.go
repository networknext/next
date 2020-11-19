package main

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"
	"bytes"
	"io/ioutil"
	"net/http"
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

func Encrypt(senderPrivateKey []byte, receiverPublicKey []byte, nonce []byte, buffer []byte, bytes int) error {
	result := C.crypto_box_easy((*C.uchar)(&buffer[0]),
		(*C.uchar)(&buffer[0]),
		C.ulonglong(bytes),
		(*C.uchar)(&nonce[0]),
		(*C.uchar)(&receiverPublicKey[0]),
		(*C.uchar)(&senderPrivateKey[0]))
	if result != 0 {
		return fmt.Errorf("failed to encrypt: result = %d", result)
	} else {
		return nil
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
	if relayPort == 0 {
		relayPort = 40000
	}
	relayAddress.Port = relayPort

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

	WriteString(initData, &index, relayAddress.String(), MaxRelayAddressLength)

	relayTokenIndex := index
	relayToken := make([]byte, RelayTokenBytes)
	core.RandomBytes(relayToken)
	WriteBytes(initData, &index, relayToken, RelayTokenBytes)

	err = Encrypt(relayPrivateKey, relayRouterPublicKey, nonce, initData[relayTokenIndex:], RelayTokenBytes)
	if err != nil {
		fmt.Printf("could not encrypt relay token data: %v\n", err)
	}

	initData = initData[:index + C.crypto_box_MACBYTES]

	// create and reuse one http client

	transport := &http.Transport{
		MaxConnsPerHost:     0,
		ForceAttemptHTTP2:   true,
        MaxIdleConns:        100000,
        MaxIdleConnsPerHost: 100000,
    }

	httpClient := http.Client{
		Transport: transport,
		Timeout: time.Second * 10,
	}

	// init relay

    fmt.Printf( "\nInitializing relay\n" );

	initialized := false

	for {

		time.Sleep(1 * time.Second)

		response, err := httpClient.Post(fmt.Sprintf("%s/relay_init", relayBackendHostnameEnv), "application/octet-stream", bytes.NewBuffer(initData))
		if err != nil {
			continue
		}

		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			continue
		}

		response.Body.Close()

		if response.StatusCode != 200 {
			continue
		}

	   	if len(responseData) != 4 + 8 + RelayTokenBytes {
	   		continue
	   	}

	   	index := 0

	   	var version uint32
		if !ReadUint32(responseData, &index, &version) {
			continue
		}

		if version != InitResponseVersion {
			continue
		}

		copy(relayToken, responseData[4+8:])

		initialized = true

		break
	}

	if !initialized {
		fmt.Printf("error: failed to init relay\n\n")
		os.Exit(1)
	}

    fmt.Printf( "\nRelay initialized\n" );

	// loop and update the relay

	for {

		time.Sleep(1 * time.Second)

		// todo

	}

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
