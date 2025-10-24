package timeWheel

import (
	"container/list"
	"errors"
	"log"
	"sync"
	"time"
)

// TimeWheel 多层时间轮结构定义
type TimeWheel struct {
	interval    time.Duration               // 当前时间轮的刻度间隔
	slots       []*list.List                // 时间轮槽位
	timer       map[interface{}]*timerEntry // 任务索引映射（仅最底层使用）
	currentPos  int                         // 当前指针位置
	slotNum     int                         // 槽位总数
	taskMutex   sync.RWMutex                // 读写锁（仅最底层使用）
	higherLevel *TimeWheel                  // 指向上层时间轮（更慢）
	lowerLevel  *TimeWheel                  // 指向下层时间wheel（更快）
	level       int                         // 当前层级（1为最底层）
	root        *TimeWheel                  // 指向最底层时间轮的指针

	// 以下字段仅在 root 时间轮中有效
	stopChannel chan struct{}
	running     bool
	maxLevels   int
}

// 定时器条目
type timerEntry struct {
	task      *Task
	slotIndex int
	element   *list.Element
}

// 任务结构
type Task struct {
	delay  time.Duration // 任务延迟时间
	key    interface{}   // 任务唯一标识
	job    func()        // 任务执行函数
	circle int           // 任务需要在当前轮转多少圈
	level  int           // 任务所在的层级
}

// New 创建时间轮实例
// interval: 最底层时间轮的刻度间隔
// slotNum: 每层的槽位数
// maxLevels: 允许的最大层级数
func New(interval time.Duration, slotNum int, maxLevels int) (*TimeWheel, error) {
	if interval <= 0 || slotNum <= 0 || maxLevels <= 0 {
		return nil, errors.New("interval, slotNum, and maxLevels must be greater than zero")
	}

	// 只创建最底层时间轮
	root := &TimeWheel{
		interval:    interval,
		slots:       make([]*list.List, slotNum),
		timer:       make(map[interface{}]*timerEntry),
		currentPos:  0,
		slotNum:     slotNum,
		stopChannel: make(chan struct{}),
		level:       1,
		maxLevels:   maxLevels,
	}
	root.root = root // 指向自己

	for i := 0; i < slotNum; i++ {
		root.slots[i] = list.New()
	}

	return root, nil
}

// Start 启动时间轮（只启动最底层）
func (tw *TimeWheel) Start() {
	tw.root.taskMutex.Lock()
	if tw.root.running {
		tw.root.taskMutex.Unlock()
		return
	}
	tw.root.running = true
	ticker := time.NewTicker(tw.root.interval)
	go tw.root.run(ticker)
	tw.root.taskMutex.Unlock()
	log.Printf("时间轮启动，底层间隔: %v", tw.root.interval)
}

// Stop 停止时间轮
func (tw *TimeWheel) Stop() {
	tw.root.taskMutex.Lock()
	if !tw.root.running {
		tw.root.taskMutex.Unlock()
		return
	}
	tw.root.running = false
	tw.root.taskMutex.Unlock()

	close(tw.root.stopChannel)
	log.Println("时间轮已停止")
}

// run 仅在最底层时间轮运行
func (tw *TimeWheel) run(ticker *time.Ticker) {
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tw.tickHandler()
		case <-tw.stopChannel:
			return
		}
	}
}

// tickHandler 时间刻度处理（仅最底层调用）
func (tw *TimeWheel) tickHandler() {
	tw.root.taskMutex.Lock()
	tw.currentPos = (tw.currentPos + 1) % tw.slotNum
	slot := tw.slots[tw.currentPos]
	tw.root.taskMutex.Unlock()

	// 执行当前槽位的任务
	tw.executeSlotTasks(slot)

	// 如果是第0个槽，说明转完一圈，需要向上层传递tick
	if tw.currentPos == 0 && tw.higherLevel != nil {
		tw.higherLevel.advance()
	}
}

// advance 手动推进上层时间轮
func (tw *TimeWheel) advance() {
	tw.root.taskMutex.Lock()
	tw.currentPos = (tw.currentPos + 1) % tw.slotNum
	slot := tw.slots[tw.currentPos]
	tw.root.taskMutex.Unlock()

	// 降级当前槽位的任务
	tw.demoteSlotTasks(slot)

	// 如果上层也转完一圈，继续向上传递
	if tw.currentPos == 0 && tw.higherLevel != nil {
		tw.higherLevel.advance()
	}
}

// executeSlotTasks 执行最底层槽位的任务
func (tw *TimeWheel) executeSlotTasks(slot *list.List) {
	tw.root.taskMutex.Lock()
	defer tw.root.taskMutex.Unlock()

	for e := slot.Front(); e != nil; {
		task := e.Value.(*Task)
		next := e.Next()

		if task.circle > 0 {
			task.circle--
			e = next
			continue
		}

		// 在锁内执行任务以保证状态一致性
		go task.job()
		log.Printf("任务执行: key=%v", task.key)

		slot.Remove(e)
		delete(tw.root.timer, task.key)
		e = next
	}
}

