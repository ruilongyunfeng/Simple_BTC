package Block

import "fmt"

/* *************************************************************
 * Copyright  2018 Bridge-ruijiezhi@163.com. All rights reserved.
 *
 * FileName: CLI_Function
 *
 * @Author: Bridge 2018/11/22 19:58
 *
 * @Version: 1.0
 * *************************************************************/

func (cli *CLI) createWallet(nodeID string) {

	wallets, _ := NewWallets(nodeID)

	wallets.CreateNewWallet(nodeID)

	fmt.Println(len(wallets.walletsMap))
}
