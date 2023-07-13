package main

import (
	"fmt"
	"sync"
)

func main() {
	c := make(chan int, 10)
	wg := sync.WaitGroup{}
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Println(i, "--> 我被调用了")
			kk := <-c
			fmt.Println("拿到 kk ", kk*0+i)
		}(i)
	}

	fmt.Println("我要放东西进去了，不知道那边怎么样")
	for i := 1; i <= 10; i++ {
		c <- i
	}

	wg.Wait()
}
