package main

import (
	"context"
	"coreRPC"
	"coreRPC/registry"
	"coreRPC/xclient"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

type Treasure struct {
	Value  int `json:"value"`
	Weight int `json:"weight"`
}

type TreasureResponse struct {
	StatusCode int        `json:"status_code"`
	StatusMsg  string     `json:"status_msg"`
	Treasure   []Treasure `json:"treasure"`
	Capacity   int        `json:"capacity"`
	Limit      int        `json:"limit"`
}

func generate(treasure []Treasure) []Treasure {
	n := len(treasure)
	for i := 0; i < n; i++ {
		treasure[i].Value = rand.Intn(50) + 1
		treasure[i].Weight = rand.Intn(50) + 1
	}
	return treasure
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
func checkValid(treasure []Treasure, capacity, limit int) bool {
	n := len(treasure)
	dp := make([]int, capacity+1)
	for i := 0; i < n; i++ {
		for j := capacity; j >= treasure[i].Weight; j-- {
			dp[j] = max(dp[j], dp[j-treasure[i].Weight]+treasure[i].Value)
		}
	}
	fmt.Println("最多这么多：", dp[capacity])
	return dp[capacity] <= limit
}

func (t Treasure) GetTreasure(args int, resp *TreasureResponse) error {
	n := args
	treasure := make([]Treasure, n)
	capacity := 0
	limit := 0
	generateSuccess := false

	for i := 0; i < 1000; i++ {
		treasure = generate(treasure)
		capacity = rand.Intn(1000) + 1
		limit = rand.Intn(100) + 500
		fmt.Printf("第 %d 次生成\n", i)
		if checkValid(treasure, capacity, limit) {
			generateSuccess = true
			break
		}
	}

	if !generateSuccess {
		treasure[0].Value, treasure[0].Weight = limit, capacity
		for i := 1; i < 50; i++ {
			treasure[i].Value, treasure[i].Weight = 0, 0
		}
	}

	resp.StatusCode = 200
	resp.StatusMsg = "success"
	resp.Treasure = treasure
	resp.Capacity = capacity
	resp.Limit = limit

	return nil
}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

func startServer(registryAddr string) {
	treasure := new(Treasure)
	l, _ := net.Listen("tcp", ":9365")
	server := coreRPC.NewServer()
	_ = server.Register(treasure)
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)
	server.Accept(l)
}

func call(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	resp := new(TreasureResponse)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	err := xc.Call(ctx, "Treasure.GetTreasure", 50, resp)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(resp)
}

func main() {
	log.SetFlags(0)
	registryAddr := "http://localhost:9999/_geerpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	startServer(registryAddr)
}
