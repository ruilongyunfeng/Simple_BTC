package Block

import "crypto/sha256"

type MerkleTree struct {
	rootNode *MerkleNode
}

type MerkleNode struct {
	left *MerkleNode
	right *MerkleNode
	data []byte
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	//判断节点数 单双
	var nodes []MerkleNode

	if len(data) % 2 != 0{//单数复制尾部节点
		data = append(data,data[len(data)-1])
	}

	//构造叶子节点
	for _,datum := range data  {
		node := NewMerkleNode(nil,nil,datum)
		nodes = append(nodes,*node)
	}

	for i:=0;i<len(data)/2;i++ {
		var newLevel []MerkleNode
		//make double
		if len(nodes)%2 != 0{
			nodes = append(nodes,nodes[len(nodes)-1])
		}

		for j:=0; j<len(nodes) ;j++  {
			node := NewMerkleNode(&nodes[j],&nodes[j+1],nil)
			newLevel = append(newLevel,*node)
		}
		nodes = newLevel
	}

	merkleTree := MerkleTree{&nodes[0]}

	return &merkleTree
}

func NewMerkleNode(left,right *MerkleNode,data []byte) *MerkleNode{
	newMerkleNode := MerkleNode{}

	//叶子节点
	if left == nil && right == nil{
		hash := sha256.Sum256(data)
		newMerkleNode.data = hash[:]
	}else {
		//非叶子节点
		hash := append(left.data,right.data...)
		newData := sha256.Sum256(hash)
		newMerkleNode.data = newData[:]
	}

	newMerkleNode.left = left
	newMerkleNode.right = right

	return &newMerkleNode
}