package crontab

// timerHeap 是一个定时器的最小堆。
type timerHeap []*Timer

func (h timerHeap) Len() int {
	return len(h)
}

func (h timerHeap) Less(i, j int) bool {
	t1, t2 := h[i].fireTime, h[j].fireTime
	if t1.Before(t2) {
		return true
	}
	if t1.After(t2) {
		return false
	}
	// 具有相同截止时间的定时器按其添加顺序触发。
	return h[i].addSeq < h[j].addSeq
}

func (h timerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *timerHeap) Push(x interface{}) {
	n := len(*h)

	// 动态扩容
	if n + 1 > cap(*h) {
		newCap := max(cap(*h)*2, 8)
		newH := make(timerHeap, n, newCap)
		copy(newH, *h)
		*h = newH
	}

	*h = (*h)[:n+1]
	item := x.(*Timer)
	(*h)[n] = item
}

func (h *timerHeap) Pop() interface{} {
	n := len(*h)
	if n == 0 {
		return nil
	}

	c := cap(*h)

	// 动态收缩
	if n < (c / 2) && c > 25 {
		newCap := c / 2
		if newCap < n {
			newCap = n
		}
		newH := make(timerHeap, n, newCap)
		copy(newH, *h)
		*h = newH
	}

	x := (*h)[n - 1]
	*h = (*h)[:n - 1]
	return x
}