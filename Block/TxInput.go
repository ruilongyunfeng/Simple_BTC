package Block

import "bytes"

type TxInput struct {
	txHash []byte

	voutindex int

	signature []byte

	publicKey []byte
}

//判断input 合法
func (txInput *TxInput)UnLockRipemd160Hash(ripemd160Hash []byte) bool{

	publicKey := Ripemd160Hash(txInput.publicKey)

	return bytes.Compare(publicKey,ripemd160Hash) == 0
}