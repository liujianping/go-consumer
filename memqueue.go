package consumer

import (
	"sync"
	"container/list"
	"log"
)

type memQueue struct {
	name string
	sync.RWMutex
	queue	 *list.List

	// exposed via ReadChan()
	readChan chan interface{}

	// internal channels
	writeChan         chan interface{}
	writeResponseChan chan error
	exitChan          chan int
	waitGroup WaitGroupWrapper
}

func newMemQueue(name string) *memQueue {
	q := list.New()
	m := memQueue{
		name: name,
		queue: q,
		readChan: make(chan interface{}),
		writeChan: make(chan interface{}),
		writeResponseChan: make(chan error),
		exitChan: make(chan int),
	}

	m.waitGroup.Wrap( func() {m.ioLoop()} )
	return &m
}

func (d *memQueue) Put(data interface{}) error {
	d.RLock()
	defer d.RUnlock()

	d.writeChan <- data
	return <-d.writeResponseChan
}

func (d *memQueue) ReadChan() chan interface{} {
	return d.readChan
}

func (d *memQueue) Close() error {
	
	close(d.exitChan)
	d.waitGroup.Wait()
	
	d.flush()
	return nil
}

func (d *memQueue) Depth() int64 {
	d.RLock()
	defer d.RUnlock()
	return int64(d.queue.Len() + len(d.readChan))
}

func (d *memQueue) write(data interface{}) error {
	d.queue.PushBack(data)
	return nil
}

func (d *memQueue) read() interface{} {
	if n := d.queue.Front(); n != nil {
		data := n.Value
		return data
	}
	return nil
}

func (d *memQueue) flush() error {
	log.Printf("memQueue(%s): flushing ...", d.name)

	for {
		if ele := d.queue.Front(); ele != nil {
			log.Printf("memQueue(%s): flush request - %v", d.name, ele.Value)
			d.queue.Remove(ele)
			continue
		}
		break
	} 
	
	return nil
}

func (d *memQueue) ioLoop() {
	log.Printf("memQueue(%s): starting ... ioLoop", d.name)
	var r chan interface{}
	var dataRead interface{}

	for {
		//! if storage has data,prepare data
		if  dataRead == nil {
			dataRead = d.read()			
		}

		if dataRead != nil {
			r = d.readChan
		} else {
			r = nil
		}
		//! select for the channel
		select {
		case r <- dataRead:
			d.queue.Remove(d.queue.Front())
			dataRead = nil			
		case dataWrite := <- d.writeChan:
			d.writeResponseChan <- d.write(dataWrite)
		case <- d.exitChan:
			goto exit
		}
	}
exit:
	log.Printf("memQueue(%s): closing ... ioLoop", d.name)
}