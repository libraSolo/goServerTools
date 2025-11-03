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
