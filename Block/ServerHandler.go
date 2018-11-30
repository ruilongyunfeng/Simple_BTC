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

func sendTx(toAddress string, tx *Transaction) {

	payload := gobEncode(MessageTx{toAddress, tx})

	request := append(commandToBytes(COMMAND_TX), payload...)

	sendData(toAddress, request)
}

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
