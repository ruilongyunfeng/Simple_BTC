package Block

import (
	"fmt"
	"os"
	"strconv"
)

/* *************************************************************
 * Copyright  2018 Bridge-ruijiezhi@163.com. All rights reserved.
 *
 * FileName: CLI_Function
 *
 * @Author: Bridge 2018/11/22 19:58
 *
 * @Version: 1.0
 * *************************************************************/

func (cli *CLI) createWallet(nodeID string) {

	wallets, _ := NewWallets(nodeID)

	wallets.CreateNewWallet(nodeID)

	fmt.Println(len(wallets.WalletsMap))
}

func (cli *CLI) createGenesisBlockchain(address string, nodeID string) {
	blockChain := CreateBlockchainWithGenesisBlock(address, nodeID)
	defer blockChain.DB.Close()

	utxoSet := &UTXOSet{blockChain}

	utxoSet.ResetUTXOSet()
}

func (cli *CLI) addressList(nodeID string) {
	fmt.Println("All wallet address:")

	wallets, _ := NewWallets(nodeID)

	for address := range wallets.WalletsMap {
		fmt.Println(address)
	}
}

func (cli *CLI) getBalance(address string, nodeID string) {
	fmt.Println("address : " + address)

	blockChain := BlockChainObject(nodeID)
	defer blockChain.DB.Close()

	utxoSet := &UTXOSet{blockChain}

	amount := utxoSet.GetBalance(address)

	fmt.Printf("%s have %d tokens.\n", address, amount)
}

func (cli *CLI) printChain(nodeID string) {
	blockChain := BlockChainObject(nodeID)

	defer blockChain.DB.Close()

	blockChain.PrintChain()
}

func (cli *CLI) send(from []string, to []string, amount []string, nodeID string, mineNow bool) {
	blockChain := BlockChainObject(nodeID)

	utxoSet := &UTXOSet{blockChain}
	defer blockChain.DB.Close()

	if mineNow {
		blockChain.MineNewBlock(from, to, amount, nodeID)
		utxoSet.Update()
	} else {
		value, _ := strconv.Atoi(amount[0])
		tx := NewSimpleTransaction(from[0], to[0], int64(value), utxoSet, []*Transaction{}, nodeID)

		sendTx(knowNodes[0], tx)
	}
}

func (cli *CLI) startNode(nodeID string, minerAdd string) {

	if minerAdd == "" || IsValidForAdress([]byte(minerAdd)) {
		fmt.Printf("start server : localhost:%s\n", nodeID)
		startServer(nodeID, minerAdd)
	} else {
		fmt.Println("The address is illegal!")
		os.Exit(0)
	}
}

func (cli *CLI) resetUTXOSet(nodeID string) {

	blockchain := BlockChainObject(nodeID)

	defer blockchain.DB.Close()

	utxoSet := &UTXOSet{blockchain}

	utxoSet.ResetUTXOSet()

}
