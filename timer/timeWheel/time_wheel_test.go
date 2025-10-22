package timeWheel

import (
    "fmt"
    "testing"
    "time"
)

func TestHierarchicalTimeWheel(t *testing.T) {
    // 创建一个3层的层级时间轮
    // level 1: 1s/slot, 10 slots
    // level 2: 10s/slot, 10 slots
    // level 3: 100s/slot, 10 slots
    tw, err := New(1*time.Second, 10, 3)
    if err != nil {
        t.Fatalf("failed to create time wheel: %v", err)
    }
    tw.Start()
    defer tw.Stop()

    // 添加一个短延时任务，应在第一层执行
    tw.AddTask(3*time.Second, "task1", func() {
        fmt.Println("Task 1 (3s) executed")
    })

    // 添加一个中等延时任务，应在第二层流转后执行
    tw.AddTask(15*time.Second, "task2", func() {
        fmt.Println("Task 2 (15s) executed")
    })

    // 添加一个长延时任务，应在第三层流转后执行
    tw.AddTask(120*time.Second, "task3", func() {
        fmt.Println("Task 3 (120s) executed")
    })

    // 等待足够的时间以确保所有任务都能执行
    time.Sleep(130 * time.Second)
}

func TestRemoveTaskHierarchical(t *testing.T) {
    tw, err := New(1*time.Second, 10, 3)
    if err != nil {
        t.Fatalf("failed to create time wheel: %v", err)
    }
    tw.Start()
    defer tw.Stop()

    // 添加一个长延时任务
    tw.AddTask(120*time.Second, "task_to_remove", func() {
        t.Error("task_to_remove should have been removed")
    })

    // 在任务执行前移除它
    time.Sleep(1 * time.Second)
    tw.RemoveTask("task_to_remove")

    // 等待足够的时间以确保任务不会被执行
    time.Sleep(130 * time.Second)
}
