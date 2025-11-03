// TopPlayersHeap 最小堆说明
//
// 该堆按玩家分数从小到大排序（Less 比较 Score），因此堆顶元素是当前前 K 集合中的“最低分”。
// 在维护前 K 名时：
// - 当集合未满，直接 Push；
// - 当集合已满且新分数更高时，弹出堆顶（最低分）再插入新玩家；
// 这样能在 O(log K) 时间维护 TopK 集合，并在读取时近似 O(1) 拿到前 K 个元素（连续内存切片）。
//
// Push/Pop 中包含容量扩缩逻辑：以倍增/减半方式调整底层切片容量，避免频繁分配。
package domain

// TopPlayersHeap 前K名最小堆

const MIN_CAP = 32 // 最小容量
type TopPlayersHeap []*Player

func (h TopPlayersHeap) Len() int           { return len(h) }
func (h TopPlayersHeap) Less(i, j int) bool { return h[i].Score < h[j].Score }
func (h TopPlayersHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *TopPlayersHeap) Push(x interface{}) {
	n := len(*h)

	if n+1 > cap(*h) {
		newCap := max(cap(*h)*2, MIN_CAP)
		newH := make(TopPlayersHeap, n, newCap)
		copy(newH, *h)
		*h = newH
	}

	*h = (*h)[:n+1]
	player := x.(*Player)
	(*h)[n] = player
}

func (h *TopPlayersHeap) Pop() interface{} {
	n := len(*h)
	if n == 0 {
		return nil
	}

	c := cap(*h)
	if n < (c/2) && c > MIN_CAP {
		newCap := c / 2
		if newCap < n {
			newCap = n
		}
		newH := make(TopPlayersHeap, n, newCap)
		copy(newH, *h)
		*h = newH
	}

	x := (*h)[n-1]
	*h = (*h)[:n-1]

	return x
}
