package Block

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
)

/* *************************************************************
 * Copyright  2018 Bridge-ruijiezhi@163.com. All rights reserved.
 *
 * FileName: Server.go
 *
 * @Author: Bridge 2018/11/30 14:59
 *
 * @Version: 1.0
 * *************************************************************/

func startServer(nodeID string, minerAdd string) {

	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)

	minerAddress = minerAdd

	ln, err := net.Listen(PROTOCOL, nodeAddress)

	if err != nil {
		log.Panic(err)
	}

	defer ln.Close()

	bc := BlockChainObject(nodeID)

	if nodeAddress != knowNodes[0] {
		sendVersion(knowNodes[0], bc)
	}

	for {
		// 收到的数据的格式是固定的，12字节+结构体字节数组
		// 接收客户端发送过来的数据
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}

		go handleConnection(conn, bc)

	}
}

func nodeIsKnown(addr string) bool {
	for _, node := range knowNodes {
		if node == addr {
			return true
		}
	}

	return false
}

func handleConnection(conn net.Conn, bc *BlockChain) {
	//read
	request, err := ioutil.ReadAll(conn)

	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Receive a Message:%s\n", request[:COMMANDLENGTH])

	//version
	command := bytesToCommand(request[:COMMANDLENGTH])

	switch command {
	case COMMAND_VERSION:
		handleVersion(request, bc)
	case COMMAND_ADDR:
		handleAddr(request, bc)
	case COMMAND_BLOCK:
		handleBlock(request, bc)
	case COMMAND_GETBLOCKS:
		handleGetblocks(request, bc)
	case COMMAND_GETDATA:
		handleGetData(request, bc)
	case COMMAND_INV:
		handleInv(request, bc)
	case COMMAND_TX:
		handleTx(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}
