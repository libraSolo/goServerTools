package persistence

import (
	"bufio"
	"fmt"
	"io"
	"leaderboard/internal/domain/model"
	"os"
	"strconv"
	"strings"
)

// AOFLogger 负责记录和回放排行榜的更新操作。
type AOFLogger struct {
	file *os.File
}

// NewAOFLogger 创建一个新的 AOFLogger。
func NewAOFLogger(filePath string) (*AOFLogger, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &AOFLogger{file: file}, nil
}

// LogUpdate 记录一次分数更新操作。
func (l *AOFLogger) LogUpdate(playerID int64, score int64) error {
	_, err := fmt.Fprintf(l.file, "update %d %d\n", playerID, score)
	return err
}

// Replay 回放 AOF 日志，重建排行榜状态。
func (l *AOFLogger) Replay(lb *model.Leaderboard) error {
	file, err := os.Open(l.file.Name())
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		parts := strings.Split(strings.TrimSpace(line), " ")
		if len(parts) != 3 || parts[0] != "update" {
			continue
		}

		playerID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}

		score, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}

		lb.UpdateScore(playerID, score)
	}
	return nil
}

// Close 关闭 AOF 日志文件。
func (l *AOFLogger) Close() error {
	return l.file.Close()
}