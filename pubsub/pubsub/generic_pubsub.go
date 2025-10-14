package pubsub // 通用发布订阅：与引擎解耦的主题发布/订阅实现（支持末尾通配 '*')

import (
    "common"
    "sync"
    "trietst"
)

// Handler 为订阅者的回调函数类型
// 当匹配到发布的主题时，服务将调用对应订阅者的 Handler
type Handler func(subject string, content string)

// subscribing 表示某主题前缀的订阅集合（精确+通配）
type subscribing struct {
    subscribers         common.StringSet // 精确订阅该主题的订阅者
    wildcardSubscribers common.StringSet // 通配订阅该主题前缀（subject + '*') 的订阅者
}

func newSubscribing() *subscribing {
    return &subscribing{
        subscribers:         common.StringSet{},
        wildcardSubscribers: common.StringSet{},
    }
}

// GenericPubSub 为通用发布订阅服务
// - 支持精确主题订阅与前缀通配订阅（仅允许末尾 '*')
// - 发布时按前缀树逐层匹配通配订阅、末端匹配精确订阅
// - 每个订阅者使用唯一的 Handler 接收消息
type GenericPubSub struct {
    mu   sync.RWMutex
    tree trietst.TrieMO

    // 记录订阅者的已订阅主题（便于取消订阅 / 取消所有）
    subscriberExactSubjects    map[string]common.StringSet
    subscriberWildcardSubjects map[string]common.StringSet
    subscriberHandlers         map[string]Handler
}

// NewGenericPubSub 创建一个新的通用发布订阅服务实例
func NewGenericPubSub() *GenericPubSub {
    return &GenericPubSub{
        subscriberExactSubjects:    map[string]common.StringSet{},
        subscriberWildcardSubjects: map[string]common.StringSet{},
        subscriberHandlers:         map[string]Handler{},
    }
}

// Subscribe 订阅主题（支持末尾通配 '*')
// 规则：
// - '*' 仅允许出现在主题末尾，且最多一次
// - 主题可为空，形如 "*" 表示订阅所有主题（匹配任意前缀）
// - 每个订阅者仅保存一个 Handler，重复订阅将更新该订阅者的 Handler
func (ps *GenericPubSub) Subscribe(subscriberID string, subject string, handler Handler) {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    // 校验 '*' 位置
    for i, c := range subject {
        if c == '*' && i != len(subject)-1 {
            panic("'*' can only be used at the end of subject while subscribing")
        }
    }

    // 更新订阅者的 Handler（最后一次订阅生效）
    if handler != nil {
        ps.subscriberHandlers[subscriberID] = handler
    }

    wildcard := false
    if subject != "" && subject[len(subject)-1] == '*' {
        wildcard = true
        subject = subject[:len(subject)-1]
    }

    subs := ps.getSubscribing(subject, true)
    if !wildcard {
        subs.subscribers.Add(subscriberID)
        exactSet := ps.subscriberExactSubjects[subscriberID]
        if exactSet == nil {
            exactSet = common.StringSet{}
            ps.subscriberExactSubjects[subscriberID] = exactSet
        }
        exactSet.Add(subject)
    } else {
        subs.wildcardSubscribers.Add(subscriberID)
        wildcardSet := ps.subscriberWildcardSubjects[subscriberID]
        if wildcardSet == nil {
            wildcardSet = common.StringSet{}
            ps.subscriberWildcardSubjects[subscriberID] = wildcardSet
        }
        wildcardSet.Add(subject)
    }
}

// Unsubscribe 取消订阅主题（支持末尾通配 '*')
func (ps *GenericPubSub) Unsubscribe(subscriberID string, subject string) {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    for i, c := range subject {
        if c == '*' && i != len(subject)-1 {
            panic("'*' can only be used at the end of subject while unsubscribing")
        }
    }

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
        if exactSet := ps.subscriberExactSubjects[subscriberID]; exactSet != nil {
            exactSet.Remove(subject)
        }
    } else {
        subs.wildcardSubscribers.Remove(subscriberID)
        if wildcardSet := ps.subscriberWildcardSubjects[subscriberID]; wildcardSet != nil {
            wildcardSet.Remove(subject)
        }
    }
}

// UnsubscribeAll 取消该订阅者的所有订阅（包括精确和通配）
func (ps *GenericPubSub) UnsubscribeAll(subscriberID string) {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    if exactSet, ok := ps.subscriberExactSubjects[subscriberID]; ok {
        delete(ps.subscriberExactSubjects, subscriberID)
        for subject := range exactSet {
            subs := ps.getSubscribing(subject, false)
            if subs != nil {
                subs.subscribers.Remove(subscriberID)
            }
        }
    }
    if wildcardSet, ok := ps.subscriberWildcardSubjects[subscriberID]; ok {
        delete(ps.subscriberWildcardSubjects, subscriberID)
        for subject := range wildcardSet {
            subs := ps.getSubscribing(subject, false)
            if subs != nil {
                subs.wildcardSubscribers.Remove(subscriberID)
            }
        }
    }
}

// Publish 发布主题与内容（主题中不允许出现 '*')
func (ps *GenericPubSub) Publish(subject string, content string) {
    ps.mu.RLock()
    defer ps.mu.RUnlock()

    for _, c := range subject {
        if c == '*' {
            panic("subject should not contains '*' while publishing")
        }
    }

    ps.publishInTree(subject, content, &ps.tree, 0)
}

// 递归沿前缀树匹配并调用订阅者
func (ps *GenericPubSub) publishInTree(subject string, content string, st *trietst.TrieMO, idx int) {
    subs := ps.getSubscribingOfTree(st, false)
    if subs != nil {
        // 调用当前层的通配订阅者（prefix+'*')
        for subscriberID := range subs.wildcardSubscribers {
            if h := ps.subscriberHandlers[subscriberID]; h != nil {
                h(subject, content)
            }
        }
    }
    if idx < len(subject) {
        ps.publishInTree(subject, content, st.Child(subject[idx]), idx+1)
    } else {
        // 精确匹配到叶子节点，调用精确订阅者
        if subs != nil {
            for subscriberID := range subs.subscribers {
                if h := ps.subscriberHandlers[subscriberID]; h != nil {
                    h(subject, content)
                }
            }
        }
    }
}

// 获取指定主题的订阅集合（按需创建）
func (ps *GenericPubSub) getSubscribing(subject string, newIfNotExists bool) *subscribing {
    t := ps.tree.Sub(subject)
    return ps.getSubscribingOfTree(t, newIfNotExists)
}

// 从树节点取订阅集合
func (ps *GenericPubSub) getSubscribingOfTree(t *trietst.TrieMO, newIfNotExists bool) *subscribing {
    var subs *subscribing
    if t.Val == nil {
        if newIfNotExists {
            subs = newSubscribing()
            t.Val = subs
        }
    } else {
        subs = t.Val.(*subscribing)
    }
    return subs
}
