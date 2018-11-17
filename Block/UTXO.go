package Block

type UTXO struct {
	txHash []byte
	index int
	output *TxOutput
}

