// PicoClaw - Task distribution for team agents
// License: MIT

package teams

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Task represents a unit of work assigned to an agent
type Task struct {
	ID          string            `json:"id"`
	TeamID      string            `json:"team_id"`
	AgentID     string            `json:"agent_id"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	Priority    int               `json:"priority"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Payload     map[string]any    `json:"payload"`
	Result      map[string]any    `json:"result,omitempty"`
	Error       string            `json:"error,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	AssignedAt  *time.Time        `json:"assigned_at,omitempty"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	CreatedBy   string            `json:"created_by,omitempty"`
}

// Task status constants
const (
	TaskStatusPending    = "pending"
	TaskStatusAssigned   = "assigned"
	TaskStatusRunning    = "running"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
	TaskStatusCancelled  = "cancelled"
)

// CreateTaskRequest is used to create a new task
type CreateTaskRequest struct {
	AgentID     string         `json:"agent_id,omitempty"` // If empty, auto-assigned
	Type        string         `json:"type"`
	Priority    int            `json:"priority,omitempty"` // 1-10, higher = more urgent
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
	CreatedBy   string         `json:"created_by,omitempty"`
}

// AssignTaskRequest assigns a pending task to an agent
type AssignTaskRequest struct {
	TaskID  string `json:"task_id"`
	AgentID string `json:"agent_id"`
}

// TaskResult is submitted by an agent when completing a task
type TaskResult struct {
	TaskID    string         `json:"task_id"`
	Status    string         `json:"status"` // completed or failed
	Result    map[string]any `json:"result,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// TaskQueue manages tasks for all teams
type TaskQueue struct {
	tasks map[string]*Task // task_id -> Task
	mu    sync.RWMutex
}

// NewTaskQueue creates a new task queue
func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks: make(map[string]*Task),
	}
}

// CreateTask creates a new task
func (tq *TaskQueue) CreateTask(teamID string, req CreateTaskRequest) (*Task, error) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task := &Task{
		ID:          generateTaskID(),
		TeamID:      teamID,
		AgentID:     req.AgentID,
		Type:        req.Type,
		Status:      TaskStatusPending,
		Priority:    req.Priority,
		Title:       req.Title,
		Description: req.Description,
		Payload:     req.Payload,
		CreatedAt:   time.Now().UTC(),
		CreatedBy:   req.CreatedBy,
	}

	if req.Priority == 0 {
		task.Priority = 5 // Default priority
	}

	if req.AgentID != "" {
		now := time.Now().UTC()
		task.AgentID = req.AgentID
		task.Status = TaskStatusAssigned
		task.AssignedAt = &now
	}

	tq.tasks[task.ID] = task
	return task, nil
}

// GetTask retrieves a task by ID
func (tq *TaskQueue) GetTask(taskID string) (*Task, error) {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	return task, nil
}

// GetTasksForAgent returns all tasks assigned to an agent
func (tq *TaskQueue) GetTasksForAgent(agentID string, status string) []*Task {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	var tasks []*Task
	for _, task := range tq.tasks {
		if task.AgentID == agentID {
			if status == "" || task.Status == status {
				tasks = append(tasks, task)
			}
		}
	}
	return tasks
}

// GetPendingTasksForTeam returns pending tasks for auto-assignment
func (tq *TaskQueue) GetPendingTasksForTeam(teamID string) []*Task {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	var tasks []*Task
	for _, task := range tq.tasks {
		if task.TeamID == teamID && task.Status == TaskStatusPending {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// AssignTask assigns a pending task to an agent
func (tq *TaskQueue) AssignTask(taskID, agentID string) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if task.Status != TaskStatusPending {
		return fmt.Errorf("task is not pending: %s", task.Status)
	}

	now := time.Now().UTC()
	task.AgentID = agentID
	task.Status = TaskStatusAssigned
	task.AssignedAt = &now
	return nil
}

// StartTask marks a task as running
func (tq *TaskQueue) StartTask(taskID string) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	now := time.Now().UTC()
	task.Status = TaskStatusRunning
	task.StartedAt = &now
	return nil
}

// CompleteTask marks a task as completed with results
func (tq *TaskQueue) CompleteTask(taskID string, result map[string]any) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	now := time.Now().UTC()
	task.Status = TaskStatusCompleted
	task.Result = result
	task.CompletedAt = &now
	return nil
}

// FailTask marks a task as failed
func (tq *TaskQueue) FailTask(taskID string, err string) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	now := time.Now().UTC()
	task.Status = TaskStatusFailed
	task.Error = err
	task.CompletedAt = &now
	return nil
}

// CancelTask cancels a pending or assigned task
func (tq *TaskQueue) CancelTask(taskID string) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if task.Status == TaskStatusRunning {
		return fmt.Errorf("cannot cancel running task")
	}

	if task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed {
		return fmt.Errorf("cannot cancel completed task")
	}

	task.Status = TaskStatusCancelled
	return nil
}

// ListTasksForTeam returns all tasks for a team
func (tq *TaskQueue) ListTasksForTeam(teamID string) []*Task {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	var tasks []*Task
	for _, task := range tq.tasks {
		if task.TeamID == teamID {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func generateTaskID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "task_" + hex.EncodeToString(b)
}
