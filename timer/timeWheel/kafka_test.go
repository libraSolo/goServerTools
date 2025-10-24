package timeWheel

import (
	"log"
	"testing"
	"time"
)

func TestTimeWheel(t *testing.T) {
	// Create a new time wheel with a tick of 1 second and a wheel size of 10
	tw := NewTimeWheel(1000, 10, time.Now().UnixNano()/1e6, NewDelayQueue(1024))
	tw.Start()

	// Add a task to be executed after 1 second
	tw.add(&TimerTaskEntity{
		DelayTime: time.Now().Add(1 * time.Second).UnixNano() / 1e6,
		Task: func() {
			log.Println("This is a delay 1s task")
		},
	})

	// Add a task to be executed after 3 seconds
	tw.add(&TimerTaskEntity{
		DelayTime: time.Now().Add(3 * time.Second).UnixNano() / 1e6,
		Task: func() {
			log.Println("This is a delay 3s task")
		},
	})

	// Add a task to be executed after 9 seconds
	tw.add(&TimerTaskEntity{
		DelayTime: time.Now().Add(9 * time.Second).UnixNano() / 1e6,
		Task: func() {
			log.Println("This is a delay 9s task")
		},
	})

	// Wait for 20 seconds to allow all tasks to be executed
	time.Sleep(20 * time.Second)

	// Stop the time wheel
	tw.Stop()
}