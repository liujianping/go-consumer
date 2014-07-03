package consumer

import (
	"sync"
	"sync/atomic"
	"log"
	"time"
	"errors"
)

type persistConsumer struct{
	sync.RWMutex
	name 	string

	context Context
	fork	int
	
	incomeChan chan interface{}
	memoryChan chan interface{}

	running		int32
	stopChan	chan bool
	
	requestCount uint64
	finishedCount uint64
	queue		 *diskQueue
	waitGroup WaitGroupWrapper	
}

func NewPersistConsumer(name string, maxMemorySize int, dataPath string, maxBytesPerFile int64, syncEvery int64, syncTimeout time.Duration) Consumer {
	q := newDiskQueue(name, dataPath, maxBytesPerFile, syncEvery, syncTimeout)

	m := persistConsumer{
		name: name,
		incomeChan: make(chan interface{}),
		memoryChan: make(chan interface{}, maxMemorySize),
		stopChan:	make(chan bool),
		queue:		q,

	}
	return &m 
}

func (c *persistConsumer) Resume(context Context, fork int) error {
	c.Lock()
	defer c.Unlock()

	if atomic.LoadInt32(&c.running) == 1 {
		log.Printf("WARN: consumer(%s) already running", c.name)
		return nil
	}
	atomic.StoreInt32(&c.running, 1)
	//! resume set
	c.context = context
	c.fork = fork

	log.Printf("INFO: consumer(%s) resume with depth (%d)", c.name, c.Depth())
	//! router the put product to memory or disk
	c.waitGroup.Wrap(func() { c.router() })	

	if c.fork == 0 || c.fork > 1 {
		log.Printf("INFO: consumer(%s) running with (%d) threads mode!", c.name, c.fork)
	}
	//! starting the productPump with fork type 
	if c.fork == 0 {
		c.waitGroup.Wrap(func() { c.requestPump(true) })
	} else {
		for i := 0; i < c.fork; i++ {
			c.waitGroup.Wrap(func() { c.requestPump(false) })	
		}	
	}

	return nil		
}

func (c *persistConsumer) Put(request interface{}) error {
	c.Lock()
	defer c.Unlock()

	if atomic.LoadInt32(&c.running) == 2 {
		return errors.New("exiting")
	}

	c.incomeChan <- request
	atomic.AddUint64(&c.requestCount, 1)
	return nil
}

func (c *persistConsumer) flush() error {
	log.Printf("INFO: consumer(%s) flush depth (%d)", c.name, c.Depth())
	var req interface{}

	for {
			select {
			case req = <-c.memoryChan:
				buf, err := c.context.Encode(req)
				if err != nil {
					log.Printf("WARN: consumer(%s): context encode request failed - %s", c.name, err.Error())
					continue
				}

				if err := c.queue.Put(buf); err != nil {
					log.Printf("WARN: consumer(%s): queue put failed - %s", c.name, err.Error())
					continue
				}
			default:
				goto finish
			}
			log.Printf("INFO: consumer(%s) flush reqeust to queue- %v", c.name, req)
		}		
finish:
	return nil
}

func (c *persistConsumer) Close() error {

	if !atomic.CompareAndSwapInt32(&c.running, 1, 0) {
	 	return errors.New("exiting")
	}

	c.Lock()
	close(c.incomeChan)
	c.Unlock()
	
	close(c.stopChan)

	// synchronize the close of router() and requestPump()
	c.waitGroup.Wait()
	log.Printf("INFO: consumer(%s) closed with req(%d) fin(%d)", 
				c.name, 
				atomic.LoadUint64(&c.requestCount), 
				atomic.LoadUint64(&c.finishedCount))

	// write anything leftover to disk
	c.flush()
	return c.queue.Close()
}

func (c *persistConsumer) Running() bool {
	return atomic.LoadInt32(&c.running) == 1
}
func (c *persistConsumer) Depth() int64 {
	log.Printf("DEBUG: consumer(%s) depth with memory chan (%d) and queue depth (%d)",
		c.name, len(c.memoryChan), c.queue.Depth())
	return int64(len(c.memoryChan)) + c.queue.Depth()
}

////////////

func (c *persistConsumer) router() {
	log.Printf("INFO: consumer(%s): starting ... router", c.name)
	for request := range c.incomeChan {
		select {
			case c.memoryChan <- request:
			default:
				buf, err := c.context.Encode(request)
				if err != nil {
					log.Printf("WARN: consumer(%s): context encode request failed - %s", c.name, err.Error())
					continue
				}

				if err := c.queue.Put(buf); err != nil {
					log.Printf("WARN: consumer(%s): queue put failed - %s", c.name, err.Error())
					continue
				}
		}		
	}
	log.Printf("INFO: consumer(%s): closing ... router", c.name)
}

func (c *persistConsumer) requestPump(fork bool) {
	log.Printf("INFO: consumer(%s): starting ... requestPump", c.name)
	var req interface{}
	var memoryChan chan interface{}
	var queueChan chan []byte
	
	memoryChan = c.memoryChan
	queueChan = c.queue.ReadChan()	
	for {
		select {
		case req = <-memoryChan:
		case buf := <-queueChan:
			var err error
			req, err = c.context.Decode(buf)
			if err != nil {
				log.Printf("WARN: consumer(%s) context decode failed - %s", c.name, err.Error())
				continue
			}
		case <-c.stopChan:
			goto exit
		}

		if fork {
			c.waitGroup.Wrap(func() { 
				if err := c.context.Do(req); err != nil {
					log.Printf("WARN: consumer(%s) context do request failed - %s", c.name, err.Error())
				}
				atomic.AddUint64(&c.finishedCount, 1)
			})	
		} else {
			if err := c.context.Do(req); err != nil {
				log.Printf("WARN: consumer(%s) context do request failed - %s", c.name, err.Error())
			}
			atomic.AddUint64(&c.finishedCount, 1)
		}
	}

exit:
	log.Printf("INFO: consumer(%s): closing ... requestPump", c.name)
}

