package Block

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

type UTXOSet struct {
	BlockChain *BlockChain
}

const utxoTableName = "utxoTableName"

//重置数据库
func (ust *UTXOSet) ResetUTXOSet() {
	err := ust.BlockChain.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {
			err := tx.DeleteBucket([]byte(utxoTableName))
			if err != nil {
				log.Panic(err)
			}
		}

		b, _ = tx.CreateBucket([]byte(utxoTableName))

		if b != nil {
			txOutputsMap := ust.BlockChain.FindUTXOMap()

			for keyHash, outs := range txOutputsMap {
				txHash, _ := hex.DecodeString(keyHash)
				b.Put(txHash, outs.Serialize())
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

func (ust *UTXOSet) GetBalance(address string) int64 {
	utxos := ust.FindUTXOForAddress(address)

	var amount int64

	for _, utxo := range utxos {
		amount += utxo.Output.Value
	}

	return amount
}

func (ust *UTXOSet) FindUTXOForAddress(address string) []*UTXO {
	var utxos []*UTXO

	ust.BlockChain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				txOutputs := DeserializeTXOutputs(v)

				for _, utxo := range txOutputs.Utxos {
					if utxo.Output.UnLockScriptPubKeyWithAddress(address) {
						utxos = append(utxos, utxo)
					}
				}
			}
		}

		return nil
	})

	return utxos
}

func (ust *UTXOSet) FindUnPackageSpendableUTXOS(from string, txs []*Transaction) []*UTXO {
	var unUTXOs []*UTXO

	spentTxOutputs := make(map[string][]int)

	for _, tx := range txs {
		if tx.IsCoinbaseTransaction() == false {
			for _, vin := range tx.Vins {
				pubKeyHash := Base58Decode([]byte(from))
				ripemd160Hash := pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
				if vin.UnLockRipemd160Hash(ripemd160Hash) {
					key := hex.EncodeToString(vin.TxHash)
					spentTxOutputs[key] = append(spentTxOutputs[key], vin.Voutindex)
				}
			}
		}
	} //end spentTxOutputs

	for _, tx := range txs {
	work1:
		for index, out := range tx.Vouts {
			if out.UnLockScriptPubKeyWithAddress(from) {
				if len(spentTxOutputs) == 0 {
					utxo := &UTXO{tx.TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash, indexArray := range spentTxOutputs {
						txHash := hex.EncodeToString(tx.TxHash)

						if hash == txHash {

							for _, outIndex := range indexArray {
								if index == outIndex {
									continue work1
								}
								utxo := &UTXO{tx.TxHash, index, out}
								unUTXOs = append(unUTXOs, utxo)
							}
						} else {
							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			} //end UnLock
		}
	}

	return unUTXOs
}

func (ust *UTXOSet) FindSpendableUTXOS(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {
	unPackageUTXOS := ust.FindUnPackageSpendableUTXOS(from, txs)

	spentableUTXO := make(map[string][]int)

	var money int64 = 0

	for _, UTXO := range unPackageUTXOS {
		money += UTXO.Output.Value
		txHash := hex.EncodeToString(UTXO.TxHash)
		spentableUTXO[txHash] = append(spentableUTXO[txHash], UTXO.Index)
		if money >= amount {
			return money, spentableUTXO
		}
	}

	ust.BlockChain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {

			c := b.Cursor()

		UTXOWK:
			for k, v := c.First(); k != nil; k, v = c.Next() {
				txOutputs := DeserializeTXOutputs(v)

				for _, utxo := range txOutputs.Utxos {

					if utxo.Output.UnLockScriptPubKeyWithAddress(from) {
						money += utxo.Output.Value
						txHash := hex.EncodeToString(utxo.TxHash)
						spentableUTXO[txHash] = append(spentableUTXO[txHash], utxo.Index)

						if money >= amount {
							break UTXOWK
						}
					}

				}
			}
		}
		return nil
	})

	if money < amount {
		log.Panic("余额不足......")
	}

	return money, spentableUTXO
}

func (ust *UTXOSet) Update() {
	// 最新的Block
	block := ust.BlockChain.Iterator().Next()

	ins := []*TxInput{}

	outsMap := make(map[string]*TxOutputs)

	// 找到所有我要删除的数据
	for _, tx := range block.Txs {

		for _, in := range tx.Vins {
			ins = append(ins, in)
		}
	}

	for _, tx := range block.Txs {

		utxos := []*UTXO{}

		for index, out := range tx.Vouts {

			isSpent := false

			for _, in := range ins {

				if in.Voutindex == index && bytes.Compare(tx.TxHash, in.TxHash) == 0 && bytes.Compare(out.Ripemd160Hash, Ripemd160Hash(in.PublicKey)) == 0 {

					isSpent = true
					continue
				}
			}

			if isSpent == false {
				utxo := &UTXO{tx.TxHash, index, out}
				utxos = append(utxos, utxo)
			}

		}

		if len(utxos) > 0 {
			txHash := hex.EncodeToString(tx.TxHash)
			outsMap[txHash] = &TxOutputs{utxos}
		}

	}

	err := ust.BlockChain.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {

			// 删除
			for _, in := range ins {

				txOutputsBytes := b.Get(in.TxHash)

				if len(txOutputsBytes) == 0 {
					continue
				}

				txOutputs := DeserializeTXOutputs(txOutputsBytes)

				fmt.Println(txOutputs)

				UTXOS := []*UTXO{}

				// 判断是否需要
				isNeedDelete := false

				for _, utxo := range txOutputs.Utxos {

					if in.Voutindex == utxo.Index && bytes.Compare(utxo.Output.Ripemd160Hash, Ripemd160Hash(in.PublicKey)) == 0 {

						isNeedDelete = true
					} else {
						UTXOS = append(UTXOS, utxo)
					}
				}

				if isNeedDelete {
					b.Delete(in.TxHash)
					if len(UTXOS) > 0 {

						preTXOutputs := outsMap[hex.EncodeToString(in.TxHash)]

						if preTXOutputs == nil {
							preTXOutputs = new(TxOutputs)
						}
						preTXOutputs.Utxos = append(preTXOutputs.Utxos, UTXOS...)

						outsMap[hex.EncodeToString(in.TxHash)] = preTXOutputs

					}
				}

			}

			// 新增
			for keyHash, outPuts := range outsMap {
				keyHashBytes, _ := hex.DecodeString(keyHash)
				b.Put(keyHashBytes, outPuts.Serialize())
			}

		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}
