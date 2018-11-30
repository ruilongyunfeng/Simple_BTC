package Block

/* *************************************************************
 * Copyright  2018 Bridge-ruijiezhi@163.com. All rights reserved.
 *
 * FileName: ServerMessageStruct
 *
 * @Author: Bridge 2018/11/30 15:04
 *
 * @Version: 1.0
 * *************************************************************/

type MessageTx struct {
	AddrFrom string
	Tx       *Transaction
}

type MessageVersion struct {
	Version    int64  // 版本
	BestHeight int64  // 当前节点区块的高度
	AddrFrom   string //当前节点的地址
}

type MessageGetBlocks struct {
	AddressFrom string
}

type MessageInv struct {
	AddrFrom string   //自己的地址
	Type     string   //类型 block tx
	Items    [][]byte //hash二维数组
}

type MessageGetData struct {
	AddrFrom string
	Type     string
	Hash     []byte
}

type MessageBlock struct {
	AddrFrom string
	Block    []byte
}
