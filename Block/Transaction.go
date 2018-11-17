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
	txHash []byte
	vins []*TxInput
	vouts []*TxOutput
}

func (tx *Transaction) Serialize() []byte {
	jsonBytes,err := json.Marshal(tx)
	if(err != nil){
		log.Panic(err)
	}

	return jsonBytes
}

//反序列
func Deserialize(txBytes []byte) *Transaction{
	var tx Transaction
	decoder := gob.NewDecoder(bytes.NewReader(txBytes))

	err := decoder.Decode(&tx)

	if err != nil{
		log.Panic(err)
	}

	return  &tx
}

//coinBaseTrans
func NewCoinbaseTransaction(address string) *Transaction{
	txInput := &TxInput{[]byte{},-1,nil,[]byte{}}

	txOutput := NewTxOutput(10,address)

	txCoinBase := &Transaction{[]byte{},[]*TxInput{txInput},[]*TxOutput{txOutput}}

	//txHash
	txCoinBase.HashTransaction()
	//coinBash 不需要签名
	return txCoinBase
}


//hash
func (tx *Transaction)HashTransaction(){
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)

	if err != nil{
		log.Panic(err)
	}

	resultBytes := bytes.Join([][]byte{IntToHex(time.Now().Unix()),result.Bytes()},[]byte{})

	hash := sha256.Sum256(resultBytes)

	tx.txHash = hash[:]
}

func (tx *Transaction) Hash() []byte {

	txCopy := tx

	txCopy.txHash = []byte{}

	hash := sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

//newTrans
func NewSimpleTransaction(from string,to string,amount int64,utxoSet *UTXOSet,txs []*Transaction,nodeID string) *Transaction{
	wallets,_ := NewWallets(nodeID)
	wallet := wallets.walletsMap[from]

	//查询余额
	money,spendableUTXODic := utxoSet.FindSpendableUTXOS(from,amount,txs)

	var txInputs []*TxInput
	var txOutputs []*TxOutput

	for txHash,indexArray := range spendableUTXODic{

		txHashBytes,_ := hex.DecodeString(txHash)
		for _,index := range indexArray{
			txInput := &TxInput{txHashBytes,index,nil,wallet.publicKey}
			txInputs = append(txInputs,txInput)
		}
	}

	txOutput := NewTxOutput(int64(amount),to)
	txOutputs = append(txOutputs,txOutput)

	txOutput = NewTxOutput(int64(money)-int64(amount),from)
	txOutputs = append(txOutputs,txOutput)

	tx := &Transaction{nil,txInputs,txOutputs}
	tx.HashTransaction()

	//sign
	utxoSet.blockChain.SignTransaction(tx,wallet.privateKey,txs)

	return tx
}

//sign
func (tx *Transaction) IsCoinbaseTransaction() bool {

	return len(tx.vins[0].txHash) == 0 && tx.vins[0].voutindex == -1
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbaseTransaction(){
		return
	}

	for _,vin := range tx.vins{
		if prevTXs[hex.EncodeToString(vin.txHash)].txHash == nil{
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for index,vin := range txCopy.vins{
		prevTx := prevTXs[hex.EncodeToString(vin.txHash)]
		txCopy.vins[index].signature = nil
		txCopy.vins[index].publicKey = prevTx.vouts[vin.voutindex].ripemd160Hash
		txCopy.txHash = txCopy.Hash()
		txCopy.vins[index].publicKey = nil

		r,s,err := ecdsa.Sign(rand.Reader,&privKey,txCopy.txHash)
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(),s.Bytes()...)

		tx.vins[index].signature = signature
	}
}

// 拷贝一份新的Transaction用于签名                                    T
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []*TxInput
	var outputs []*TxOutput

	for _, vin := range tx.vins {
		inputs = append(inputs, &TxInput{vin.txHash, vin.voutindex, nil, nil})
	}

	for _, vout := range tx.vouts {
		outputs = append(outputs, &TxOutput{vout.value, vout.ripemd160Hash})
	}

	txCopy := Transaction{tx.txHash, inputs, outputs}

	return txCopy
}
//verify
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbaseTransaction(){
		return true
	}

	for _,vin := range tx.vins{
		if prevTXs[hex.EncodeToString(vin.txHash)].txHash == nil{
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	curve := elliptic.P256()

	for index,vin := range txCopy.vins {
		prevTx := prevTXs[hex.EncodeToString(vin.txHash)]
		txCopy.vins[index].signature = nil
		txCopy.vins[index].publicKey = prevTx.vouts[vin.voutindex].ripemd160Hash
		txCopy.txHash = txCopy.Hash()
		txCopy.vins[index].publicKey = nil

		//r,s
		r := big.Int{}
		s := big.Int{}

		signLen := len(vin.signature)

		r.SetBytes(vin.signature[:(signLen/2)])
		s.SetBytes(vin.signature[(signLen/2):])

		x := big.Int{}
		y := big.Int{}

		keyLen := len(vin.publicKey)

		x.SetBytes(vin.signature[:(keyLen/2)])
		y.SetBytes(vin.signature[(keyLen/2):])

		rawPubKey := ecdsa.PublicKey{curve,&x,&y}

		if ecdsa.Verify(&rawPubKey,txCopy.txHash,&r,&s) == false{
			return false
		}
	}

	return true
}