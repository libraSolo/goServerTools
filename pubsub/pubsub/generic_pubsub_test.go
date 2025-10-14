package pubsub

import (
    "github.com/bmizerany/assert"
    "sync"
    "testing"
)

// helper 用于记录收到的事件
type recorder struct {
    mu     sync.Mutex
    events []string
}

func (r *recorder) handle(subject, content string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.events = append(r.events, subject+":"+content)
}

func (r *recorder) list() []string {
    r.mu.Lock()
    defer r.mu.Unlock()
    cp := make([]string, len(r.events))
    copy(cp, r.events)
    return cp
}

func TestExactSubscription(t *testing.T) {
    ps := NewGenericPubSub()
    r := &recorder{}
    ps.Subscribe("A", "apple", r.handle)

    ps.Publish("apple", "hello")
    assert.Equal(t, []string{"apple:hello"}, r.list())
}

func TestWildcardSubscription(t *testing.T) {
    ps := NewGenericPubSub()
    r := &recorder{}
    ps.Subscribe("B", "apple.*", r.handle)

    ps.Publish("apple.", "x")
    ps.Publish("apple.1", "y")
    ps.Publish("banana", "z")
    assert.Equal(t, []string{"apple.:x", "apple.1:y"}, r.list())
}

func TestStarOnlySubscription(t *testing.T) {
    ps := NewGenericPubSub()
    r := &recorder{}
    ps.Subscribe("C", "*", r.handle) // 订阅所有主题

    ps.Publish("anything", "ok")
    assert.Equal(t, []string{"anything:ok"}, r.list())
}

func TestUnsubscribeExact(t *testing.T) {
    ps := NewGenericPubSub()
    r := &recorder{}
    ps.Subscribe("A", "apple", r.handle)
    ps.Unsubscribe("A", "apple")

    ps.Publish("apple", "hello")
    assert.Equal(t, 0, len(r.list()))
}

func TestUnsubscribeWildcard(t *testing.T) {
    ps := NewGenericPubSub()
    r := &recorder{}
    ps.Subscribe("B", "apple.*", r.handle)
    ps.Unsubscribe("B", "apple.*")

    ps.Publish("apple.1", "y")
    assert.Equal(t, 0, len(r.list()))
}

func TestUnsubscribeAll(t *testing.T) {
    ps := NewGenericPubSub()
    r := &recorder{}
    ps.Subscribe("C", "apple", r.handle)
    ps.Subscribe("C", "banana.*", r.handle)
    ps.UnsubscribeAll("C")

    ps.Publish("apple", "x")
    ps.Publish("banana.1", "y")
    assert.Equal(t, 0, len(r.list()))
}

func TestSubscribeIllegalWildcard(t *testing.T) {
    ps := NewGenericPubSub()
    defer func() {
        if recover() == nil {
            t.Fatal("expected panic for illegal wildcard in subscribe")
        }
    }()
    r := &recorder{}
    ps.Subscribe("X", "ap*le", r.handle)
}

func TestPublishIllegalWildcard(t *testing.T) {
    ps := NewGenericPubSub()
    defer func() {
        if recover() == nil {
            t.Fatal("expected panic for wildcard in publish subject")
        }
    }()
    ps.Publish("ap*ple", "oops")
}
