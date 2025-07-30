package web

import (
	"net/http"
	"time"

	"webpcompressor/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket升级器配置
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源（生产环境应该更严格）
		return true
	},
}

// handleProgressWebSocket 处理WebSocket进度推送
func (s *Server) handleProgressWebSocket(c *gin.Context) {
	taskID := c.Param("taskId")

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("WebSocket升级失败", "error", err)
		return
	}
	defer conn.Close()

	s.logger.Info("WebSocket连接建立", "task_id", taskID, "remote_addr", conn.RemoteAddr())

	// 订阅任务进度
	progressChan := s.progressReporter.Subscribe(taskID)
	defer s.progressReporter.Unsubscribe(taskID)

	// 设置WebSocket连接参数
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 启动ping定时器
	ticker := time.NewTicker(30 * time.Second) // 减少ping频率
	defer ticker.Stop()

	// 发送初始任务状态
	if task, err := s.taskManager.GetTask(taskID); err == nil {
		s.sendTaskUpdate(conn, task)
	}

	// 创建退出通道
	done := make(chan struct{})

	// 启动消息读取协程
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					s.logger.Error("WebSocket意外关闭", "task_id", taskID, "error", err)
				} else {
					s.logger.Debug("WebSocket连接关闭", "task_id", taskID)
				}
				return
			}
		}
	}()

	// 处理消息循环
	for {
		select {
		case update, ok := <-progressChan:
			if !ok {
				s.logger.Debug("进度通道已关闭", "task_id", taskID)
				return
			}

			if err := s.sendProgressUpdate(conn, &update); err != nil {
				s.logger.Error("发送进度更新失败", "task_id", taskID, "error", err)
				return
			}

		case <-ticker.C:
			// 发送ping保持连接
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				s.logger.Error("发送ping失败", "task_id", taskID, "error", err)
				return
			}

		case <-done:
			s.logger.Debug("WebSocket读取协程退出", "task_id", taskID)
			return
		}
	}
}

// sendTaskUpdate 发送任务更新
func (s *Server) sendTaskUpdate(conn *websocket.Conn, task *domain.TaskInfo) error {
	message := map[string]interface{}{
		"type":      "task_update",
		"task_id":   task.ID,
		"status":    task.Status,
		"progress":  task.Progress,
		"message":   task.Message,
		"timestamp": time.Now(),
	}

	if task.Result != nil {
		message["result"] = task.Result
	}

	if task.Error != "" {
		message["error"] = task.Error
	}

	return conn.WriteJSON(message)
}

// sendProgressUpdate 发送进度更新
func (s *Server) sendProgressUpdate(conn *websocket.Conn, update *domain.TaskInfo) error {
	message := map[string]interface{}{
		"type":      "progress_update",
		"task_id":   update.ID,
		"progress":  update.Progress,
		"message":   update.Message,
		"timestamp": time.Now(),
	}

	return conn.WriteJSON(message)
}

// broadcastTaskUpdate 广播任务更新（给所有相关的WebSocket连接）
func (s *Server) broadcastTaskUpdate(task *domain.TaskInfo) {
	// 通过进度报告器发送更新
	s.progressReporter.ReportProgress(task.ID, task.Progress, task.Message)
}

// WebSocketMessage WebSocket消息结构
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	TaskID    string                 `json:"task_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// createWebSocketMessage 创建WebSocket消息
func createWebSocketMessage(msgType, taskID string, data map[string]interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      msgType,
		TaskID:    taskID,
		Data:      data,
		Timestamp: time.Now(),
	}
}
