package pubsub

import (
	"common"
	"fmt"
	"sync"
	"trietst"
)

// Handler 为泛型订阅者的回调函数类型
type Handler[T any] func(subject string, content T)

// subscribing 表示某主题前缀的订阅集合
type subscribing struct {
	subscribers         common.StringSet
	wildcardSubscribers common.StringSet
}

func newSubscribing() *subscribing {
	return &subscribing{
		subscribers:         common.StringSet{},
		wildcardSubscribers: common.StringSet{},
	}
}

// GenericPubSub 为通用发布订阅服务（泛型版）
type GenericPubSub[T any] struct {
	mu   sync.RWMutex
	tree trietst.Trie

	subscriberExactSubjects    map[string]common.StringSet
	subscriberWildcardSubjects map[string]common.StringSet
	subscriberHandlers         map[string]Handler[T]
}

// NewGenericPubSub 创建一个新的通用发布订阅服务实例
func NewGenericPubSub[T any]() *GenericPubSub[T] {
	return &GenericPubSub[T]{
		subscriberExactSubjects:    map[string]common.StringSet{},
		subscriberWildcardSubjects: map[string]common.StringSet{},
		subscriberHandlers:         map[string]Handler[T]{},
	}
}

// Subscribe 订阅主题，返回错误而不是 panic
func (ps *GenericPubSub[T]) Subscribe(subscriberID string, subject string, handler Handler[T]) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subscriberID == "" {
		return fmt.Errorf("subscriberID cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	for i, c := range subject {
		if c == '*' && i != len(subject)-1 {
			return fmt.Errorf("'*' can only be used at the end of subject")
		}
	}

	ps.subscriberHandlers[subscriberID] = handler

	wildcard := false
	if subject != "" && subject[len(subject)-1] == '*' {
		wildcard = true
		subject = subject[:len(subject)-1]
	}

	subs := ps.getSubscribing(subject, true)
	if !wildcard {
		subs.subscribers.Add(subscriberID)
		exactSet, ok := ps.subscriberExactSubjects[subscriberID]
		if !ok {
			exactSet = common.StringSet{}
			ps.subscriberExactSubjects[subscriberID] = exactSet
		}
		exactSet.Add(subject)
	} else {
		subs.wildcardSubscribers.Add(subscriberID)
		wildcardSet, ok := ps.subscriberWildcardSubjects[subscriberID]
		if !ok {
			wildcardSet = common.StringSet{}
			ps.subscriberWildcardSubjects[subscriberID] = wildcardSet
		}
		wildcardSet.Add(subject)
	}
	return nil
}

// Unsubscribe 取消订阅
func (ps *GenericPubSub[T]) Unsubscribe(subscriberID string, subject string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	wildcard := false
	if subject != "" && subject[len(subject)-1] == '*' {
		wildcard = true
		subject = subject[:len(subject)-1]
	}

	subs := ps.getSubscribing(subject, false)
	if subs == nil {
		return
	}

	if !wildcard {
		subs.subscribers.Remove(subscriberID)
		if exactSet, ok := ps.subscriberExactSubjects[subscriberID]; ok {
			exactSet.Remove(subject)
		}
	} else {
		subs.wildcardSubscribers.Remove(subscriberID)
		if wildcardSet, ok := ps.subscriberWildcardSubjects[subscriberID]; ok {
			wildcardSet.Remove(subject)
		}
	}
}

// UnsubscribeAll 取消该订阅者的所有订阅
func (ps *GenericPubSub[T]) UnsubscribeAll(subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if exactSet, ok := ps.subscriberExactSubjects[subscriberID]; ok {
		delete(ps.subscriberExactSubjects, subscriberID)
		for subject := range exactSet {
			if subs := ps.getSubscribing(subject, false); subs != nil {
				subs.subscribers.Remove(subscriberID)
			}
		}
	}

	if wildcardSet, ok := ps.subscriberWildcardSubjects[subscriberID]; ok {
		delete(ps.subscriberWildcardSubjects, subscriberID)
		for subject := range wildcardSet {
			if subs := ps.getSubscribing(subject, false); subs != nil {
				subs.wildcardSubscribers.Remove(subscriberID)
			}
		}
	}
}

// Publish 发布主题与内容，返回错误而不是 panic
func (ps *GenericPubSub[T]) Publish(subject string, content T) error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, c := range subject {
		if c == '*' {
			return fmt.Errorf("subject should not contain '*' while publishing")
		}
	}

	ps.publishInTree(subject, content, &ps.tree, 0)
	return nil
}

// 递归发布
func (ps *GenericPubSub[T]) publishInTree(subject string, content T, st *trietst.Trie, idx int) {
	if subs := ps.getSubscribingOfTree(st, false); subs != nil {
		for subscriberID := range subs.wildcardSubscribers {
			if h, ok := ps.subscriberHandlers[subscriberID]; ok {
				h(subject, content)
			}
		}
	}

	if idx < len(subject) {
		ps.publishInTree(subject, content, st.Child(subject[idx]), idx+1)
	} else {
		if subs := ps.getSubscribingOfTree(st, false); subs != nil {
			for subscriberID := range subs.subscribers {
				if h, ok := ps.subscriberHandlers[subscriberID]; ok {
					h(subject, content)
				}
			}
		}
	}
}

// 获取订阅集合
func (ps *GenericPubSub[T]) getSubscribing(subject string, create bool) *subscribing {
	t := ps.tree.Sub(subject)
	return ps.getSubscribingOfTree(t, create)
}

// 从树节点获取订阅集合
func (ps *GenericPubSub[T]) getSubscribingOfTree(t *trietst.Trie, create bool) *subscribing {
	if t.Val == nil {
		if create {
			subs := newSubscribing()
			t.Val = subs
			return subs
		}
		return nil
	}
	return t.Val.(*subscribing)
}