// demoteSlotTasks 降级上层槽位的任务
func (tw *TimeWheel) demoteSlotTasks(slot *list.List) {
	tw.root.taskMutex.Lock()
	defer tw.root.taskMutex.Unlock()

	for e := slot.Front(); e != nil; {
		task := e.Value.(*Task)
		next := e.Next()

		if task.circle > 0 {
			task.circle--
			e = next
			continue
		}

		// 任务到期，需要降级
		// 计算剩余延迟并重新添加到根轮
		remainingDelay := task.delay % (tw.interval)
		tw.root.AddTask(remainingDelay, task.key, task.job)

		log.Printf("任务降级: key=%v 从层级 %d 到下层", task.key, tw.level)

		slot.Remove(e)
		delete(tw.root.timer, task.key)
		e = next
	}
}

// AddTask 添加一个新任务
func (tw *TimeWheel) AddTask(delay time.Duration, key interface{}, job func()) error {
	if delay < 0 {
		return errors.New("delay must be non-negative")
	}

	task := &Task{delay: delay, key: key, job: job}

	// 必须在根节点的锁下执行，因为可能创建新层级
	tw.root.taskMutex.Lock()
	defer tw.root.taskMutex.Unlock()

	// 先删除旧任务（如果存在）
	tw.root.removeTaskInternal(key)

	return tw.root.addTask(task)
}

// addTask 负责找到合适的轮并添加任务（必须在锁内调用）
func (tw *TimeWheel) addTask(task *Task) error {
	current := tw // 从根开始

	// 寻找合适的层级，如果层级不够则创建
	for task.delay >= current.interval*time.duration(current.slotNum) {
		if current.higherLevel == nil {
			if current.level >= current.root.maxLevels {
				log.Printf("任务延迟 %v 过大，已超出最大层级 %d 的范围", task.delay, current.root.maxLevels)
				// 可以选择报错或将其放入最高层
				break
			}
			log.Printf("创建新层级: %d", current.level+1)
			newHigherLevel := &TimeWheel{
				interval:   current.interval * time.Duration(current.slotNum),
				slots:      make([]*list.List, current.slotNum),
				currentPos: 0,
				slotNum:    current.slotNum,
				level:      current.level + 1,
				lowerLevel: current,
				root:       tw.root, // 共享 root
			}
			for i := 0; i < newHigherLevel.slotNum; i++ {
				newHigherLevel.slots[i] = list.New()
			}
			current.higherLevel = newHigherLevel
		}
		current = current.higherLevel
	}

	// 在找到的层级中添加任务
	return current.addTaskInternal(task)
}

// addTaskInternal 将任务添加到当前时间轮（必须在锁内调用）
func (tw *TimeWheel) addTaskInternal(task *Task) error {
	delay := task.delay
	if delay < tw.interval {
		delay = tw.interval // 至少延迟一个刻度
	}

	ticks := int64(delay / tw.interval)
	circle := int(ticks / int64(tw.slotNum))
	pos := (tw.currentPos + int(ticks)) % tw.slotNum
	task.circle = circle
	task.level = tw.level // 记录任务所在的层级

	element := tw.slots[pos].PushBack(task)
	// 所有任务索引都存储在 root timer 中
	tw.root.timer[task.key] = &timerEntry{
		task:      task,
		slotIndex: pos,
		element:   element,
	}

	log.Printf("添加任务: key=%v, 延迟=%v, 层级=%d, 位置=%d, 圈数=%d",
		task.key, task.delay, tw.level, pos, circle)
	return nil
}

// RemoveTask 删除任务
func (tw *TimeWheel) RemoveTask(key interface{}) bool {
	tw.root.taskMutex.Lock()
	defer tw.root.taskMutex.Unlock()
	return tw.root.removeTaskInternal(key)
}

// removeTaskInternal 直接在 root.timer 中查找并删除任务（必须在锁内调用）
func (tw *TimeWheel) removeTaskInternal(key interface{}) bool {
	// 从 root 的 timer map 中查找任务
	if entry, exists := tw.root.timer[key]; exists {
		// 从任务所在的层级的槽位中移除
		// 注意：entry.task.wheel.slots[entry.slotIndex].Remove(entry.element) 这种方式更理想，但需要给Task增加wheel指针
		// 为了简化，我们先找到对应的wheel
		wheel := tw.findWheelByLevel(entry.task.level)
		if wheel != nil {
			wheel.slots[entry.slotIndex].Remove(entry.element)
			delete(tw.root.timer, key)
			log.Printf("删除任务: key=%v, 层级=%d", key, entry.task.level)
			return true
		}
	}
	return false
}

// findWheelByLevel 根据层级找到对应的 wheel 指针
func (tw *TimeWheel) findWheelByLevel(level int) *TimeWheel {
	current := tw.root
	for current != nil && current.level != level {
		current = current.higherLevel
	}
	return current
}
