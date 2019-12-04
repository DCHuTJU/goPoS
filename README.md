# goPoS
#### 1. 基本结构

##### 1.1 区块结构

```go
type Block struct {
	Number    int
	BPM       int
	Hash      string
	PrevHash  string
	Validator string
	Timestamp time.Time
}
```

其中`BPM`用于之后的共识。

##### 1.2 区块链结构

```go
var Blockchain []Block
```

#### 2. 区块操作

##### 2.1 生成区块

```go
func GenerateBlock(prevBlock Block, BPM int, address string) Block {
	return newBlock
}
```

##### 2.2 计算Hash

```go
func (b *Block) CalculateBlockHash() string {
	record := string(b.Number) + b.Timestamp.String() + string(b.BPM) + b.PrevHash
	return calculateHash(record)
}
```

在这里使用的是`sha256`的`Hash`计算方法。

##### 2.3 验证区块

```go
func IsBlockValid(newBlock, prevBlock Block) bool {
	if prevBlock.Number+1 != newBlock.Number {
		return false
	}
	if prevBlock.Hash != newBlock.PrevHash {
		return false
	}
	if newBlock.CalculateBlockHash() != newBlock.Hash {
		return false
	}
	return true
}
```

#### 3. 共识过程

##### 3.1 生成创世区块，并生成区块链

```go
genesisBlock = Block{0, 0, "", "", "", t}
Blockchain = append(Blockchain, genesisBlock)
```

##### 3.2 获取数据

执行`PoS`需要获取每个节点的`Balance`和`BPM`信息：

```go
// 获取 Balance 信息
for scanBalance.Scan() {
	balance, err := strconv.Atoi(scanBalance.Text())
	validators[address] = balance
}
// 获取 BPM 信息
for {
	bpm, err := strconv.Atoi(scanBPM.Text())
	newBlock := GenerateBlock(prevLastIndex, bpm, address)
	if IsBlockValid(newBlock, prevLastIndex) {
		candidateBlocks = append(candidateBlocks, newBlock) // 将新生成的区块放入到候选区块中
	}
}
```

##### 3.3 选择获胜者

将每个用户的`Stake`作为权重，并从中随机选择获胜者：

```go
k, ok := setValidators[block.Validator]
	if ok {
		for i := 0; i < k; i++ {
			lotteryPool = append(lotteryPool, block.Validator)
		}
	}
}
s := rand.NewSource(time.Now().Unix())
r := rand.New(s)
lotteryMiner := lotteryPool[r.Intn(len(lotteryPool))]
```

