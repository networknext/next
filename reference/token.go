package reference

// todo: not today, friends

/*
import (
	"encoding/base64"
	"fmt"
	"net"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/transport"
)

var (
	RELAY_PUBLIC_KEY         = "8hUCRvzKh2aknL9RErM/Vj22+FGJW0tWMRz5KlHKryE="
	RELAY_PRIVATE_KEY        = "ZiCSchVFo6T5gJvbQfcwU7yfELsNJaYIC2laQm9DSuA="
	RELAY_ROUTER_PUBLIC_KEY  = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RELAY_ROUTER_PRIVATE_KEY = "ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M="
)

func main() {

	rPub, _ := base64.StdEncoding.DecodeString(RELAY_PUBLIC_KEY)
	rPriv, _ := base64.StdEncoding.DecodeString(RELAY_PRIVATE_KEY)
	bPub, _ := base64.StdEncoding.DecodeString(RELAY_ROUTER_PUBLIC_KEY)
	bPriv, _ := base64.StdEncoding.DecodeString(RELAY_ROUTER_PRIVATE_KEY)

	nonce := []byte("123456781234567812345678")
	fmt.Printf("nonce size: %v \n", len(nonce))
	data := []byte("12345678123456781234567812345678")
	token := crypto.Seal(data, nonce, bPub, rPriv)
	fmt.Printf("token size: %v \n", len(token))

	addr, _ := net.ResolveUDPAddr("udp", "100.1.1.1:40000")

	initRequest := transport.RelayInitRequest{
		Magic:          1,
		Version:        1,
		Nonce:          nonce,
		Address:        *addr,
		EncryptedToken: token,
		RelayVersion:   "0.0.0",
	}

	bin, err := initRequest.MarshalBinary()
	if err != nil {
		fmt.Printf("marshal error: %v", err)
		return
	}

	var nIR transport.RelayInitRequest
	err = nIR.UnmarshalBinary(bin)
	if err != nil {
		fmt.Printf("error unmarshaling error: %v", err)
		return
	}

	arr, ok := crypto.Open(nIR.EncryptedToken, nIR.Nonce, rPub, bPriv)
	fmt.Println(ok)
	fmt.Println(arr)
}
*/
