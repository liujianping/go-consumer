
package consumer 

/*
import "time"

type Core struct{
	main chan bool
	count int
}

type MyProduct struct{
	Product
}

func (p *MyProduct) Do(core interface{}) bool {
	println("myproduct do")

	c,_ := core.(*Core)
	println("core", c.count)

	return false
}

func timeout(core interface{}) bool {
	c,_ := core.(*Core)
	c.count ++
	println("timeout", c.count)
	if c.count >= 10 {
		c.main <- true
		return true
	}
	return false
}

func main() {

	core := &Core{ main: make(chan bool, 0), count:0}

	consumer := NewConsumer(core, 10)
	consumer.SetTimeout(time.Second, timeout)
	go consumer.Run()

	p1 := &MyProduct{}
	consumer.Produce <- p1

	time.Sleep(time.Second*2)

	p2 := &MyProduct{}
	consumer.Produce <- p2	

	<-core.main
	
	consumer.Close()
}
*/