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
	"strconv"
	"time"
)

type BlockChain struct {
	tip []byte
	DB  *bolt.DB
}

const dbName = "blockchain_%s.db"

const blockTableName = "blocks"

func DBExists(dbName string) bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}

	return true
}

func (bc *BlockChain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.DB}
}
func CreateBlockchainWithGenesisBlock(address string, nodeID string) *BlockChain {

	fmt.Println(address)
	fmt.Println(nodeID)
	dbName := fmt.Sprintf(dbName, nodeID)

	if DBExists(dbName) {
		fmt.Println("创世区块已经存在。。。。")
		os.Exit(1)
	}

	fmt.Println("创建创世区块。。。。。")

	db, err := bolt.Open(dbName, 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	var genesisHash []byte

	err = db.Update(func(tx *bolt.Tx) error {
		//创建数据库
		db, err := tx.CreateBucket([]byte(blockTableName))

		if err != nil {
			log.Panic(err)
		}

		if db != nil {
			txCoinbase := NewCoinbaseTransaction(address)

			genesisBlock := CreateGenesisBlock([]*Transaction{txCoinbase})

			err := db.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = db.Put([]byte("tip"), genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

			genesisHash = genesisBlock.Hash
		}
		return nil
	})

	return &BlockChain{genesisHash, db}
}

func (bc *BlockChain) AddBlockToBlockChain(txs []*Transaction) {
	//获取区块
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		//获取表
		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			blockBytes := b.Get(bc.tip)
			//反序列
			blockHeader := DeSerializeBlock(blockBytes)

			//存新区块
			newBlock := NewBlock(txs, blockHeader.Height+1, blockHeader.Hash)

			err := b.Put(newBlock.Hash, newBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = b.Put([]byte("tip"), newBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

			bc.tip = newBlock.Hash
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

func BlockChainObject(nodeID string) *BlockChain {

	dbName := fmt.Sprintf(dbName, nodeID)

	if DBExists(dbName) == false {
		fmt.Println("DB is not exist!")
		//os.Exit(1)
		return CreateBlockchainWithGenesisBlock(minerAddress, nodeID)
	}

	db, err := bolt.Open(dbName, 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	var tip []byte

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			tip = b.Get([]byte("tip"))
		}

		return nil
	})

	return &BlockChain{tip, db}
}

func (bc *BlockChain) UnUTXOs(address string, txs []*Transaction) []*UTXO {
	var unUTXOs []*UTXO

	spentTxOutputs := make(map[string][]int)

	for _, tx := range txs {
		if tx.IsCoinbaseTransaction() == false {
			for _, in := range tx.Vins {
				publicKeyHash := Base58Decode([]byte(address))

				ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]

				if in.UnLockRipemd160Hash(ripemd160Hash) {
					key := hex.EncodeToString(in.TxHash)
					spentTxOutputs[key] = append(spentTxOutputs[key], in.Voutindex)
				}
			}
		}
	}

	for _, tx := range txs {
	Work:
		for index, out := range tx.Vouts {
			if out.UnLockScriptPubKeyWithAddress(address) {
				if len(spentTxOutputs) == 0 {
					utxo := &UTXO{tx.TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash, indexArray := range spentTxOutputs {
						txHashStr := hex.EncodeToString(tx.TxHash)

						if hash == txHashStr {
							var isUnSpentUTXO bool
							for _, outIndex := range indexArray {
								if index == outIndex {
									isUnSpentUTXO = true
									continue Work
								}

								if isUnSpentUTXO == false {
									utxo := &UTXO{tx.TxHash, index, out}
									unUTXOs = append(unUTXOs, utxo)
								}
							}
						} else {
							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}
	}

	blockIterator := bc.Iterator()

	for {
		block := blockIterator.Next()

		for i := len(block.Txs) - 1; i >= 0; i-- {
			tx := block.Txs[i]
			//ins
			if tx.IsCoinbaseTransaction() == false {
				for _, in := range tx.Vins {
					publicKeyHash := Base58Decode([]byte(address))

					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]

					if in.UnLockRipemd160Hash(ripemd160Hash) {
						key := hex.EncodeToString(in.TxHash)
						spentTxOutputs[key] = append(spentTxOutputs[key], in.Voutindex)
					}
				}
			}
			//outs
		workOut:
			for index, out := range tx.Vouts {
				if out.UnLockScriptPubKeyWithAddress(address) {
					if spentTxOutputs != nil {
						if len(spentTxOutputs) != 0 {
							var isSpentUTXO bool

							for txHash, indexArray := range spentTxOutputs {
								for _, i := range indexArray {
									if index == i && txHash == hex.EncodeToString(tx.TxHash) {
										isSpentUTXO = true
										continue workOut
									}
								}
							}

							if isSpentUTXO == false {
								utxo := &UTXO{tx.TxHash, index, out}
								unUTXOs = append(unUTXOs, utxo)
							}
						} else {
							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}

		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break //genesis
		}

	}

	return unUTXOs
}

func (bc *BlockChain) FindSpendableUTXOS(from string, amount int, txs []*Transaction) (int64, map[string][]int) {
	//Bridge
	utxos := bc.UnUTXOs(from, txs)

	spendableUtxo := make(map[string][]int)

	var value int64

	for _, utxo := range utxos {
		value = value + utxo.Output.Value
		hash := hex.EncodeToString(utxo.TxHash)

		spendableUtxo[hash] = append(spendableUtxo[hash], utxo.Index)

		if value >= int64(amount) {
			break
		} else {
			fmt.Printf("%s's fund is not enough\n", from)
			os.Exit(1)
		}
	}
	return value, spendableUtxo
}

func (bc *BlockChain) FindUTXOMap() map[string]*TxOutputs {
	//Bridge
	bcIterator := bc.Iterator()

	spendableUtxosMap := make(map[string][]*TxInput)

	utxoMaps := make(map[string]*TxOutputs)

	for {
		block := bcIterator.Next()

		for i := len(block.Txs) - 1; i >= 0; i-- {
			txOutputs := &TxOutputs{[]*UTXO{}}

			tx := block.Txs[i]

			if tx.IsCoinbaseTransaction() == false {
				for _, txInput := range tx.Vins {
					txHash := hex.EncodeToString(txInput.TxHash)
					spendableUtxosMap[txHash] = append(spendableUtxosMap[txHash], txInput)
				}
			}

			txHash := hex.EncodeToString(tx.TxHash)

			txInputs := spendableUtxosMap[txHash]

			if len(txInputs) > 0 {
			workOut:
				for index, out := range tx.Vouts {
					for _, in := range txInputs {
						outPublicKey := out.Ripemd160Hash
						inPublicKey := in.PublicKey

						if bytes.Compare(outPublicKey, Ripemd160Hash(inPublicKey)) == 0 {
							if index == in.Voutindex {
								continue workOut
							} else {
								utxo := &UTXO{tx.TxHash, index, out}
								txOutputs.Utxos = append(txOutputs.Utxos, utxo)
							}
						}
					}
				}
			} else { //no txInput
				for index, out := range tx.Vouts {
					utxo := &UTXO{tx.TxHash, index, out}
					txOutputs.Utxos = append(txOutputs.Utxos, utxo)
				}
			}

			utxoMaps[txHash] = txOutputs
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return utxoMaps
}

func (bc *BlockChain) MineNewBlock(from []string, to []string, amount []string, nodeID string) {
	utxoSet := &UTXOSet{bc}

	var txs []*Transaction

	for index, address := range from {
		value, _ := strconv.Atoi(amount[index])
		tx := NewSimpleTransaction(address, to[index], int64(value), utxoSet, txs, nodeID)
		txs = append(txs, tx)
	}

	tx := NewCoinbaseTransaction(from[0])
	txs = append(txs, tx)

	var block *Block

	bc.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			hash := b.Get([]byte("tip"))

			blockBytes := b.Get(hash)

			block = DeSerializeBlock(blockBytes)
		}

		return nil
	})

	_txs := []*Transaction{}

	for _, tx := range txs {
		if bc.VerifyTransaction(tx, _txs) != true {
			log.Panic("ERROR: Invalid transaction!")
		}

		_txs = append(_txs, tx)
	}

	block = NewBlock(txs, block.Height+1, block.Hash)

	//insert new block

	bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			b.Put(block.Hash, block.Serialize())

			b.Put([]byte("tip"), block.Hash)

			bc.tip = block.Hash
		}

		return nil
	})
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey, txs []*Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vins {
		prevTX, err := bc.FindTransaction(vin.TxHash, txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) FindTransaction(txHash []byte, txs []*Transaction) (Transaction, error) {

	for _, tx := range txs {
		if bytes.Compare(txHash, tx.TxHash) == 0 {
			return *tx, nil
		}
	}

	iterator := bc.Iterator()

	for {
		block := iterator.Next()

		for _, tx := range block.Txs {
			if bytes.Compare(tx.TxHash, txHash) == 0 {
				return *tx, nil
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}

	return Transaction{}, nil
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction, txs []*Transaction) bool {

	prevTxs := make(map[string]Transaction)

	for _, vin := range tx.Vins {
		prevTx, err := bc.FindTransaction(vin.TxHash, txs)

		if err != nil {
			log.Panic(err)
		}

		prevTxs[hex.EncodeToString(prevTx.TxHash)] = prevTx
	}

	return tx.Verify(prevTxs)
}

func (bc *BlockChain) GetBalance(address string) int64 {
	utxos := bc.UnUTXOs(address, []*Transaction{})

	var amount int64

	for _, utxo := range utxos {
		amount = amount + utxo.Output.Value
	}

	return amount
}

func (bc *BlockChain) GetBestHeight() int64 {
	block := bc.Iterator().Next()

	return block.Height
}

func (bc *BlockChain) GetBlockHahes() [][]byte {
	blockIterator := bc.Iterator()

	var blockHashs [][]byte

	for {
		block := blockIterator.Next()

		blockHashs = append(blockHashs, block.Hash)

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	return blockHashs
}

func (bc *BlockChain) GetBlock(blockHash []byte) ([]byte, error) {
	var blockBytes []byte

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			blockBytes = b.Get(blockHash)
		}

		return nil
	})

	return blockBytes, err
}

func (bc *BlockChain) AddBlock(block *Block) {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			blockExist := b.Get(block.Hash)

			if blockExist != nil {
				return nil
			}

			err := b.Put(block.Hash, block.Serialize())

			if err != nil {
				log.Panic(err)
			}

			blockHash := b.Get([]byte("tip"))

			blockBytes := b.Get(blockHash)

			blockInDB := DeSerializeBlock(blockBytes)

			if blockInDB.Height < block.Height {
				b.Put([]byte("tip"), block.Hash)
				bc.tip = block.Hash
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

func (bc *BlockChain) PrintChain() {
	fmt.Println("All BlockInfo:")

	bcIterator := bc.Iterator()

	for {
		block := bcIterator.Next()

		fmt.Printf("Height：%d\n", block.Height)
		fmt.Printf("PrevBlockHash：%x\n", block.PrevBlockHash)
		fmt.Printf("Timestamp：%s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash：%x\n", block.Hash)
		fmt.Printf("Nonce：%d\n", block.Nonce)
		fmt.Println("Txs:")

		for _, tx := range block.Txs {
			fmt.Printf("%x\n", tx.TxHash)
			fmt.Println("Vins:")

			for _, in := range tx.Vins {
				fmt.Printf("%x\n", in.TxHash)
				fmt.Printf("%d\n", in.Voutindex)
				fmt.Printf("%x\n", in.PublicKey)
			}

			fmt.Println("Vouts:")
			for _, out := range tx.Vouts {
				fmt.Printf("%d\n", out.Value)
				fmt.Printf("%x\n", out.Ripemd160Hash)
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
}
