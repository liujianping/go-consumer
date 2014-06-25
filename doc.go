
package consumer 

/*
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

	time.Sleep(time.Millisecond * 250)
	log.Println("core", c.count)
	c.count++

	return false
}

func timeout(core interface{}) bool {
	c,_ := core.(*Core)
	log.Println("timeout ...............", c.count)
	return false
}

func main() {

	core := &Core{ main: make(chan bool, 0), count:0}

	consumer := consumer.NewConsumer("sleepy",core, 10)
	consumer.SetTimeout(time.Second, timeout)

	for i:= 1; i <= 30; i++ {
		consumer.Produce(&MyProduct{i})
	}
	
	consumer.Close()
}
*/