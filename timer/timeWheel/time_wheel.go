package timeWheel

import (
	"container/list"
	"errors"
	"log"
	"math"
	"sync"
	"time"
)

// TimeWheel 结构定义
type TimeWheel struct {
	interval          time.Duration
	ticker            *time.Ticker
	slots             []*list.List
	timer             map[interface{}]*list.Element
	currentPos        int
	slotNum           int
	addTaskChannel    chan Task
	removeTaskChannel chan interface{}
	stopChannel       chan bool
	taskMutex         sync.Mutex
	nextLevel         *TimeWheel // 指向下一层（更快）的时间轮
	level             int        // 当前时间轮的层级（慢->快：level -> 1）
}

// Task 结构定义
type Task struct {
	delay  time.Duration
	circle int
	key    interface{}
	job    func()
}

// New 创建一个新的层级时间轮。interval 是最快轮的刻度时间。
func New(interval time.Duration, slotNum int, level int) (*TimeWheel, error) {
	if interval <= 0 || slotNum <= 0 || level <= 0 {
		return nil, errors.New("interval, slotNum, or level must be greater than zero")
	}

	wheels := make([]*TimeWheel, level)
	// 从最快到最慢创建所有轮
	for i := 0; i < level; i++ {
		currentInterval := interval * time.Duration(math.Pow(float64(slotNum), float64(i)))
		wheels[i] = newTimeWheel(currentInterval, slotNum, level-i)
	}

	// 从最慢到最快连接它们
	for i := level - 1; i > 0; i-- {
		wheels[i].nextLevel = wheels[i-1]
	}

	// 返回最慢的轮作为根
	return wheels[level-1], nil
}

func newTimeWheel(interval time.Duration, slotNum int, level int) *TimeWheel {
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List, slotNum),
		timer:             make(map[interface{}]*list.Element),
		currentPos:        0,
		slotNum:           slotNum,
		addTaskChannel:    make(chan Task, 1024),
		removeTaskChannel: make(chan interface{}),
		stopChannel:       make(chan bool),
		level:             level,
	}
	for i := 0; i < slotNum; i++ {
		tw.slots[i] = list.New()
	}
	return tw
}

func (tw *TimeWheel) Start() {
	// 启动所有层级的轮
	current := tw
	for current != nil {
		current.ticker = time.NewTicker(current.interval)
		go current.start()
		current = current.nextLevel
	}
}

func (tw *TimeWheel) Stop() {
	current := tw
	for current != nil {
		current.stopChannel <- true
		current = current.nextLevel
	}
}

func (tw *TimeWheel) AddTask(delay time.Duration, key interface{}, job func()) {
	if delay < 0 {
		return
	}
	tw.addTaskChannel <- Task{delay: delay, key: key, job: job}
}

func (tw *TimeWheel) RemoveTask(key interface{}) {
	if key == nil {
		return
	}
	tw.removeTaskChannel <- key
}

func (tw *TimeWheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.addTask(&task)
		case key := <-tw.removeTaskChannel:
			tw.removeTask(key)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimeWheel) tickHandler() {
	tw.currentPos = (tw.currentPos + 1) % tw.slotNum
	l := tw.slots[tw.currentPos]
	tw.scanAndRunTask(l)
}

func (tw *TimeWheel) scanAndRunTask(l *list.List) {
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}

		next := e.Next()
		l.Remove(e)

		if tw.nextLevel != nil {
			// 重新计算延迟并交给更快的轮处理
			task.delay = task.delay % tw.interval
			tw.nextLevel.addTask(task)
		} else {
			// 最快的轮，直接执行
			go task.job()
			tw.taskMutex.Lock()
			delete(tw.timer, task.key)
			tw.taskMutex.Unlock()
		}
		e = next
	}
}

func (tw *TimeWheel) addTask(task *Task) {
	delay := task.delay

	if tw.nextLevel != nil && delay < tw.interval {
		tw.nextLevel.addTask(task)
		return
	}

	pos, circle := tw.getPositionAndCircle(delay)
	task.circle = circle

	elem := tw.slots[pos].PushBack(task)
	tw.taskMutex.Lock()
	tw.timer[task.key] = elem
	tw.taskMutex.Unlock()
	log.Printf("add task key: %v, pos: %d, circle: %d, level: %d", task.key, pos, circle, tw.level)
}

func (tw *TimeWheel) removeTask(key interface{}) {
	tw.taskMutex.Lock()
	defer tw.taskMutex.Unlock()

	if elem, ok := tw.timer[key]; ok {
		// 由于无法从 list.Element 直接找到其所属的 list，
		// 我们需要遍历所有 slots 来找到并删除它。
		// 这是一个性能瓶颈，但在 container/list 中难以避免。
		// 更好的实现可能需要自定义链表。
		for _, l := range tw.slots {
			// 尝试从每个 list 中删除，只有一个会成功
			l.Remove(elem)
		}
		delete(tw.timer, key)
		log.Printf("removed task key: %v from level %d", key, tw.level)
	} else if tw.nextLevel != nil {
		// 如果当前轮没有，去更快的轮寻找
		tw.nextLevel.removeTask(key)
	}
}

func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	if d < 0 {
		d = 0
	}
	ticks := int64(d / tw.interval)
	circle = int(ticks / int64(tw.slotNum))
	pos = int((int64(tw.currentPos) + ticks) % int64(tw.slotNum))
	return
}
