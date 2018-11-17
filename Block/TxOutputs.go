package Block

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TxOutputs struct {
	utxos []*UTXO
}

// 将区块序列化成字节数组
func (txOutputs *TxOutputs) Serialize() []byte {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(txOutputs)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// 反序列化
func DeserializeTXOutputs(txOutputsBytes []byte) *TxOutputs {
	var txOutputs *TxOutputs
	decoder := gob.NewDecoder(bytes.NewReader(txOutputsBytes))
	err := decoder.Decode(&txOutputs)
	if err != nil {
		log.Panic(err)
	}

	return txOutputs
}
