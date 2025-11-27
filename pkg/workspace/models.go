package workspace

import "time"

// ProjectStatus expresses lifecycle of a project.
type ProjectStatus string

const (
	ProjectPlanning  ProjectStatus = "planning"
	ProjectActive    ProjectStatus = "active"
	ProjectCompleted ProjectStatus = "completed"
	ProjectCanceled  ProjectStatus = "canceled"
)

// TaskStatus expresses lifecycle of a task.
type TaskStatus string

const (
	TaskTodo       TaskStatus = "todo"
	TaskInProgress TaskStatus = "in_progress"
	TaskBlocked    TaskStatus = "blocked"
	TaskDone       TaskStatus = "done"
)

// Project is the root entity of the delivery domain.
type Project struct {
	ID          uint          `gorm:"primaryKey"`
	Name        string        `gorm:"size:120;not null"`
	ClientName  string        `gorm:"size:120;not null"`
	Description string        `gorm:"size:500"`
	Status      ProjectStatus `gorm:"size:20;not null;default:planning"`
	OwnerID     uint          `gorm:"not null"`
	StartDate   time.Time     `gorm:"not null"`
	EndDate     *time.Time    ``
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Task represents work units inside a project.
type Task struct {
	ID          uint       `gorm:"primaryKey"`
	ProjectID   uint       `gorm:"not null"`
	Title       string     `gorm:"size:150;not null"`
	Description string     `gorm:"size:500"`
	Status      TaskStatus `gorm:"size:20;not null;default:todo"`
	AssigneeID  uint       `gorm:"not null"`
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Project     Project `gorm:"foreignKey:ProjectID"`
}

// TimeEntry tracks time spent on tasks.
type TimeEntry struct {
	ID         uint      `gorm:"primaryKey"`
	TaskID     uint      `gorm:"not null"`
	UserID     uint      `gorm:"not null"`
	EntryDate  time.Time `gorm:"not null"`
	Hours      float64   `gorm:"type:numeric(5,2);not null"`
	Notes      string    `gorm:"size:255"`
	ApprovedAt *time.Time
	ApprovedBy *uint
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
