package Block

import "bytes"

type TxInput struct {
	TxHash []byte

	Voutindex int

	Signature []byte

	PublicKey []byte
}

//判断input 合法
func (txInput *TxInput) UnLockRipemd160Hash(ripemd160Hash []byte) bool {

	publicKey := Ripemd160Hash(txInput.PublicKey)

	return bytes.Compare(publicKey, ripemd160Hash) == 0
}
