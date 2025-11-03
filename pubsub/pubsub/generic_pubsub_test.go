package pubsub

import (
	"fmt"
	"sort"
	"sync"
	"testing"

	"github.com/bmizerany/assert"
)

// recorder 记录接收到的事件
type recorder[T any] struct {
	mu     sync.Mutex
	events []string
}

func (r *recorder[T]) handle(subject string, content T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	event := fmt.Sprintf("%s: %v", subject, content)
	r.events = append(r.events, event)
	fmt.Printf("Event Recorded: %s\n", event)
}

func (r *recorder[T]) getEvents() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	sort.Strings(r.events)
	return r.events
}

func TestExactSubscription(t *testing.T) {
	t.Log("--- Running TestExactSubscription ---")
	ps := NewGenericPubSub[string]()
	r := &recorder[string]{}
	err := ps.Subscribe("A", "apple", r.handle)
	assert.Equal(t, nil, err)
	t.Logf("Subscribed 'A' to 'apple'")

	t.Logf("Publishing 'hello' to 'apple'")
	ps.Publish("apple", "hello")

	events := r.getEvents()
	t.Logf("Recorded events: %v", events)
	assert.Equal(t, []string{"apple: hello"}, events)
	t.Log("--- TestExactSubscription PASSED ---")
}

func TestWildcardSubscription(t *testing.T) {
	t.Log("--- Running TestWildcardSubscription ---")
	ps := NewGenericPubSub[string]()
	r := &recorder[string]{}
	err := ps.Subscribe("B", "apple.*", r.handle)
	assert.Equal(t, nil, err)
	t.Logf("Subscribed 'B' to 'apple.*'")

	t.Logf("Publishing 'x' to 'apple.'")
	ps.Publish("apple.", "x")
	t.Logf("Publishing 'y' to 'apple.1'")
	ps.Publish("apple.1", "y")
	t.Logf("Publishing 'z' to 'banana' (should not be received)")
	ps.Publish("banana", "z")

	events := r.getEvents()
	t.Logf("Recorded events: %v", events)
	sort.Strings(events)
	assert.Equal(t, []string{"apple.: x", "apple.1: y"}, events)
	t.Log("--- TestWildcardSubscription PASSED ---")
}

func TestStarOnlySubscription(t *testing.T) {
	t.Log("--- Running TestStarOnlySubscription ---")
	ps := NewGenericPubSub[string]()
	r := &recorder[string]{}
	err := ps.Subscribe("C", "*", r.handle) // 订阅所有主题
	assert.Equal(t, nil, err)
	t.Logf("Subscribed 'C' to '*'")

	t.Logf("Publishing 'ok' to 'anything'")
	ps.Publish("anything", "ok")

	events := r.getEvents()
	t.Logf("Recorded events: %v", events)
	assert.Equal(t, []string{"anything: ok"}, events)
	t.Log("--- TestStarOnlySubscription PASSED ---")
}

func TestUnsubscribeExact(t *testing.T) {
	t.Log("--- Running TestUnsubscribeExact ---")
	ps := NewGenericPubSub[string]()
	r := &recorder[string]{}
	ps.Subscribe("A", "apple", r.handle)
	t.Logf("Subscribed 'A' to 'apple'")
	ps.Unsubscribe("A", "apple")
	t.Logf("Unsubscribed 'A' from 'apple'")

	t.Logf("Publishing 'hello' to 'apple' (should not be received)")
	ps.Publish("apple", "hello")

	events := r.getEvents()
	t.Logf("Recorded events: %v", events)
	assert.Equal(t, 0, len(events))
	t.Log("--- TestUnsubscribeExact PASSED ---")
}

func TestUnsubscribeWildcard(t *testing.T) {
	t.Log("--- Running TestUnsubscribeWildcard ---")
	ps := NewGenericPubSub[string]()
	r := &recorder[string]{}
	ps.Subscribe("B", "apple.*", r.handle)
	t.Logf("Subscribed 'B' to 'apple.*'")
	ps.Unsubscribe("B", "apple.*")
	t.Logf("Unsubscribed 'B' from 'apple.*'")

	t.Logf("Publishing 'y' to 'apple.1' (should not be received)")
	ps.Publish("apple.1", "y")

	events := r.getEvents()
	t.Logf("Recorded events: %v", events)
	assert.Equal(t, 0, len(events))
	t.Log("--- TestUnsubscribeWildcard PASSED ---")
}

