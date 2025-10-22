package trietst

// Trie 是一个可变前缀树节点（Mutable Object Trie），
// 提供对任意字节序列的逐层访问与存储能力。
//
// 设计要点：
// - Val：节点挂载的任意值（由使用方决定其类型），可为 nil。
// - Child(b)：返回指定字节的子节点；若不存在则懒创建。
// - Sub(s)：返回指定前缀路径对应的子树；路径上节点若不存在则逐层懒创建。
type Trie struct {
	Val      interface{}
	children map[byte]*Trie
}

// NewTrie 创建一个新的可变前缀树节点。
func NewTrie() *Trie { return &Trie{} }

// Child 返回（并在必要时创建）该节点在字节 b 上的子节点。
func (t *Trie) Child(b byte) *Trie {
	if t.children == nil {
		t.children = make(map[byte]*Trie)
	}
	if child := t.children[b]; child != nil {
		return child
	}
	child := &Trie{}
	t.children[b] = child
	return child
}

// ChildIfExists 返回该节点在字节 b 上的子节点（若不存在则返回 nil，不创建）。
func (t *Trie) ChildIfExists(b byte) *Trie {
	if t.children == nil {
		return nil
	}
	return t.children[b]
}

// HasChild 检查在字节 b 上是否存在子节点。
func (t *Trie) HasChild(b byte) bool {
	if t.children == nil {
		return false
	}
	_, ok := t.children[b]
	return ok
}

// Sub 沿着字符串 s 的每个字节逐层访问并返回对应子树节点。
// 若路径上的节点不存在，则在访问过程中懒创建。
func (t *Trie) Sub(s string) *Trie {
	node := t
	for i := 0; i < len(s); i++ {
		node = node.Child(s[i])
	}
	return node
}

// Find 沿着字符串 s 的每个字节逐层访问并返回对应子树节点。
// 若路径上的节点不存在，返回 nil（不进行创建）。
func (t *Trie) Find(s string) *Trie {
	node := t
	for i := 0; i < len(s); i++ {
		if node == nil {
			return nil
		}
		node = node.ChildIfExists(s[i])
	}
	return node
}

// Keys 返回当前节点所有子分支的字节列表（无序）。
func (t *Trie) Keys() []byte {
	if t.children == nil {
		return nil
	}
	keys := make([]byte, 0, len(t.children))
	for b := range t.children {
		keys = append(keys, b)
	}
	return keys
}

// DeleteChild 删除该节点在字节 b 上的子节点（若不存在则忽略）。
func (t *Trie) DeleteChild(b byte) {
	if t.children == nil {
		return
	}
	delete(t.children, b)
}

// Size 计算从当前节点出发（包含自身）的节点总数。
// 注意：map 遍历无序，本方法仅用于粗略度量与监控统计。
func (t *Trie) Size() int {
	if t == nil {
		return 0
	}
	cnt := 1 // 自身
	if t.children != nil {
		for _, c := range t.children {
			cnt += c.Size()
		}
	}
	return cnt
}

// Walk 对从当前节点出发的整棵子树进行深度优先遍历。
// visit 的 path 参数为沿路径累积的字符串；返回 false 将终止遍历。
func (t *Trie) Walk(visit func(path string, node *Trie) bool) {
	if t == nil || visit == nil {
		return
	}
	// 根节点路径为空字符串
	if !visit("", t) {
		return
	}
	t.walkInternal("", visit)
}

// WalkFrom 从指定前缀路径开始进行深度优先遍历。
// 若前缀不存在则不做任何操作。
func (t *Trie) WalkFrom(prefix string, visit func(path string, node *Trie) bool) {
	if visit == nil {
		return
	}
	start := t.Find(prefix)
	if start == nil {
		return
	}
	// 对起点先调用
	if !visit(prefix, start) {
		return
	}
	start.walkInternal(prefix, visit)
}

// 内部递归遍历实现。
func (t *Trie) walkInternal(prefix string, visit func(path string, node *Trie) bool) {
	if t.children == nil {
		return
	}
	for b, child := range t.children {
		p := prefix + string([]byte{b})
		if !visit(p, child) {
			continue
		}
		child.walkInternal(p, visit)
	}
}
