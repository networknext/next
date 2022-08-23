package main

import (
	"fmt"
	"time"
	"hash/fnv"
	"encoding/binary"
)

func hashCounter(counter int64) []byte {
	hash := fnv.New64a()
	var inputValue [8]byte
	binary.LittleEndian.PutUint64(inputValue[:], uint64(counter))
	hash.Write(inputValue[:])
	hash.Write([]byte("don't worry. be happy."))
	hashValue := hash.Sum64()
	var result [8]byte
	binary.LittleEndian.PutUint64(result[:], uint64(hashValue))
	return result[:]
}

func main() {

	fmt.Printf("\nmagic test\n\n")

	for {

		secondsPerChange := int64(10)

		timestamp := time.Now().Unix()

		counter := timestamp / secondsPerChange

		fmt.Printf("counter = %d\n", counter)

		upcomingMagic := hashCounter(counter+2)
		currentMagic := hashCounter(counter+1)
		previousMagic := hashCounter(counter+0)

		fmt.Printf("magic values: %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x\n",
			upcomingMagic[0],
			upcomingMagic[1],
			upcomingMagic[2],
			upcomingMagic[3],
			upcomingMagic[4],
			upcomingMagic[5],
			upcomingMagic[6],
			upcomingMagic[7],
			currentMagic[0],
			currentMagic[1],
			currentMagic[2],
			currentMagic[3],
			currentMagic[4],
			currentMagic[5],
			currentMagic[6],
			currentMagic[7],
			previousMagic[0],
			previousMagic[1],
			previousMagic[2],
			previousMagic[3],
			previousMagic[4],
			previousMagic[5],
			previousMagic[6],
			previousMagic[7])

		time.Sleep(time.Second)
	}
}
