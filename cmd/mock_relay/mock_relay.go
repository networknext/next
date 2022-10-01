package main

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
    "bytes"
    "encoding/base64"
    "encoding/binary"
    "fmt"
    "io/ioutil"
    "math"
    "net"
    "net/http"
    "os"
    "strconv"
    "time"
    "unsafe"
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

func RandomBytes(buffer []byte) {
    C.randombytes_buf(unsafe.Pointer(&buffer[0]), C.size_t(len(buffer)))
}

func ParseAddress(input string) *net.UDPAddr {
    address := &net.UDPAddr{}
    ip_string, port_string, err := net.SplitHostPort(input)
    if err != nil {
        address.IP = net.ParseIP(input)
        address.Port = 0
        return address
    }
    address.IP = net.ParseIP(ip_string)
    address.Port, _ = strconv.Atoi(port_string)
    return address
}

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

func WriteUint8(data []byte, index *int, value byte) {
    data[*index] = value
    *index += 1
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

    relayAddress := ParseAddress(relayAddressEnv)
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
    RandomBytes(nonce)
    WriteBytes(initData, &index, nonce, NonceBytes)

    WriteString(initData, &index, relayAddress.String(), MaxRelayAddressLength)

    relayTokenIndex := index
    relayToken := make([]byte, RelayTokenBytes)
    RandomBytes(relayToken)
    WriteBytes(initData, &index, relayToken, RelayTokenBytes)

    err = Encrypt(relayPrivateKey, relayRouterPublicKey, nonce, initData[relayTokenIndex:], RelayTokenBytes)
    if err != nil {
        fmt.Printf("could not encrypt relay token data: %v\n", err)
    }

    initData = initData[:index+C.crypto_box_MACBYTES]

    // create and reuse one http client

    httpClient := http.Client{
        Timeout: time.Second * 10,
    }

    // init relay

    fmt.Printf("\nInitializing relay\n")

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

        if len(responseData) != 4+8+RelayTokenBytes {
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

    fmt.Printf("\nRelay initialized\n\n")

    // loop and update the relay

    iteration := 0

    for {

        time.Sleep(1 * time.Second)

        go func(updateNumber int) {

            // build update data

            updateData := make([]byte, 1024*10)

            index := 0

            WriteUint32(updateData, &index, UpdateRequestVersion)
            WriteString(updateData, &index, relayAddress.String(), MaxRelayAddressLength)
            WriteBytes(updateData, &index, relayToken, RelayTokenBytes)

            numRelays := uint32(320)
            WriteUint32(updateData, &index, numRelays)
            for i := 0; i < int(numRelays); i++ {
                WriteUint64(updateData, &index, 0)
                WriteUint32(updateData, &index, 0)
                WriteUint32(updateData, &index, 0)
                WriteUint32(updateData, &index, 0)
            }

            WriteUint64(updateData, &index, 0)
            WriteUint64(updateData, &index, 0)
            WriteUint64(updateData, &index, 0)
            WriteUint8(updateData, &index, 0)
            WriteUint64(updateData, &index, 0)
            WriteUint64(updateData, &index, 0)
            WriteString(updateData, &index, "1.0.0", 5)

            requestTime := time.Now()

            response, err := httpClient.Post(fmt.Sprintf("%s/relay_update", relayBackendHostnameEnv), "application/octet-stream", bytes.NewBuffer(updateData))
            if err != nil {
                fmt.Printf("update %d failed to post (%v)\n", updateNumber, time.Since(requestTime))
                return
            }

            responseData, err := ioutil.ReadAll(response.Body)
            if err != nil {
                fmt.Printf("update %d failed to read response (%v)\n", updateNumber, time.Since(requestTime))
                return
            }

            response.Body.Close()

            if response.StatusCode != 200 {
                fmt.Printf("update %d http status %d (%v)\n", updateNumber, response.StatusCode, time.Since(requestTime))
                return
            }

            fmt.Printf("update %d ok (%v)\n", updateNumber, time.Since(requestTime))

            _ = responseData

        }(iteration)

        iteration++
    }

    fmt.Printf("\n")
}
