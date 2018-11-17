package main

import (
	"fmt"
	"BlockTest/1-block/Block"
)

func main() {
	blockchain := Block.CreateBlockchainWithGenesisBlock("myself","123")

	fmt.Println(blockchain)
}