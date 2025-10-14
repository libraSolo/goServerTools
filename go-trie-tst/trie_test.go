package trietst

import "testing"

// TestSubCreatesPathAndStoresVal：验证 Sub 会创建路径，
// 并允许通过返回的节点存取值。
func TestSubCreatesPathAndStoresVal(t *testing.T) {
    var root TrieMO
    node := root.Sub("foo")
    if node == nil {
        t.Fatalf("Sub returned nil")
    }

    node.Val = 123
    sameNode := root.Sub("foo")
    if sameNode != node {
        t.Fatalf("Sub should return the same node for the same path")
    }
    if v, ok := sameNode.Val.(int); !ok || v != 123 {
        t.Fatalf("Val mismatch: got %v, expected 123", sameNode.Val)
    }
}

// TestChildReturnsSameNode：确保对同一字节 Child 返回同一节点，
// 不同字节返回不同节点。
func TestChildReturnsSameNode(t *testing.T) {
    var root TrieMO
    a1 := root.Child('a')
    a2 := root.Child('a')
    if a1 != a2 {
        t.Fatalf("Child('a') should return the same node on repeated calls")
    }
    b := root.Child('b')
    if a1 == b {
        t.Fatalf("Child('a') and Child('b') should return different nodes")
    }
}

// TestSubEmptyReturnsRoot：检查 Sub("") 返回根节点。
func TestSubEmptyReturnsRoot(t *testing.T) {
    var root TrieMO
    n := root.Sub("")
    if n != &root {
        t.Fatalf("Sub(\"\") should return the root node")
    }
}

// TestDeepPath：验证深路径创建正确且彼此独立。
func TestDeepPath(t *testing.T) {
    var root TrieMO
    abc := root.Sub("abc")
    abd := root.Sub("abd")
    if abc == nil || abd == nil {
        t.Fatalf("Sub returned nil for deep path")
    }
    if abc == abd {
        t.Fatalf("Different paths should yield different nodes")
    }

    // 确保中间节点存在且一致
    ab1 := root.Sub("ab")
    ab2 := root.Child('a').Child('b')
    if ab1 != ab2 {
        t.Fatalf("Intermediate nodes via Sub and Child chain should be identical")
    }
}

// TestSeparateValues：确保不同路径上的值不会互相覆盖。
func TestSeparateValues(t *testing.T) {
    var root TrieMO
    root.Sub("ab").Val = "v1"
    root.Sub("ac").Val = "v2"

    if v := root.Sub("ab").Val; v != "v1" {
        t.Fatalf("Value on path 'ab' mismatch: %v", v)
    }
    if v := root.Sub("ac").Val; v != "v2" {
        t.Fatalf("Value on path 'ac' mismatch: %v", v)
    }
}
