package main

import (
	"fmt"
	"time"
)

const PhilosophersCount = 5
const Tick = time.Millisecond * 500


// The Tunnel struct encapsules two channels and preventing using channels in the wrong direction.
type Tunnel struct {
    RequestCh  chan bool
    RespondCh chan bool
}

func makeTunnel() Tunnel {
    return Tunnel{
        RequestCh: make(chan bool),
        RespondCh: make(chan bool),
    }
}
func (t Tunnel) SendRequestAndWait() bool {
	t.RequestCh <- true
	return <-t.RespondCh
}
func (t Tunnel) SendDone() {
	t.RequestCh <- false
}
func (t Tunnel) SendRespond(affirmative bool) {
	t.RespondCh <- affirmative
}

func philosopher(id int, left, right Tunnel) {
	//Every other philosopher starts with a different hand so we introduce some asymmetry.
	if id % 2 == 0 {
		temp := left
		left = right
		right = temp
	}
	
	fmt.Printf("P%d: Thinking\n", id)
	for {
		if left.SendRequestAndWait() {
			if right.SendRequestAndWait() {
				fmt.Printf("P%d: Eating\n", id)
				time.Sleep(Tick)
				fmt.Printf("P%d: Thinking\n", id)
				left.SendDone()
				right.SendDone()
				time.Sleep(Tick)
			} else {
				left.SendDone()
			}
		}
		time.Sleep(Tick)
	}
}

func fork(id int, left, right Tunnel) {
	// The fork is always responding immediatelly based on the availability.
	available := true

	answer := func(msg bool, ch Tunnel) {
		if msg {
			if available {
				available = false
				ch.SendRespond(true)
			} else {
				ch.SendRespond(false)
			}
		} else {
			// msg = "Done" then
			available = true
		}
	}

	for {
		var msg bool
		select {
		case msg = <- left.RequestCh:
			answer(msg, left)
		case msg = <- right.RequestCh:
			answer(msg, right)
		}
	}
}

func main() {
	firstTunnel := makeTunnel()
	sharedTunnel := firstTunnel
	for i := 0; i < PhilosophersCount; i++ {
		middleTunnel := makeTunnel()
		var newTunnel Tunnel
		if i+1 == PhilosophersCount {
			newTunnel = firstTunnel
		} else {
			newTunnel = makeTunnel()
		}
		go philosopher(i, sharedTunnel, middleTunnel)
		go fork(i, middleTunnel, newTunnel)
		sharedTunnel = newTunnel
	}
	select {}
}