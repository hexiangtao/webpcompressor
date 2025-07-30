package infrastructure

import (
	"fmt"
	"sync"
	"time"

	"webpcompressor/internal/domain"
	"webpcompressor/pkg/logger"

	"github.com/google/uuid"
)

// MemoryTaskManager 基于内存的任务管理器
type MemoryTaskManager struct {
	tasks  map[string]*domain.TaskInfo
	mu     sync.RWMutex
	logger logger.Logger
}

// NewMemoryTaskManager 创建内存任务管理器
func NewMemoryTaskManager(logger logger.Logger) *MemoryTaskManager {
	return &MemoryTaskManager{
		tasks:  make(map[string]*domain.TaskInfo),
		logger: logger,
	}
}

// CreateTask 创建任务
func (m *MemoryTaskManager) CreateTask(inputFile, outputFile string, config *domain.CompressionConfig) (*domain.TaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	taskID := uuid.New().String()
	task := &domain.TaskInfo{
		ID:         taskID,
		Status:     domain.TaskStatusPending,
		Progress:   0.0,
		Message:    "任务已创建",
		InputFile:  inputFile,
		OutputFile: outputFile,
		Config:     config,
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	m.tasks[taskID] = task

	m.logger.Info("创建新任务",
		"task_id", taskID,
		"input", inputFile,
		"output", outputFile,
		"quality", config.Quality,
	)

	return task, nil
}

// GetTask 获取任务信息
func (m *MemoryTaskManager) GetTask(taskID string) (*domain.TaskInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}

	// 返回副本以避免并发修改
	taskCopy := *task
	return &taskCopy, nil
}

// UpdateTask 更新任务
func (m *MemoryTaskManager) UpdateTask(task *domain.TaskInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[task.ID]; !exists {
		return fmt.Errorf("任务不存在: %s", task.ID)
	}

	m.tasks[task.ID] = task
	return nil
}

// ListTasks 列出任务
func (m *MemoryTaskManager) ListTasks(limit, offset int) ([]*domain.TaskInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tasks []*domain.TaskInfo
	count := 0
	skipped := 0

	// 按创建时间倒序
	for _, task := range m.tasks {
		if skipped < offset {
			skipped++
			continue
		}
		if count >= limit {
			break
		}

		// 返回副本
		taskCopy := *task
		tasks = append(tasks, &taskCopy)
		count++
	}

	return tasks, nil
}

// DeleteTask 删除任务
func (m *MemoryTaskManager) DeleteTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	delete(m.tasks, taskID)

	m.logger.Info("删除任务", "task_id", taskID)
	return nil
}

// CleanupOldTasks 清理旧任务
func (m *MemoryTaskManager) CleanupOldTasks(olderThan time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	var deletedCount int

	for taskID, task := range m.tasks {
		// 清理已完成且超过指定时间的任务
		if (task.Status == domain.TaskStatusCompleted ||
			task.Status == domain.TaskStatusFailed ||
			task.Status == domain.TaskStatusCancelled) &&
			task.CreatedAt.Before(cutoff) {
			delete(m.tasks, taskID)
			deletedCount++
		}
	}

	if deletedCount > 0 {
		m.logger.Info("清理旧任务", "deleted_count", deletedCount)
	}

	return nil
}

// GetStats 获取统计信息
func (m *MemoryTaskManager) GetStats() *domain.WebPStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &domain.WebPStats{}
	var totalQuality, totalRatio float64
	var qualityCount, ratioCount int

	for _, task := range m.tasks {
		stats.TotalTasks++

		switch task.Status {
		case domain.TaskStatusCompleted:
			stats.CompletedTasks++
			if task.Result != nil {
				stats.TotalSavedBytes += task.Result.OriginalSize - task.Result.CompressedSize
				totalRatio += task.Result.CompressionRatio
				ratioCount++
			}
		case domain.TaskStatusFailed:
			stats.FailedTasks++
		case domain.TaskStatusProcessing:
			stats.ProcessingTasks++
		case domain.TaskStatusPending:
			stats.PendingTasks++
		}

		if task.Config != nil {
			totalQuality += float64(task.Config.Quality)
			qualityCount++
		}
	}

	if qualityCount > 0 {
		stats.AverageQuality = totalQuality / float64(qualityCount)
	}
	if ratioCount > 0 {
		stats.AverageRatio = totalRatio / float64(ratioCount)
	}

	return stats
}

// ProgressReporter WebSocket进度报告器
type ProgressReporter struct {
	subscribers map[string][]chan domain.TaskInfo
	mu          sync.RWMutex
	logger      logger.Logger
}

// NewProgressReporter 创建进度报告器
func NewProgressReporter(logger logger.Logger) *ProgressReporter {
	return &ProgressReporter{
		subscribers: make(map[string][]chan domain.TaskInfo),
		logger:      logger,
	}
}

// ReportProgress 报告进度
func (p *ProgressReporter) ReportProgress(taskID string, progress float64, message string) {
	p.mu.RLock()
	channels, exists := p.subscribers[taskID]
	p.mu.RUnlock()

	if !exists {
		return
	}

	update := domain.TaskInfo{
		ID:       taskID,
		Progress: progress,
		Message:  message,
	}

	// 向所有订阅者发送更新
	for _, ch := range channels {
		select {
		case ch <- update:
		default:
			// 如果通道已满，跳过
			p.logger.Warn("进度更新通道已满", "task_id", taskID)
		}
	}
}

// Subscribe 订阅进度更新
func (p *ProgressReporter) Subscribe(taskID string) <-chan domain.TaskInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan domain.TaskInfo, 10)
	p.subscribers[taskID] = append(p.subscribers[taskID], ch)

	p.logger.Debug("订阅任务进度", "task_id", taskID)
	return ch
}

// Unsubscribe 取消订阅
func (p *ProgressReporter) Unsubscribe(taskID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.subscribers, taskID)
	p.logger.Debug("取消订阅任务进度", "task_id", taskID)
}

// CleanupSubscribers 清理订阅者
func (p *ProgressReporter) CleanupSubscribers() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for taskID, channels := range p.subscribers {
		for _, ch := range channels {
			close(ch)
		}
		delete(p.subscribers, taskID)
	}
}
