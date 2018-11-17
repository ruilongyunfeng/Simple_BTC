package Block

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"os"
)

type BlockChain struct {
	tip []byte
	DB *bolt.DB
}

const dbName = "blockchain_%s.db"

// 表的名字
const blockTableName = "blocks"

func DBExists(dbName string) bool{
	if _,err := os.Stat(dbName);os.IsNotExist(err) {
		return false
	}

	return true
}

func (bc *BlockChain) Iterator() *BlockchainIterator{
	return &BlockchainIterator{bc.tip,bc.DB}
}
func CreateBlockchainWithGenesisBlock(address string,nodeID string) *BlockChain{

	dbName := fmt.Sprint(dbName,nodeID)

	if DBExists(dbName){
		fmt.Println("创世区块已经存在。。。。")
		os.Exit(1)
	}

	fmt.Println("创建创世区块。。。。。")

	db,err := bolt.Open(dbName,0600,nil)

	if err != nil{
		log.Fatal(err)
	}

	var genesisHash []byte

	err = db.Update(func(tx *bolt.Tx) error {
		//创建数据库
		db,err := tx.CreateBucket([]byte(blockTableName))

		if err != nil {
			log.Panic(err)
		}

		if db != nil {
			txCoinbase := NewCoinbaseTransaction(address)

			genesisBlock := CreateGenesisBlock([]*Transaction{txCoinbase})

			err := db.Put(genesisBlock.Hash,genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = db.Put([]byte("l"),genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

			genesisHash = genesisBlock.Hash
		}
		return nil
	})

	return &BlockChain{genesisHash, db}
}

func (bc *BlockChain) AddBlockToBlockChain(txs []*Transaction){
	//获取区块
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		//获取表
		b := tx.Bucket([]byte(blockTableName))

		if b != nil{
			blockBytes := b.Get(bc.tip)
			//反序列
			blockHeader := DeSerializeBlock(blockBytes)

			//存新区块
			newBlock := NewBlock(txs,blockHeader.Height+1,blockHeader.Hash)

			err := b.Put(newBlock.Hash,newBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = b.Put([]byte("l"),newBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

			//更新blockChain
			bc.tip = newBlock.Hash
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

func (bc *BlockChain) FindUTXOMap() map[string]*TxOutputs {
	//Bridge
	return nil
}

func (bc *BlockChain) SignTransaction(tx *Transaction,privKey ecdsa.PrivateKey,txs []*Transaction) {
	if tx.IsCoinbaseTransaction(){
		return
	}

	prevTXs := make(map[string]Transaction)

	for _,vin := range tx.vins{
		prevTX,err := bc.FindTransaction(vin.txHash,txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.txHash)] = prevTX
	}

	tx.Sign(privKey,prevTXs)
}

func (bc *BlockChain) FindTransaction(txHash []byte,txs []*Transaction)(Transaction,error){

	for _,tx := range txs{
		if bytes.Compare(txHash,tx.txHash) == 0{
			return *tx,nil
		}
	}

	iterator := bc.Iterator()

	for  {
		block := iterator.Next()

		for _,tx := range block.Txs{
			if bytes.Compare(tx.txHash,txHash) == 0{
				return *tx,nil
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if big.NewInt(0).Cmp(&hashInt) == 0{
			break
		}
	}

	return Transaction{},nil
}