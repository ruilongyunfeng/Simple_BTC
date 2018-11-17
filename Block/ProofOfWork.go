package Block

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

type ProofOfWork struct {
	Block *Block
	target *big.Int
}
// 256位Hash里面前面至少要有16个零
const  targetBit  = 20

func(pow *ProofOfWork) prepareData(nonce int) []byte{
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,
			pow.Block.HashTransactions(),
			IntToHex(pow.Block.Timestamp),
			IntToHex(int64(targetBit)),
			IntToHex(int64(nonce)),
			IntToHex(int64(pow.Block.Height)),
		},
		[]byte{},
		)
	return data
}

func (pow *ProofOfWork) Run()([]byte,int64){
	nonce := 0

	var hashInt big.Int
	var hash [32]byte

	for {
		dataBytes := pow.prepareData(nonce)

		//hash
		hash = sha256.Sum256(dataBytes)

		hashInt.SetBytes(hash[:])

		if pow.target.Cmp(&hashInt) == 1{
			break
		}
		nonce = nonce + 1
	}
	return hash[:] ,int64(nonce)
}

func NewProofOfWork(block *Block) *ProofOfWork{
	target := big.NewInt(1)

	target = target.Lsh(target,256-targetBit)

	return &ProofOfWork{block,target}
}