func TestUnsubscribeAll(t *testing.T) {
	t.Log("--- Running TestUnsubscribeAll ---")
	ps := NewGenericPubSub[string]()
	r := &recorder[string]{}
	ps.Subscribe("C", "apple", r.handle)
	t.Logf("Subscribed 'C' to 'apple'")
	ps.Subscribe("C", "banana.*", r.handle)
	t.Logf("Subscribed 'C' to 'banana.*'")
	ps.UnsubscribeAll("C")
	t.Logf("Unsubscribed all from 'C'")

	t.Logf("Publishing 'x' to 'apple' (should not be received)")
	ps.Publish("apple", "x")
	t.Logf("Publishing 'y' to 'banana.1' (should not be received)")
	ps.Publish("banana.1", "y")

	events := r.getEvents()
	t.Logf("Recorded events: %v", events)
	assert.Equal(t, 0, len(events))
	t.Log("--- TestUnsubscribeAll PASSED ---")
}

func TestErrorHandling(t *testing.T) {
	t.Log("--- Running TestErrorHandling ---")
	ps := NewGenericPubSub[string]()

	err := ps.Subscribe("s1", "a.*.c", func(s string, c string) {})
	assert.NotEqual(t, nil, err)
	t.Logf("Caught expected error for invalid wildcard: %v", err)

	err = ps.Publish("a.*.c", "hello")
	assert.NotEqual(t, nil, err)
	t.Logf("Caught expected error for publishing with wildcard: %v", err)

	err = ps.Subscribe("", "a.b.c", func(s string, c string) {})
	assert.NotEqual(t, nil, err)
	t.Logf("Caught expected error for empty subscriber ID: %v", err)

	err = ps.Subscribe("s1", "a.b.c", nil)
	assert.NotEqual(t, nil, err)
	t.Logf("Caught expected error for nil handler: %v", err)
	t.Log("--- TestErrorHandling PASSED ---")
}

func TestMiddleware(t *testing.T) {
	t.Log("--- Running TestMiddleware ---")
	ps := NewPubSubWithMiddleware[string]()
	r := &recorder[string]{}

	prefixMiddleware := func(subject string, content string, next Handler[string]) {
		t.Logf("Middleware: adding prefix to content '%s'", content)
		next(subject, "prefixed-"+content)
	}

	ps.Use(prefixMiddleware)
	t.Log("Applied prefix middleware")

	err := ps.Subscribe("s1", "a.b.c", r.handle)
	assert.Equal(t, nil, err)
	t.Logf("Subscribed 's1' to 'a.b.c' with middleware")

	ps.Publish("a.b.c", "hello")
	t.Logf("Published 'hello' to 'a.b.c'")

	events := r.getEvents()
	t.Logf("Recorded events: %v", events)
	assert.Equal(t, []string{"a.b.c: prefixed-hello"}, events)
	t.Log("--- TestMiddleware PASSED ---")
}

func TestStats(t *testing.T) {
	t.Log("--- Running TestStats ---")
	ps := NewGenericPubSub[string]()
	ps.Subscribe("A", "apple", func(s string, c string) {})
	ps.Subscribe("B", "banana", func(s string, c string) {})
	ps.Subscribe("C", "apple.*", func(s string, c string) {})

	ps.Publish("apple", "fruit")
	ps.Publish("banana", "fruit")
	ps.Publish("apple.pie", "dessert")

	stats := ps.Stats()
	t.Logf("Collected stats: %+v", stats)

	assert.Equal(t, 3, stats.SubscribersCount)
	assert.Equal(t, 2, stats.ExactSubscriptions)
	assert.Equal(t, 1, stats.WildcardSubscriptions)
	assert.Equal(t, int64(3), stats.MessagesPublished)
	assert.Equal(t, int64(4), stats.MessagesDelivered) // apple (A, C), banana (B), apple.pie (C)
	t.Log("--- TestStats PASSED ---")
}

func TestBatchSubscribe(t *testing.T) {
	t.Log("--- Running TestBatchSubscribe ---")
	ps := NewGenericPubSub[string]()
	r := &recorder[string]{}
	subjects := []string{"topic1", "topic2", "topic3.*"}
	err := ps.BatchSubscribe("subscriber1", subjects, r.handle)
	assert.Equal(t, nil, err)
	t.Logf("Batch subscribed 'subscriber1' to %v", subjects)

	ps.Publish("topic1", "data1")
	ps.Publish("topic2", "data2")
	ps.Publish("topic3.sub", "data3")

	events := r.getEvents()
	t.Logf("Recorded events: %v", events)
	assert.Equal(t, []string{"topic1: data1", "topic2: data2", "topic3.sub: data3"}, events)
	t.Log("--- TestBatchSubscribe PASSED ---")
}

