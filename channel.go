// todo: support generic types

package main

import (
	"fmt"
	"sync"
	"time"
)

type Channel struct {
	cond           *sync.Cond
	data           int
	isReadPending  bool
	isWritePending bool
	isClosed       bool
}

func (ch *Channel) Send(data int) {
	// if channel is closed panic
	if ch.isClosed {
		panic("send on a closed channel")
	}

	// acquire lock
	ch.cond.L.Lock()

	// if data is not yet read, wait
	for ch.isReadPending {
		ch.cond.Wait()
	}

	// copy data to channel
	ch.data = data

	// reset the writepending flag
	ch.isWritePending = false

	// set the readpending flag
	ch.isReadPending = true

	// signal the channel that writing is finished
	ch.cond.Signal()

	// wait till read is finished
	for ch.isReadPending {
		ch.cond.Wait()
	}

	// release the lock
	ch.cond.L.Unlock()
}

func (ch *Channel) Recv() int {
	// if channel is closed return zero value
	if ch.isClosed {
		var data int
		return data
	}

	// acquire lock
	ch.cond.L.Lock()

	// if no data in channel wait
	for ch.isWritePending {
		ch.cond.Wait()
	}

	// copy the data from channel
	temp := ch.data

	// reset the readpending flag
	ch.isReadPending = false

	// set the writepending flag
	ch.isWritePending = true

	// signal the channel that reading is finished
	ch.cond.Signal()

	// release the lock
	ch.cond.L.Unlock()

	// return the copied data from channel
	return temp

}

func (ch *Channel) Close() {
	ch.isClosed = true
}

func NewChannel() Channel {
	return Channel{
		cond:           sync.NewCond(&sync.Mutex{}),
		isReadPending:  false,
		isWritePending: true,
	}
}

func receiver(ch *Channel, wg *sync.WaitGroup) {
	defer wg.Done()
	data := ch.Recv()
	time.Sleep(time.Second * 2)
	data2 := ch.Recv()
	fmt.Println("data read from channel ", data)
	fmt.Println("data2 read from channel ", data2)
}

func main() {
	ch := NewChannel()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go receiver(&ch, &wg)
	ch.Send(1)
	ch.Send(2)
	ch.Close()

	wg.Wait()
}
