package gwutils // 通用工具函数：提供 panic 捕获与安全执行、字符串键辅助等
import (
    "fmt"
)

// CatchPanic 调用函数并在发生 panic 时返回错误值（使用 recover 捕获）
// 返回值 err 为 panic 时的错误信息；若无 panic，则为 nil
func CatchPanic(f func()) (err interface{}) {
    defer func() {
        err = recover() // 捕获可能的 panic
        if err != nil {
            fmt.Errorf("%s panic: %s", f, err) // 记录错误详情，便于排查
        }
    }()

    f() // 执行目标函数
    return
}

// RunPanicless 安全执行函数：若函数产生 panic，则吞掉并返回 false；否则返回 true
// 该函数适用于“尽量不中断流程”的场景，如批量回调执行
func RunPanicless(f func()) (panicless bool) {
    defer func() {
        err := recover()
        panicless = err == nil // 无 panic 时为 true
        if err != nil {
            fmt.Errorf("%s panic: %s", f, err) // 记录 panic 细节
        }
    }()

    f() // 执行目标函数
    return
}

// RepeatUntilPanicless 重复执行指定函数，直到其不再产生 panic
// 适合需要“确保成功完成”的循环场景（例如后台循环），避免单次 panic 中断服务
func RepeatUntilPanicless(f func()) {
    for !RunPanicless(f) { // 若执行出现 panic，则继续重试
    }
}

// NextLargerKey 返回一个“严格大于 key 且尽可能小”的下一字符串键
// 通过在末尾追加一个最小的非空字符 \x00，实现字典序下的“紧邻大于”效果
func NextLargerKey(key string) string {
    return key + "\x00" // 紧邻大于 key 的字符串，且小于任何其他 > key 的字符串
}
