package Block

import (
	"bytes"
	"io"
	"log"
	"net"
)

/* *************************************************************
 * Copyright  2018 Bridge-ruijiezhi@163.com. All rights reserved.
 *
 * FileName: ServerHandler
 *
 * @Author: Bridge 2018/11/30 15:00
 *
 * @Version: 1.0
 * *************************************************************/

func sendData(to string, data []byte) {

	conn, err := net.Dial(PROTOCOL, to)

	if err != nil {
		panic("error")
	}
	defer conn.Close()

	//transfer
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}

}

func sendVersion(toAddress string, bc *BlockChain) {
	bestHeight := bc.GetBestHeight()

	payload := gobEncode(MessageVersion{NODE_VERSION, bestHeight, nodeAddress})

	request := append(commandToBytes(COMMAND_VERSION), payload...)

	sendData(toAddress, request)
}

func sendTx(toAddress string, tx *Transaction) {

	payload := gobEncode(MessageTx{toAddress, tx})

	request := append(commandToBytes(COMMAND_TX), payload...)

	sendData(toAddress, request)
}

func sendGetBlocks(toAddress string) {

	payload := gobEncode(MessageGetBlocks{nodeAddress})

	request := append(commandToBytes(COMMAND_GETBLOCKS), payload...)

	sendData(toAddress, request)
}

func sendInv(toAddress string, messageType string, hashes [][]byte) {

	payload := gobEncode(MessageInv{nodeAddress, messageType, hashes})

	request := append(commandToBytes(COMMAND_INV), payload...)

	sendData(toAddress, request)
}

func sendGetData(toAddress string, messageType string, blockHash []byte) {
	payload := gobEncode(MessageGetData{nodeAddress, messageType, blockHash})

	request := append(commandToBytes(COMMAND_GETDATA), payload...)

	sendData(toAddress, request)
}

func sendBlock(toAddress string, block []byte) {
	payload := gobEncode(MessageBlock{nodeAddress, block})

	request := append(commandToBytes(COMMAND_BLOCK), payload...)

	sendData(toAddress, request)
}
