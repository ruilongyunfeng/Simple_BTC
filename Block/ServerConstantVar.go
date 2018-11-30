package Block

/* *************************************************************
 * Copyright  2018 Bridge-ruijiezhi@163.com. All rights reserved.
 *
 * FileName: ServerConstantVar
 *
 * @Author: Bridge 2018/11/30 15:11
 *
 * @Version: 1.0
 * *************************************************************/

//存储节点全局变量
var knowNodes = []string{"localhost:3000"}
var nodeAddress string //全局变量，节点地址
// 存储hash值
var hashCacheArray [][]byte
var minerAddress string
var memoryTxPool = make(map[string]*Transaction)
