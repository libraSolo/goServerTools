// 本文件提供通用集合类型及其操作，包含字符串集合、整型集合等。
package common

// StringSet is a set of strings
// StringSet 字符串集合类型
type StringSet map[string]struct{}

// Contains checks if Stringset contains the string
// Contains 判断集合是否包含指定字符串
func (ss StringSet) Contains(elem string) bool {
    _, ok := ss[elem]
    return ok
}

// Add adds the string to StringSet
// Add 将字符串加入集合
func (ss StringSet) Add(elem string) {
    ss[elem] = struct{}{}
}

// Remove removes the string from StringList
// Remove 从集合移除字符串
func (ss StringSet) Remove(elem string) {
    delete(ss, elem)
}

// ToList convert StringSet to string slice
// ToList 将集合转换为切片（无序）
func (ss StringSet) ToList() []string {
    keys := make([]string, 0, len(ss))
    for s := range ss {
        keys = append(keys, s)
    }
    return keys
}

// StringList is a list of string (slice)
// StringList 字符串切片类型
type StringList []string

// Remove removes the string from StringList
// Remove 从列表中移除指定字符串（原地压缩）
func (sl *StringList) Remove(elem string) {
    widx := 0
    cpsl := *sl
    for idx, _elem := range cpsl {
        if _elem == elem {
            // ignore this elem by doing nothing
        } else {
            if idx != widx {
                cpsl[widx] = _elem
            }
            widx += 1
        }
    }

    *sl = cpsl[:widx]
}

// Append add the string to the end of StringList
// Append 将字符串追加到列表末尾
func (sl *StringList) Append(elem string) {
    *sl = append(*sl, elem)
}

// Find get the index of string in StringList, returns -1 if not found
// Find 查找字符串在列表中的索引，未找到返回 -1
func (sl *StringList) Find(s string) int {
    for idx, elem := range *sl {
        if elem == s {
            return idx
        }
    }
    return -1
}

// IntSet is a set of int
// IntSet 整型集合类型
type IntSet map[int]struct{}

// Contains checks if Stringset contains the string
// Contains 判断集合是否包含指定整数
func (is IntSet) Contains(elem int) bool {
    _, ok := is[elem]
    return ok
}

// Add adds the string to IntSet
// Add 将整数加入集合
func (is IntSet) Add(elem int) {
    is[elem] = struct{}{}
}

// Remove removes the string from IntSet
// Remove 从集合移除整数
func (is IntSet) Remove(elem int) {
    delete(is, elem)
}

// ToList convert IntSet to int slice
// ToList 将集合转换为切片（无序）
func (is IntSet) ToList() []int {
    keys := make([]int, 0, len(is))
    for s := range is {
        keys = append(keys, s)
    }
    return keys
}

// Uint16Set is a set of int
// Uint16Set 无符号 16 位整数集合类型
type Uint16Set map[uint16]struct{}

// Contains checks if Stringset contains the string
// Contains 判断集合是否包含指定 uint16
func (is Uint16Set) Contains(elem uint16) bool {
    _, ok := is[elem]
    return ok
}

// Add adds the string to Uint16Set
// Add 将 uint16 加入集合
func (is Uint16Set) Add(elem uint16) {
    is[elem] = struct{}{}
}

// Remove removes the string from Uint16Set
// Remove 从集合移除 uint16
func (is Uint16Set) Remove(elem uint16) {
    delete(is, elem)
}

// ToList convert Uint16Set to int slice
// ToList 将集合转换为切片（无序）
func (is Uint16Set) ToList() []uint16 {
    keys := make([]uint16, 0, len(is))
    for s := range is {
        keys = append(keys, s)
    }
    return keys
}
