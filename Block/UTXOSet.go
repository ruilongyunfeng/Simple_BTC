package Block

import (
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

type UTXOSet struct {
	blockChain *BlockChain
}

const utxoTableName  = "utxoTableName"

//重置数据库
func (ust *UTXOSet)ResetUTXOSet()  {
	err := ust.blockChain.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {
			err := tx.DeleteBucket([]byte(utxoTableName))
			if err != nil{
				log.Panic(err)
			}
		}

		b ,_= tx.CreateBucket([]byte(utxoTableName))

		if b != nil{
			txOutputsMap := ust.blockChain.FindUTXOMap()

			for keyHash,outs := range txOutputsMap{
				txHash,_ := hex.DecodeString(keyHash)
				b.Put(txHash,outs.Serialize())
			}
		}
		return nil
	})

	if err != nil{
		log.Panic(err)
	}
}

func (ust *UTXOSet) GetBalance(address string) int64 {
	utxos := ust.FindUTXOForAddress(address)

	var amount int64

	for _,utxo := range utxos  {
		amount += utxo.output.value
	}

	return amount
}

func (ust *UTXOSet) FindUTXOForAddress(address string) []*UTXO{
	var utxos []*UTXO

	ust.blockChain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil{
			c := b.Cursor()

			for k,v := c.First(); k != nil ; k,v = c.Next() {
				txOutputs := DeserializeTXOutputs(v)

				for _,utxo := range txOutputs.utxos{
					if utxo.output.UnLockScriptPubKeyWithAddress(address){
						utxos = append(utxos,utxo)
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

	for _,tx := range txs  {
		if tx.IsCoinbaseTransaction() == false{
			for _,vin := range tx.vins{
				pubKeyHash := Base58Decode([]byte(from))
				ripemd160Hash := pubKeyHash[1:len(pubKeyHash)-addressChecksumLen]
				if vin.UnLockRipemd160Hash(ripemd160Hash){
					key := hex.EncodeToString(vin.txHash)
					spentTxOutputs[key] = append(spentTxOutputs[key],vin.voutindex)
				}
			}
		}
	}//end spentTxOutputs

	for _,tx := range txs{
		work1:
		for index,out := range tx.vouts{
			if out.UnLockScriptPubKeyWithAddress(from){
				if len(spentTxOutputs)==0{
					utxo := &UTXO{tx.txHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				}else {
					for hash,indexArray := range spentTxOutputs{
						txHash := hex.EncodeToString(tx.txHash)

						if hash == txHash {

							for _,outIndex := range indexArray{
								if index == outIndex{
									continue work1
								}
								utxo := &UTXO{tx.txHash, index, out}
								unUTXOs = append(unUTXOs, utxo)
							}
						}else {
							utxo := &UTXO{tx.txHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}//end UnLock
		}
	}

	return unUTXOs
}

func (ust *UTXOSet) FindSpendableUTXOS(from string,amount int64,txs []*Transaction) (int64,map[string][]int)  {
	unPackageUTXOS := ust.FindUnPackageSpendableUTXOS(from,txs)

	spentableUTXO := make(map[string][]int)

	var money int64 = 0

	for _,UTXO := range unPackageUTXOS {
		money += UTXO.output.value
		txHash := hex.EncodeToString(UTXO.txHash)
		spentableUTXO[txHash] = append(spentableUTXO[txHash],UTXO.index)
		if money >= amount{
			return  money,spentableUTXO
		}
	}

	//不够
	ust.blockChain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {

			c := b.Cursor()

		UTXOWK:
			for k,v := c.First();k != nil;k,v = c.Next(){
				txOutputs := DeserializeTXOutputs(v)

				for _,utxo := range txOutputs.utxos {

					if utxo.output.UnLockScriptPubKeyWithAddress(from) {
						money += utxo.output.value
						txHash := hex.EncodeToString(utxo.txHash)
						spentableUTXO[txHash] = append(spentableUTXO[txHash],utxo.index)

						if money >= amount {
							break UTXOWK
						}
					}

				}
			}
		}
		return nil
	})

	if money < amount{
		log.Panic("余额不足......")
	}


	return  money,spentableUTXO
}

func (ust *UTXOSet) Update()  {

}