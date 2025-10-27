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
	*h = append(*h, x.(*Timer))

	
}

func (h *timerHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}