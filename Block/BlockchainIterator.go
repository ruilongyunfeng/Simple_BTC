package Block

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockchainIterator struct {
	CurrentHash []byte
	DB  *bolt.DB
}

func (bct *BlockchainIterator) Next() *Block{

	var block *Block

	err := bct.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))

		if b != nil{
			currentBlockBytes := b.Get(bct.CurrentHash)
			block = DeSerializeBlock(currentBlockBytes)

			bct.CurrentHash = block.PrevBlockHash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return block
}
