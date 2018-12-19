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

const walletFile = "Wallets.dat"

type Wallets struct {
	WalletsMap map[string]*Wallet
}

//创建钱包
func NewWallets(nodeID string) (*Wallets, error) {

	//walletFile := fmt.Sprintf(walletFile, nodeID)

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		wallets := &Wallets{}
		wallets.WalletsMap = make(map[string]*Wallet)
		return wallets, err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	return &wallets, nil
}

func (w *Wallets) CreateNewWallet(nodeID string) {
	wallet := NewWallet()

	fmt.Printf("New Wallet Address: %s\n", wallet.GetAddress())

	w.WalletsMap[string(wallet.GetAddress())] = wallet
	w.SaveWallets(nodeID)
}

func (w *Wallets) SaveWallets(nodeID string) {
	//walletFile := fmt.Sprintf(walletFile, nodeID)

	var content bytes.Buffer

	gob.Register(elliptic.P256()) //TODO

	encoder := gob.NewEncoder(&content)

	err := encoder.Encode(&w)

	if err != nil {
		log.Print(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
