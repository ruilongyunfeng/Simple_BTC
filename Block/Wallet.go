package Block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
)

const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey

	PublicKey []byte
}

func IsValidForAdress(address []byte) bool {
	versionPublicKeyChecksumBytes := Base58Decode(address)

	checkSumBytes := versionPublicKeyChecksumBytes[len(versionPublicKeyChecksumBytes)-addressChecksumLen:]
	//fmt.Println(checkSumBytes)
	versionRipemd160 := versionPublicKeyChecksumBytes[:len(versionPublicKeyChecksumBytes)-addressChecksumLen]
	fmt.Println()
	checkBytes := CheckSum(versionRipemd160)
	//fmt.Println(checkBytes)
	if bytes.Compare(checkSumBytes, checkBytes) == 0 {
		return true
	}

	return false
}

func CheckSum(payload []byte) []byte {

	hash1 := sha256.Sum256(payload)
	hash2 := sha256.Sum256(hash1[:])

	return hash2[:addressChecksumLen]
}

func (w *Wallet) GetAddress() []byte {

	ripemd160Hash := Ripemd160Hash(w.PublicKey)

	versionRipemd160 := append([]byte{version}, ripemd160Hash...)

	checkSumBytes := CheckSum(versionRipemd160)

	resultBytes := append(versionRipemd160, checkSumBytes...)

	return Base58Encode(resultBytes)
}

func NewWallet() *Wallet {
	privateKey, publicKey := newKeyPair()

	return &Wallet{privateKey, publicKey}
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {

	curve := elliptic.P256()

	priKey, err := ecdsa.GenerateKey(curve, rand.Reader)

	if err != nil {
		log.Panic(err)
	}

	pubkey := append(priKey.PublicKey.X.Bytes(), priKey.PublicKey.Y.Bytes()...)

	return *priKey, pubkey
}
