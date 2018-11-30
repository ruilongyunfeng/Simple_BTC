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
