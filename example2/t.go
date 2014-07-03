package main

import (
	"time"
	"log"
	"encoding/json"
	"github.com/liujianping/consumer"
)

type context struct{
	stop chan bool
}

func (c *context) Do(req interface{}) error {
	r, _ := req.(*MyProduct)
	return r.Do(c)
}

func (c *context) Encode(request interface{}) ([]byte, error) {
	return json.Marshal(request)
}
func (c *context) Decode(data []byte) (interface{}, error) {
	var p MyProduct
	err := json.Unmarshal(data, &p)
	return &p, err
}


func (p *MyProduct) Do(c *context) error {
	log.Printf("product No(%d) do", p.No)

	time.Sleep(time.Millisecond * 250)
	
	log.Printf("product No(%d) Done", p.No)
	return nil
}


type MyProduct struct{
	No int
}

func main() {
	core := &context{ stop: make(chan bool, 0)}

	consumer := consumer.NewPersistConsumer("sleepy", 10, "./", 10240, 8, time.Second)
	consumer.Resume(core, 2)
	
	//! uncomment and mod for your test
	// for i:= 31; i <= 60; i++ {
	// 	consumer.Put(&MyProduct{i})
	// }
	log.Printf("consumer running %v", consumer.Running())
	
	consumer.Close()	
	
	log.Printf("consumer running %v", consumer.Running())
}