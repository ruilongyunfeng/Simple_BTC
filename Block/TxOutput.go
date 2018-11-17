package Block

import "bytes"

type TxOutput struct {
	value int64
	ripemd160Hash []byte  //用户名
}

func (txOutput *TxOutput)Lock(address string)  {
	publickeyHash := Base58Decode([]byte(address))

	txOutput.ripemd160Hash = publickeyHash[1:len(publickeyHash)-4]
}

func NewTxOutput(value int64,address string) *TxOutput{
	txOutput := &TxOutput{value,nil}

	txOutput.Lock(address)

	return txOutput
}

func (txOutput *TxOutput) UnLockScriptPubKeyWithAddress(address string) bool{
	publicKeyHash := Base58Decode([]byte(address))
	hash160 := publicKeyHash[1:len(publicKeyHash) - 4]

	return bytes.Compare(txOutput.ripemd160Hash,hash160) == 0
}