package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// TasksPage wraps a paginated list of tasks.
type TasksPage struct {
	Items []Task
	Total int64
}

// TaskFilter holds list parameters.
type TaskFilter struct {
	ProjectID  uint
	Status     []TaskStatus
	AssigneeID *uint
	Page       int
	PageSize   int
}

// TaskInput is used for task creation.
type TaskInput struct {
	ProjectID   uint
	Title       string
	Description string
	AssigneeID  uint
	DueDate     *time.Time
}

// TaskUpdateInput is used for task updates.
type TaskUpdateInput struct {
	Title       string
	Description string
	Status      TaskStatus
	AssigneeID  uint
	DueDate     *time.Time
}

// TaskService encapsulates use cases for tasks.
type TaskService struct {
	db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{db: db}
}

func (s *TaskService) CreateTask(ctx context.Context, in TaskInput) (*Task, error) {
	if err := validateTaskInput(in); err != nil {
		return nil, err
	}

	project, err := s.getProject(ctx, in.ProjectID)
	if err != nil {
		return nil, err
	}
	if project.Status == ProjectCanceled || project.Status == ProjectCompleted {
		return nil, errors.New("cannot add tasks to closed project")
	}
	if err := validateDueDate(in.DueDate, project); err != nil {
		return nil, err
	}

	task := &Task{
		ProjectID:   in.ProjectID,
		Title:       in.Title,
		Description: in.Description,
		AssigneeID:  in.AssigneeID,
		DueDate:     in.DueDate,
		Status:      TaskTodo,
	}
	if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id uint, in TaskUpdateInput) (*Task, error) {
	if err := validateTaskUpdateInput(in); err != nil {
		return nil, err
	}

	var task Task
	if err := s.db.WithContext(ctx).Preload("Project").First(&task, id).Error; err != nil {
		return nil, err
	}

	if task.Project.Status == ProjectCanceled || task.Project.Status == ProjectCompleted {
		return nil, errors.New("cannot modify tasks in closed project")
	}

	if err := validateDueDate(in.DueDate, &task.Project); err != nil {
		return nil, err
	}

	task.Title = in.Title
	task.Description = in.Description
	task.Status = in.Status
	task.AssigneeID = in.AssigneeID
	task.DueDate = in.DueDate

	if err := s.db.WithContext(ctx).Save(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TaskService) GetTask(ctx context.Context, id uint) (*Task, error) {
	var task Task
	if err := s.db.WithContext(ctx).Preload("Project").First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TaskService) ListTasks(ctx context.Context, filter TaskFilter) (TasksPage, error) {
	filter = sanitizeTaskFilter(filter)
	tx := s.db.WithContext(ctx).Model(&Task{}).Preload("Project")

	if filter.ProjectID != 0 {
		tx = tx.Where("project_id = ?", filter.ProjectID)
	}
	if len(filter.Status) > 0 {
		tx = tx.Where("status IN ?", filter.Status)
	}
	if filter.AssigneeID != nil {
		tx = tx.Where("assignee_id = ?", *filter.AssigneeID)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return TasksPage{}, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	var tasks []Task
	if err := tx.Order("due_date NULLS LAST, created_at DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return TasksPage{}, err
	}

	return TasksPage{Items: tasks, Total: total}, nil
}

func (s *TaskService) getProject(ctx context.Context, id uint) (*Project, error) {
	var project Project
	if err := s.db.WithContext(ctx).First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func validateTaskInput(in TaskInput) error {
	if in.ProjectID == 0 {
		return errors.New("project is required")
	}
	if in.AssigneeID == 0 {
		return errors.New("assignee is required")
	}
	if strings.TrimSpace(in.Title) == "" {
		return errors.New("title is required")
	}
	return nil
}

func validateTaskUpdateInput(in TaskUpdateInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return errors.New("title is required")
	}
	if in.AssigneeID == 0 {
		return errors.New("assignee is required")
	}
	switch in.Status {
	case TaskTodo, TaskInProgress, TaskBlocked, TaskDone:
	default:
		return fmt.Errorf("invalid task status %q", in.Status)
	}
	return nil
}

func validateDueDate(due *time.Time, project *Project) error {
	if due == nil {
		return nil
	}
	if due.Before(project.StartDate) {
		return errors.New("due date cannot be before project start")
	}
	if project.EndDate != nil && due.After(*project.EndDate) {
		return errors.New("due date cannot be after project end")
	}
	return nil
}

func sanitizeTaskFilter(filter TaskFilter) TaskFilter {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = defaultPageSize
	}
	if filter.PageSize > maxPageSize {
		filter.PageSize = maxPageSize
	}
	return filter
}
