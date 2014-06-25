# golang channel 经常由于消费速度慢于生产速度，导致channel操作阻塞(即使设置了channel得缓存空间，仍然会出问题)，严重影响了程序性能。

[![GoDoc](http://godoc.org/github.com/liujianping/consumer?status.png)](http://godoc.org/github.com/liujianping/consumer)

##  安装
	
	go get github.com/liujianping/consumer

##  特点

- 支持基于本地内存的队列缓存
- 支持基于本地文件的队列缓存（todo）

##  用例

````go
package main

import (
	"time"
	"log"
	"github.com/liujianping/consumer"
)

type Core struct{
	main chan bool
	count int
}

type MyProduct struct{
	No int
}

func (p *MyProduct) Do(core interface{}) bool {
	log.Printf("product No(%d) do", p.No)

	c,_ := core.(*Core)

	time.Sleep(time.Millisecond * 100)
	log.Println("core", c.count)
	c.count++

	return false
}

func timeout(core interface{}) bool {
	c,_ := core.(*Core)
	log.Println("timeout", c.count)
	return false
}

func main() {

	core := &Core{ main: make(chan bool, 0), count:0}

	consumer := consumer.NewConsumer("sleepy",core, 10)
	
	for i:= 1; i <= 30; i++ {
		consumer.Produce(&MyProduct{i})
	}
	
	consumer.Close()
}

````

##  更多
