package Block

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile  = "Wallets_%s.dat"

type Wallets struct {
	walletsMap map[string]*Wallet
}

//创建钱包
func NewWallets(nodeID string)(*Wallets,error){

	walletFile := fmt.Sprintf(walletFile,nodeID)

	if _,err := os.Stat(walletFile);os.IsNotExist(err){
		wallets := &Wallets{}
		wallets.walletsMap = make(map[string]*Wallet)
		return wallets,err
	}

	fileContent,err := ioutil.ReadFile(walletFile)
	if err != nil{
		log.Panic(err)
	}

	var wallets Wallets

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil{
		log.Panic(err)
	}

	return &wallets,nil
}