package Block

import (
	"flag"
	"fmt"
	"log"
	"os"
)

/* *************************************************************
 * Copyright  2018 Bridge-ruijiezhi@163.com. All rights reserved.
 *
 * FileName: CLI
 *
 * @Author: Bridge 2018/11/6 15:35
 *
 * @Version: 1.0
 * *************************************************************/

type CLI struct {
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\taddressLists -- 输出所有钱包地址.")
	fmt.Println("\tcreateWallet -- 创建钱包.")
	fmt.Println("\tcreateBlockChain -address -- 交易数据.")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT -mine -- 交易明细.")
	fmt.Println("\tprintChain -- 输出区块信息.")
	fmt.Println("\tgetBalance -address -- balance.")
	fmt.Println("\tresetUTXO -- 重置.")
	fmt.Println("\tstartNode -miner ADDRESS -- 启动节点服务器，并且指定挖矿奖励的地址.")
}

func isValidArgs() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	isValidArgs()

	nodeID := os.Getenv("Node_ID")

	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!\n")
		os.Exit(1)
	}

	fmt.Printf("NODE_ID:%s\n", nodeID)

	resetUTXOCmd := flag.NewFlagSet("resetUTXO", flag.ExitOnError)

	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)

	createBlockChainCmd := flag.NewFlagSet("createBlockChain", flag.ExitOnError)

	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)

	startNodeCmd := flag.NewFlagSet("startNode", flag.ExitOnError)

	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	addressListCmd := flag.NewFlagSet("addressList", flag.ExitOnError)

	//send suffix
	flagFrom := sendCmd.String("from", "", "source address")
	flagTo := sendCmd.String("to", "", "destination address")
	flagAmount := sendCmd.String("amount", "", "transfer amount")
	flagMine := sendCmd.Bool("mine", false, "Mining Now")

	//startNode suffix
	flagMiner := startNodeCmd.String("miner", "", "Address of Miner")
	//blockChain suffix
	flagCreateBlockChainAddress := createBlockChainCmd.String("address", "", "Genesis Address")
	//getBalance suffix
	flagGetBalanceAddress := getBalanceCmd.String("adddress", "", "query balance")

	switch os.Args[1] {

	case "createWallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "resetUTXO":
		err := resetUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "createBlockChain":
		err := createBlockChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "printChain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "startNode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "getBalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "addressList":
		err := addressListCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		printUsage()
		os.Exit(1)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if resetUTXOCmd.Parsed() {
		cli.resetUTXOSet(nodeID)
	}

	if createBlockChainCmd.Parsed() {
		if IsValidForAdress([]byte(*flagCreateBlockChainAddress)) == false {
			fmt.Println("The address is illegal!")
			printUsage()
			os.Exit(1)
		}
		cli.createGenesisBlockchain(*flagCreateBlockChainAddress, nodeID)
	}

	if getBalanceCmd.Parsed() {
		if IsValidForAdress([]byte(*flagGetBalanceAddress)) == false {
			fmt.Println("The address is illegal!")
			printUsage()
			os.Exit(1)
		}

		cli.getBalance(*flagGetBalanceAddress, nodeID)
	}

	if startNodeCmd.Parsed() {
		cli.startNode(nodeID, *flagMiner)
	}

	if sendCmd.Parsed() {
		if *flagFrom == "" || *flagTo == "" || *flagAmount == "" {
			printUsage()
			os.Exit(1)
		}

		from := JsonToStringArray(*flagFrom)
		to := JsonToStringArray(*flagTo)

		for index, fromAdress := range from {
			if IsValidForAdress([]byte(fromAdress)) == false || IsValidForAdress([]byte(to[index])) == false {
				fmt.Println("Address is illegal!")
				printUsage()
				os.Exit(1)
			}
		}
		amount := JsonToStringArray(*flagAmount)

		cli.send(from, to, amount, nodeID, *flagMine)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if addressListCmd.Parsed() {
		cli.addressList(nodeID)
	}
}
