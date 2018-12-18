package Block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/big"
	"time"
)

type Transaction struct {
	TxHash []byte
	Vins   []*TxInput
	Vouts  []*TxOutput
}

func (tx *Transaction) Serialize() []byte {
	jsonBytes, err := json.Marshal(tx)
	if err != nil {
		log.Panic(err)
	}

	return jsonBytes
}

//反序列
func Deserialize(txBytes []byte) *Transaction {
	var tx Transaction
	decoder := gob.NewDecoder(bytes.NewReader(txBytes))

	err := decoder.Decode(&tx)

	if err != nil {
		log.Panic(err)
	}

	return &tx
}

//coinBaseTrans
func NewCoinbaseTransaction(address string) *Transaction {
	txInput := &TxInput{[]byte{}, -1, nil, []byte{}}

	txOutput := NewTxOutput(10, address)

	txCoinBase := &Transaction{[]byte{}, []*TxInput{txInput}, []*TxOutput{txOutput}}

	//txHash
	txCoinBase.HashTransaction()

	return txCoinBase
}

//hash
func (tx *Transaction) HashTransaction() {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	resultBytes := bytes.Join([][]byte{IntToHex(time.Now().Unix()), result.Bytes()}, []byte{})

	hash := sha256.Sum256(resultBytes)

	tx.TxHash = hash[:]
}

func (tx *Transaction) Hash() []byte {

	txCopy := tx

	txCopy.TxHash = []byte{}

	hash := sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

//newTrans
func NewSimpleTransaction(from string, to string, amount int64, utxoSet *UTXOSet, txs []*Transaction, nodeID string) *Transaction {
	wallets, _ := NewWallets(nodeID)
	wallet := wallets.WalletsMap[from]
	if wallet == nil {
		return nil
	}
	//查询余额
	money, spendableUTXODic := utxoSet.FindSpendableUTXOS(from, amount, txs)

	var txInputs []*TxInput
	var txOutputs []*TxOutput

	for txHash, indexArray := range spendableUTXODic {

		txHashBytes, _ := hex.DecodeString(txHash)
		for _, index := range indexArray {
			txInput := &TxInput{txHashBytes, index, nil, wallet.PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}

	txOutput := NewTxOutput(int64(amount), to)
	txOutputs = append(txOutputs, txOutput)

	txOutput = NewTxOutput(int64(money)-int64(amount), from)
	txOutputs = append(txOutputs, txOutput)

	tx := &Transaction{nil, txInputs, txOutputs}
	tx.HashTransaction()

	//sign
	utxoSet.BlockChain.SignTransaction(tx, wallet.PrivateKey, txs)

	return tx
}

//sign
func (tx *Transaction) IsCoinbaseTransaction() bool {

	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Voutindex == -1
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}

	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for index, vin := range txCopy.Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[index].Signature = nil
		txCopy.Vins[index].PublicKey = prevTx.Vouts[vin.Voutindex].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[index].PublicKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.TxHash)
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vins[index].Signature = signature
	}
}

// 拷贝一份新的Transaction用于签名                                    T
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []*TxInput
	var outputs []*TxOutput

	for _, vin := range tx.Vins {
		inputs = append(inputs, &TxInput{vin.TxHash, vin.Voutindex, nil, nil})
	}

	for _, vout := range tx.Vouts {
		outputs = append(outputs, &TxOutput{vout.Value, vout.Ripemd160Hash})
	}

	txCopy := Transaction{tx.TxHash, inputs, outputs}

	return txCopy
}

//verify
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbaseTransaction() {
		return true
	}

	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	curve := elliptic.P256()

	for index, vin := range tx.Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[index].Signature = nil
		txCopy.Vins[index].PublicKey = prevTx.Vouts[vin.Voutindex].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[index].PublicKey = nil

		//r,s
		r := big.Int{}
		s := big.Int{}

		signLen := len(vin.Signature)

		r.SetBytes(vin.Signature[:(signLen / 2)])
		s.SetBytes(vin.Signature[(signLen / 2):])

		x := big.Int{}
		y := big.Int{}

		keyLen := len(vin.PublicKey)

		x.SetBytes(vin.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.PublicKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}

		if ecdsa.Verify(&rawPubKey, txCopy.TxHash, &r, &s) == false {
			return false
		}
	}

	return true
}
