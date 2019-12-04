package main

import (
	"bufio"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

type Block struct {
	Number    int
	BPM       int
	Hash      string
	PrevHash  string
	Validator string
	Timestamp time.Time
}

var Blockchain []Block
var tempBlocks []Block

// 候选区块
var candidateBlocks = make([]Block, 0)

var validators map[string]int

var announcements = make(chan string)

var BPMacceptor = make(chan int, 1)

var mutex = sync.Mutex{}

func GenerateBlock(prevBlock Block, BPM int, address string) Block {
	t := time.Now()
	var newBlock = Block{
		prevBlock.Number + 1,
		BPM,
		"",
		prevBlock.Hash,
		address,
		t,
	}
	newHash := newBlock.CalculateBlockHash()
	newBlock.Hash = newHash
	return newBlock
}

func (b *Block) CalculateBlockHash() string {
	record := string(b.Number) + b.Timestamp.String() + string(b.BPM) + b.PrevHash
	return calculateHash(record)
}

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

func handleConn(conn net.Conn) {
	defer conn.Close()

	go func() {
		for {
			msg := <-announcements
			_, err := io.WriteString(conn, msg)
			if err != nil {
				panic(err)
			}
		}
	}()

	var address string
	_, err := io.WriteString(conn, "Enter token balance: ")
	if err != nil {
		panic(err)
	}
	scanBalance := bufio.NewScanner(conn)
	for scanBalance.Scan() {
		balance, err := strconv.Atoi(scanBalance.Text())
		if err != nil {
			log.Fatal(err)
		}

		t := time.Now()
		address = calculateHash(t.String())
		validators[address] = balance
	}

	_, err = io.WriteString(conn, "\n Enter a new BPM: ")
	if err != nil {
		panic(err)
	}
	mutex.Lock()
	go func() {
		bpm := <-BPMacceptor
		io.WriteString(conn, string(bpm))
	}()
	mutex.Unlock()
	scanBPM := bufio.NewScanner(conn)

	go func() {
		for {
			bpm, err := strconv.Atoi(scanBPM.Text())
			if err != nil {
				log.Printf("%v not a number: %v", scanBPM.Text(), err)
				delete(validators, address)
				conn.Close()
			}
			mutex.Lock()
			if len(Blockchain) < 1 {
				log.Printf("There are no block in the blockchain.")
				conn.Close()
			}
			prevLastIndex := Blockchain[len(Blockchain)-1]
			mutex.Unlock()

			newBlock := GenerateBlock(prevLastIndex, bpm, address)

			spew.Dump(newBlock, prevLastIndex)
			if IsBlockValid(newBlock, prevLastIndex) {
				candidateBlocks = append(candidateBlocks, newBlock)
			}
		}
	}()

	for {
		time.Sleep(time.Minute)
		mutex.Lock()
		output, err := json.Marshal(Blockchain)
		mutex.Unlock()
		if err != nil {
			log.Println(err)
		}
		io.WriteString(conn, string(output)+"\n")
	}
}

func PickWinner() {
	time.Sleep(10 * time.Second)
	mutex.Lock()
	tmp := tempBlocks
	mutex.Unlock()

	lotteryPool := []string{}
	if len(tmp) > 0 {
	OUTER:
		for _, block := range tmp {
			for _, node := range lotteryPool {
				if block.Validator == node {
					continue OUTER
				}
			}

			mutex.Lock()
			setValidators := validators
			mutex.Unlock()

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

		for _, block := range tmp {
			if block.Validator == lotteryMiner {
				mutex.Lock()
				Blockchain = append(Blockchain, block)
				mutex.Unlock()
				for _ = range validators {
					announcements <- "\nwinning validator: " + lotteryMiner + "\n"
				}
				break
			}
		}
	}

	mutex.Lock()
	tempBlocks = []Block{}
	mutex.Unlock()
}

func Run() {
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, 0, "", "", "", t}
	hash := genesisBlock.CalculateBlockHash()
	genesisBlock.Hash = hash

	Blockchain = append(Blockchain, genesisBlock)

	server, err := net.Listen("tcp", "9090")
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	go func() {
		for _, candidate := range candidateBlocks {
			mutex.Lock()
			tempBlocks = append(tempBlocks, candidate)
			mutex.Unlock()
		}
	}()

	go func() {
		for {
			PickWinner()
		}
	}()

	go func() {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}()
}