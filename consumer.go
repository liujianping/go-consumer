package consumer 

import (
	"time"
	"log"
	"errors"
	"sync"
	"sync/atomic"
	"container/list"
)

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}


type IProduct interface {
	Do(ctx interface{}) error
}

type TimeOut func(ctx interface{}) error

type Consumer struct{
	name 	string
	fork	int
	sync.RWMutex
	timeout <-chan   time.Time
	timefun TimeOut
	context	interface{}
	produce chan IProduct
	memChan chan IProduct
	exitFlag int32
	queue   *list.List
	waitGroup WaitGroupWrapper
}

func NewConsumer(name string, ctx interface{}, memorySize int, fork int) *Consumer {
	c := &Consumer{
					name: name,
					fork: fork,
					context: ctx, 
					produce: make(chan IProduct, 1),
					memChan: make(chan IProduct, memorySize),
					queue: list.New(),
				}
	c.waitGroup.Wrap(func() { c.router() })
	
	if c.fork == 0 {
		c.waitGroup.Wrap(func() { c.productPump(true) })
	} else {
		for i := 1; i <= c.fork; i++ {
			c.waitGroup.Wrap(func() { c.productPump(false) })	
		}	
	}
		
	return c
}

func (c *Consumer) Produce(p IProduct) error {
	c.RLock()
	defer c.RUnlock()
	if atomic.LoadInt32(&c.exitFlag) == 1 {
		return errors.New("exiting")
	}
	c.produce <- p
	return nil
}

func (c *Consumer) Exiting() bool {
	return atomic.LoadInt32(&c.exitFlag) == 1
}


func (c *Consumer) Depth() int64 {
	return int64(len(c.memChan)) + int64(c.queue.Len())
}

func (c *Consumer) exit() error {
	if !atomic.CompareAndSwapInt32(&c.exitFlag, 0, 1) {
		return errors.New("exiting")
	}

	atomic.StoreInt32(&c.exitFlag, 1)
	c.Lock()
	close(c.produce)
	c.Unlock()

	// synchronize the close of router() and messagePump()
	c.waitGroup.Wait()
	return nil
}

func (c *Consumer) router() {
	log.Printf("consumer(%s): starting ... router", c.name)
	for product := range c.produce {
		select {
			case c.memChan <- product:
			default:
				c.Lock()
				c.queue.PushBack(product)		
				c.Unlock()
		}		
	}
	log.Printf("consumer(%s): closing ... router", c.name)
}

func (c *Consumer) SetTimeout(duration time.Duration, timeout TimeOut) {
	c.timeout = time.Tick(duration)
	c.timefun = timeout
}

func (c *Consumer) exec(p IProduct) {
	if err := p.Do(c.context); err != nil {
		log.Printf("WARN: consumer(%s) do failed - %s", c.name, err.Error())
	}
}

func (c *Consumer) productPump(fork bool) {
	for {
		select {
		case <-c.timeout:
			if err := c.timefun(c.context); err != nil {
				log.Printf("WARN: consumer(%s) timeout failed - %s", c.name, err.Error())
			}
			continue
		case p := <-c.memChan:
			if fork {
				c.waitGroup.Wrap(func() { c.exec(p)})	
			} else {
				c.exec(p)					
			}			
			continue
		default:
			if ele := c.queue.Front(); ele != nil {
				c.Lock()
				p, _ := ele.Value.(IProduct)
				c.queue.Remove(ele)
				c.Unlock()							
				if fork {
					c.waitGroup.Wrap(func() { c.exec(p)})	
				} else {
					c.exec(p)					
				}
			} else {
			   if atomic.LoadInt32(&c.exitFlag) == 1 {
			   		goto exit
			   }
			}
		}
	}
exit:
	log.Printf("consumer(%s): closing ... productPump", c.name)
}

func (c *Consumer) Close() error{
	return c.exit()
}