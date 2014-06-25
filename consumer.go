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
	Do(core interface{}) bool
}

type ctrl bool 

func (s ctrl) Do(core interface{}) bool {
	log.Println("consumer ctrl", s)
	return bool(s)
}

type TimeOut func(core interface{}) bool

type Consumer struct{
	name 	string
	sync.RWMutex
	timeout <-chan   time.Time
	timefun TimeOut
	core	interface{}
	produce chan IProduct
	memChan chan IProduct
	exitFlag int32
	queue   *list.List
	waitGroup WaitGroupWrapper
}

func NewConsumer(name string, core interface{}, size int) *Consumer {
	c := &Consumer{
					name: name,
					core: core, 
					produce: make(chan IProduct, 1),
					memChan: make(chan IProduct, size),
					queue: list.New(),
				}
	c.waitGroup.Wrap(func() { c.router() })
	c.waitGroup.Wrap(func() { c.productPump() })
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

	c.produce <- ctrl(true)
	c.Lock()
	close(c.produce)
	c.Unlock()

	// synchronize the close of router() and messagePump()
	c.waitGroup.Wait()
	return nil
}

func (c *Consumer) router() {
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

func (c *Consumer) productPump() {
	
	for {
		select {
		case <-c.timeout:
			if c.timefun(c.core) {
				goto exit
			}
			continue
		case p := <-c.memChan:
			if p.Do(c.core) {
				goto exit
			}
			continue
		default:
			if ele := c.queue.Front(); ele != nil {
				c.Lock()
				p, _ := ele.Value.(IProduct)
				c.queue.Remove(ele)
				c.Unlock()			
				if p.Do(c.core) {
					goto exit
				}
			} else {

			}
		}
	}
exit:
	log.Printf("consumer(%s): closing ... run", c.name)
}

func (c *Consumer) Close() {
	c.produce <- ctrl(true)

	c.Lock()
	close(c.produce)
	c.Unlock()

	c.waitGroup.Wait()
}