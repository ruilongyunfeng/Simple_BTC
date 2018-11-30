package Block

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

/* *************************************************************
 * Copyright  2018 Bridge-ruijiezhi@163.com. All rights reserved.
 *
 * FileName: ServerHandler
 *
 * @Author: Bridge 2018/11/30 17:15
 *
 * @Version: 1.0
 * *************************************************************/

func handleVersion(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload MessageVersion

	dataBytes := request[COMMANDLENGTH:]

	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)

	if err != nil {
		log.Panic(err)
	}

	bestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if bestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	} else if bestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	}

	if !nodeIsKnown(payload.AddrFrom) {
		knowNodes = append(knowNodes, payload.AddrFrom)
	}
}

func handleGetblocks(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload MessageGetBlocks

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockHashes := bc.GetBlockHahes()

	sendInv(payload.AddressFrom, BLOCK_TYPE, blockHashes)
}

func handleInv(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload MessageInv

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == BLOCK_TYPE {
		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, BLOCK_TYPE, blockHash)

		if len(payload.Items) >= 1 {
			hashCacheArray = payload.Items[1:] //for what?
		}
	}

	if payload.Type == TX_TYPE {

		txHash := payload.Items[0]
		if memoryTxPool[hex.EncodeToString(txHash)] == nil {
			sendGetData(payload.AddrFrom, TX_TYPE, txHash)
		}
	}
}

func handleAddr(request []byte, bc *BlockChain) {

}

func handleGetData(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload MessageGetData

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == BLOCK_TYPE {

		block, err := bc.GetBlock([]byte(payload.Hash))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, block)
	}

	if payload.Type == TX_TYPE {

		tx := memoryTxPool[hex.EncodeToString(payload.Hash)]

		sendTx(payload.AddrFrom, tx)

	}
}

func handleBlock(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload MessageBlock

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockBytes := payload.Block

	block := DeSerializeBlock(blockBytes)

	//verify?
	fmt.Println("Recevied a new block!")
	bc.AddBlock(block)
	UTXOSet := &UTXOSet{bc}
	UTXOSet.Update()

	fmt.Printf("Added block %x\n", block.Hash)

	if len(hashCacheArray) > 0 {
		blockHash := hashCacheArray[0]
		sendGetData(payload.AddrFrom, BLOCK_TYPE, blockHash)

		hashCacheArray = hashCacheArray[1:]
	}
}

func handleTx(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload MessageTx

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	tx := payload.Tx

	memoryTxPool[hex.EncodeToString(tx.txHash)] = tx

	if nodeAddress == knowNodes[0] { //self
		for _, nodeAddr := range knowNodes { //transfer miner

			if nodeAddr != nodeAddress && nodeAddr != payload.AddrFrom {
				sendInv(nodeAddr, TX_TYPE, [][]byte{tx.txHash})
			}

		}
	}

	if len(minerAddress) > 0 {
		utxoSet := &UTXOSet{bc}
		txs := []*Transaction{tx}
		coinbaseTx := NewCoinbaseTransaction(minerAddress)
		txs = append(txs, coinbaseTx)

		_txs := []*Transaction{}

		for _, tx := range txs {

			if bc.VerifyTransaction(tx, _txs) != true {
				log.Panic("ERROR: Invalid transaction")
			}

			_txs = append(_txs, tx)
		}

		var block *Block

		bc.DB.View(func(tx *bolt.Tx) error {

			b := tx.Bucket([]byte(blockTableName))
			if b != nil {

				hash := b.Get([]byte("tip"))

				blockBytes := b.Get(hash)

				block = DeSerializeBlock(blockBytes)

			}

			return nil
		}) //end db

		//new block
		block = NewBlock(txs, block.Height+1, block.Hash)

		bc.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blockTableName))
			if b != nil {

				b.Put(block.Hash, block.Serialize())

				b.Put([]byte("tip"), block.Hash)

				bc.tip = block.Hash

			}
			return nil
		})

		utxoSet.Update()
		sendBlock(knowNodes[0], block.Serialize())
	}

}
