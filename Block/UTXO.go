package Block

type UTXO struct {
	TxHash []byte
	Index  int
	Output *TxOutput
}
