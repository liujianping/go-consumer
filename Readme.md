# golang consumer 消费者 

golang channel 经常由于消费速度慢于生产速度，导致channel操作阻塞(即使设置了channel得缓存空间，仍然会出问题)，严重影响了程序性能。

[![GoDoc](http://godoc.org/github.com/liujianping/consumer?status.png)](http://godoc.org/github.com/liujianping/consumer)

##  安装
	
	go get github.com/liujianping/consumer

##  特点

- 支持基于本地内存的队列缓存
- 支持基于本地文件的队列缓存（todo）
- 支持可配置的并发数设置

##  接口

### 创建对象

````go
	
	consumer.NewConsumer(name string, ctx interface{}, memorySize int, fork int)

	- name: 消费者名称
	- ctx:	消费者执行上下文
	- memorySize: chan memory 最大预分配的内存
	- fork:	是否使用 go 产生子执行绪 进行 IProduct.Do()
			0, 每个IProduct对象均使用 go IProduct.Do() 开启新的执行绪操作 
			1, 单线程操作 IProduct.Do() 
			n, 固定n个线程操作 IProduct.Do()  

````

### 设置超时回调

	type TimeOut func(ctx interface{}) error

	consumer.SetTimeout(duration time.Duration, timeout TimeOut)

### Put IProduct对象 入队

````go
	
	type IProduct interface {
		Do(ctx interface{}) error
	}

	consumer.Produce(p IProduct) error

````

##  用例

````go
package main

import (
	"time"
	"log"
	"errors"
	"sync/atomic"
	"github.com/liujianping/consumer"
)

type Core struct{
	main chan bool
	count int32
}

type MyProduct struct{
	No int
}

func (p *MyProduct) Do(core interface{}) error {
	log.Printf("product No(%d) do", p.No)

	c,_ := core.(*Core)

	time.Sleep(time.Millisecond * 250)
	
	atomic.AddInt32(&c.count, 1)

	curr := atomic.LoadInt32(&c.count)

	log.Println("core", curr)
	if curr % 3 == 0 {
		return errors.New("error .....")
	}

	return nil
}

func timeout(core interface{}) error {
	c,_ := core.(*Core)
	log.Println("timeout ...............", c.count)
	return nil
}

func main() {

	core := &Core{ main: make(chan bool, 0), count:0}


	//consumer := consumer.NewConsumer("sleepy", core, 10, 0)
	//consumer := consumer.NewConsumer("sleepy", core, 10, 1)
	consumer := consumer.NewConsumer("sleepy", core, 10, 8)
	
	consumer.SetTimeout(time.Second, timeout)

	for i:= 1; i <= 60; i++ {
		consumer.Produce(&MyProduct{i})
	}
	
	consumer.Close()
}

````

##  更多
