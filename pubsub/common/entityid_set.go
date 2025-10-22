// 本文件定义 EntityID 及其集合类型与常用操作，提供增删查与遍历。
package common

// EntityID 实体唯一标识类型（字符串别名）
type EntityID any

// EntityIDSet 实体 ID 的集合类型
type EntityIDSet map[EntityID]struct{}

// Add 向集合中添加一个实体 ID
func (es EntityIDSet) Add(id EntityID) {
	es[id] = struct{}{}
}

// Del 从集合中删除一个实体 ID
func (es EntityIDSet) Del(id EntityID) {
	delete(es, id)
}

// Contains 判断集合中是否包含指定实体 ID
func (es EntityIDSet) Contains(id EntityID) bool {
	_, ok := es[id]
	return ok
}

// ToList 将集合转换为实体 ID 切片（无序）
func (es EntityIDSet) ToList() []EntityID {
	list := make([]EntityID, 0, len(es))
	for eid := range es {
		list = append(list, eid)
	}
	return list
}

// ForEach 遍历集合（当回调返回 false 时提前结束）
func (es EntityIDSet) ForEach(cb func(eid EntityID) bool) {
	for eid := range es {
		if !cb(eid) {
			break
		}
	}
}
