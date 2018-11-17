package Block

import(
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"golang.org/x/crypto/ripemd160"
	"log"
)

func IntToHex(num int64) []byte{
	buff := new(bytes.Buffer)
	err := binary.Write(buff,binary.BigEndian,num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func ReverseBytes(input []byte){
	for i, j := 0, len(input)-1; i < j; i, j = i+1, j-1  {
		input[i],input[j] = input[j],input[i]
	}
}

func Ripemd160Hash(publicKey []byte) []byte {
	// 256

	hash256 := sha256.New()
	hash256.Write(publicKey)
	hash := hash256.Sum(nil)

	// 160
	rp160 := ripemd160.New()
	rp160.Write(hash)

	return rp160.Sum(nil)
}