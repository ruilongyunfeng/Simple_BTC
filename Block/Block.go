package Block

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Height        int64
	PrevBlockHash []byte
	Txs           []*Transaction
	Timestamp     int64
	Hash          []byte
	Nonce         int64
}

func (block *Block) Serialize() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func DeSerializeBlock(blockBytes []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))

	err := decoder.Decode(&block)

	if err != nil {
		log.Panic(err)
	}

	return &block
}

func (block *Block) HashTransactions() []byte {
	var txs [][]byte

	for _, tx := range block.Txs {
		txs = append(txs, tx.Serialize())
	}
	merkleTree := NewMerkleTree(txs)

	return merkleTree.rootNode.data
}

func NewBlock(txs []*Transaction, height int64, prevBlockHash []byte) *Block {

	block := &Block{height, prevBlockHash, txs, time.Now().Unix(), nil, 0}

	pow := NewProofOfWork(block)

	hash, nonce := pow.Run()

	block.Hash = hash[:]

	block.Nonce = nonce

	return block
}

func CreateGenesisBlock(txs []*Transaction) *Block {
	return NewBlock(txs, 1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}