func TestBatchPublish(t *testing.T) {
	t.Log("--- Running TestBatchPublish ---")
	ps := NewGenericPubSub[string]()
	r := &recorder[string]{}
	ps.Subscribe("s1", "topic1", r.handle)
	ps.Subscribe("s2", "topic2", r.handle)

	messages := map[string]string{
		"topic1": "message1",
		"topic2": "message2",
	}

	err := ps.BatchPublish(messages)
	assert.Equal(t, nil, err)
	t.Logf("Batch published messages: %v", messages)

	events := r.getEvents()
	t.Logf("Recorded events: %v", events)
	assert.Equal(t, []string{"topic1: message1", "topic2: message2"}, events)
	t.Log("--- TestBatchPublish PASSED ---")
}

func TestGetSubscriptions(t *testing.T) {
	t.Log("--- Running TestGetSubscriptions ---")
	ps := NewGenericPubSub[string]()
	ps.Subscribe("A", "exact1", func(s string, c string) {})
	ps.Subscribe("A", "exact2", func(s string, c string) {})
	ps.Subscribe("A", "wild.*", func(s string, c string) {})

	exact, wildcard := ps.GetSubscriptions("A")
	t.Logf("Retrieved subscriptions for 'A': exact=%v, wildcard=%v", exact, wildcard)

	sort.Strings(exact)
	sort.Strings(wildcard)

	assert.Equal(t, []string{"exact1", "exact2"}, exact)
	assert.Equal(t, []string{"wild.*"}, wildcard)
	t.Log("--- TestGetSubscriptions PASSED ---")
}

func TestIsSubscribed(t *testing.T) {
	t.Log("--- Running TestIsSubscribed ---")
	ps := NewGenericPubSub[string]()
	ps.Subscribe("A", "exact", func(s string, c string) {})
	ps.Subscribe("A", "wild.*", func(s string, c string) {})

	assert.Equal(t, true, ps.IsSubscribed("A", "exact"))
	assert.Equal(t, true, ps.IsSubscribed("A", "wild.sub"))
	assert.Equal(t, false, ps.IsSubscribed("A", "something.else"))
	assert.Equal(t, false, ps.IsSubscribed("B", "exact"))
	t.Log("--- TestIsSubscribed PASSED ---")
}

func TestAsyncPublish(t *testing.T) {
	t.Log("--- Running TestAsyncPublish ---")
	ps := NewAsyncPubSub[string](2)
	defer ps.Shutdown()

	r := &recorder[string]{}
	ps.Subscribe("A", "async.topic", r.handle)

	// Use a wait group to wait for the handler to be called
	var wg sync.WaitGroup
	wg.Add(1)
	originalHandler := r.handle
	r.handle = func(subject string, content string) {
		originalHandler(subject, content)
		wg.Done()
	}

	errChan := ps.PublishAsync("async.topic", "async_data")
	
	err := <-errChan
	assert.Equal(t, nil, err)

	wg.Wait() // Wait for the message to be processed

	events := r.getEvents()
	t.Logf("Recorded events: %v", events)
	assert.Equal(t, []string{"async.topic: async_data"}, events)
	t.Log("--- TestAsyncPublish PASSED ---")
}

func TestConcurrentPublish(t *testing.T) {
	t.Log("--- Running TestConcurrentPublish ---")
	ps := NewGenericPubSub[string]()
	r := &recorder[string]{}
	ps.Subscribe("A", "test.*", r.handle)

	var wg sync.WaitGroup
	numMessages := 100

	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ps.Publish(fmt.Sprintf("test.%d", i), "data")
		}(i)
	}

	wg.Wait()

	// Allow some time for handlers to complete
	// In a real-world scenario, you might need a more robust synchronization mechanism
	// but for this test, a short sleep is often sufficient.
	// A better approach is to use a WaitGroup inside the handler, but that complicates the recorder.
	// We will use a channel to signal completion.

	completion := make(chan struct{}, numMessages)
	ps.UnsubscribeAll("A")
	ps.Subscribe("A", "test.*", func(subject string, content string) {
		r.handle(subject, content)
		completion <- struct{}{}
	})

	// Republish to ensure handlers are counted correctly
	for i := 0; i < numMessages; i++ {
		ps.Publish(fmt.Sprintf("test.%d", i), "data")
	}

	for i := 0; i < numMessages; i++ {
		<-completion
	}

	events := r.getEvents()
	t.Logf("Recorded %d events", len(events))
	assert.Equal(t, numMessages, len(events))
	t.Log("--- TestConcurrentPublish PASSED ---")
}
