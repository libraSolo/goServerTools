package trietst

// TrieMO 是一个可变前缀树节点（Mutable Object Trie），
// 提供对任意字节序列的逐层访问与存储能力。
//
// 设计要点：
// - Val：节点挂载的任意值（由使用方决定其类型），可为 nil。
// - Child(b)：返回指定字节的子节点；若不存在则懒创建。
// - Sub(s)：返回指定前缀路径对应的子树；路径上节点若不存在则逐层懒创建。
type TrieMO struct {
    Val      interface{}
    children map[byte]*TrieMO
}

// Child 返回（并在必要时创建）该节点在字节 b 上的子节点。
func (t *TrieMO) Child(b byte) *TrieMO {
    if t.children == nil {
        t.children = make(map[byte]*TrieMO)
    }
    if child := t.children[b]; child != nil {
        return child
    }
    child := &TrieMO{}
    t.children[b] = child
    return child
}

// Sub 沿着字符串 s 的每个字节逐层访问并返回对应子树节点。
// 若路径上的节点不存在，则在访问过程中懒创建。
func (t *TrieMO) Sub(s string) *TrieMO {
    node := t
    for i := 0; i < len(s); i++ {
        node = node.Child(s[i])
    }
    return node
}