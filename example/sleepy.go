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