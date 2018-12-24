### Test
### 1696heCDbKf3edTvynCjEL9k1aQ4axRNV2
### 1CcRr6KecoEAm1aTT8xCKhN5F9nRbBhuvg

### 假设GO环境已安装好

## 主节点
### step 1：编译工程
go build -o my-btc main.go

### step 2：创建钱包
./my-btc createWallet
### 得到一个钱包地址

### step 3：查看钱包addressList ，不想看可以跳过这步

### step 4：提供一个矿工地址，创建区块链
./my-btc createBlockChain -address "1CcRr6KecoEAm1aTT8xCKhN5F9nRbBhuvg"
### 这里的地址应该是你钱包中生成的地址

### step 5：查看该地址余额信息，区块奖励10
./my-btc getBalance -address "1CcRr6KecoEAm1aTT8xCKhN5F9nRbBhuvg"

### step 6：查看区块信息
./my-btc printChain

### step 7：启动主节点
./my-btc startNode -miner "1CcRr6KecoEAm1aTT8xCKhN5F9nRbBhuvg"
### 指定的矿工地址
### 至此，你在8888端口启动了主节点服务

## 矿工节点服务
### step 1：设置环境变量 如：NODE_ID ：3000
### 否则，仍然使用默认端口8888

### step 2：创建文件夹minerNode，将./my-btc copy到该目录下

### step 3：创建钱包和创世区块过程参数主节点

### step 4：启动节点
./my-btc startNode -miner "3000的矿工地址"

## client节点服务
### 矿工节点服务
### step 1：设置环境变量 如：NODE_ID ：3001
### 否则，仍然使用默认端口8888

### step 2：创建文件夹minerNode，将./my-btc和钱包 copy到该目录下

### step 3：创建3001的矿工地址

### step 4：启动节点
./my-btc startNode -miner "3001的矿工地址"

### step 5：发送交易
./my-btc send -from '["1CcRr6KecoEAm1aTT8xCKhN5F9nRbBhuvg"]' -to '["1M2Li1i2yupV8Qku2UkRcNu5iGWQ8UUo7a"]' -amount '["2"]'

### 查看节点同步数据