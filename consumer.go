package consumer 

import "time"

type IProduct interface {
	Do(core interface{}) bool
}

type Product struct{	
}

func (p *Product) Do(core interface{}) bool{
	return false
}

type TimeOut func(core interface{}) bool

type Consumer struct{
	timeout <-chan   time.Time
	timefun TimeOut
	core	interface{}
	Produce chan IProduct
}

func NewConsumer(core interface{}, size int) *Consumer {
	return &Consumer{
					core: core, 
					Produce: make(chan IProduct, size),
					}
}

func (c *Consumer) SetTimeout(duration time.Duration, timeout TimeOut) {
	c.timeout = time.Tick(duration)
	c.timefun = timeout
}

func (c *Consumer) Run() {
	if c.timeout == nil {
		for p := range c.Produce {
			if p.Do(c.core) {
				break
			}
		}
	} else {
		for {
			select {
			case <-c.timeout:
				if c.timefun(c.core) {
					break
				}
			case p := <-c.Produce:
				if p.Do(c.core) {
					break
				}
			}
		}
	}
}

func (c *Consumer) Close() {
	close(c.Produce)
}