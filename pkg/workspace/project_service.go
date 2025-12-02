package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// pagination settings shared by list endpoints.
const (
	defaultPageSize = 10
	maxPageSize     = 50
)

// ProjectsPage is returned by list queries.
type ProjectsPage struct {
	Items []Project
	Total int64
}

// ProjectFilter defines query parameters for projects.
type ProjectFilter struct {
	Status    []ProjectStatus
	Client    string
	OwnerID   *uint
	Page      int
	PageSize  int
	FromDate  *time.Time
	UntilDate *time.Time
}

// ProjectInput represents creation payload.
type ProjectInput struct {
	Name        string
	ClientName  string
	Description string
	StartDate   time.Time
	EndDate     *time.Time
	OwnerID     uint
}

// ProjectUpdateInput represents update payload.
type ProjectUpdateInput struct {
	Name        string
	ClientName  string
	Description string
	Status      ProjectStatus
	StartDate   time.Time
	EndDate     *time.Time
}

// ProjectService orchestrates use-cases for projects.
type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

func (s *ProjectService) CreateProject(ctx context.Context, in ProjectInput) (*Project, error) {
	if err := validateProjectInput(in); err != nil {
		return nil, err
	}

	project := &Project{
		Name:        in.Name,
		ClientName:  in.ClientName,
		Description: in.Description,
		Status:      ProjectPlanning,
		OwnerID:     in.OwnerID,
		StartDate:   in.StartDate,
		EndDate:     in.EndDate,
	}
	if err := s.db.WithContext(ctx).Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, id uint, in ProjectUpdateInput) (*Project, error) {
	if err := validateProjectUpdate(in); err != nil {
		return nil, err
	}

	var project Project
	if err := s.db.WithContext(ctx).First(&project, id).Error; err != nil {
		return nil, err
	}

	if err := validateProjectStatusTransition(project.Status, in.Status); err != nil {
		return nil, err
	}

	project.Name = in.Name
	project.ClientName = in.ClientName
	project.Description = in.Description
	project.StartDate = in.StartDate
	project.EndDate = in.EndDate
	project.Status = in.Status

	if err := s.db.WithContext(ctx).Save(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var taskIDs []uint
		if err := tx.Model(&Task{}).
			Where("project_id = ?", id).
			Pluck("id", &taskIDs).Error; err != nil {
			return err
		}

		if len(taskIDs) > 0 {
			if err := tx.Where("task_id IN ?", taskIDs).Delete(&TimeEntry{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("project_id = ?", id).Delete(&Task{}).Error; err != nil {
			return err
		}

		return tx.Delete(&Project{}, id).Error
	})
}

func (s *ProjectService) GetProject(ctx context.Context, id uint) (*Project, error) {
	var project Project
	if err := s.db.WithContext(ctx).First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *ProjectService) ListProjects(ctx context.Context, filter ProjectFilter) (ProjectsPage, error) {
	filter = sanitizeProjectFilter(filter)
	var (
		items []Project
		total int64
	)

	tx := s.db.WithContext(ctx).Model(&Project{})

	if len(filter.Status) > 0 {
		tx = tx.Where("status IN ?", filter.Status)
	}
	if filter.Client != "" {
		tx = tx.Where("LOWER(client_name) LIKE ?", "%"+strings.ToLower(filter.Client)+"%")
	}
	if filter.OwnerID != nil {
		tx = tx.Where("owner_id = ?", *filter.OwnerID)
	}
	if filter.FromDate != nil {
		tx = tx.Where("start_date >= ?", filter.FromDate)
	}
	if filter.UntilDate != nil {
		tx = tx.Where("start_date <= ?", filter.UntilDate)
	}

	if err := tx.Count(&total).Error; err != nil {
		return ProjectsPage{}, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := tx.Order("created_at DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Find(&items).Error; err != nil {
		return ProjectsPage{}, err
	}

	return ProjectsPage{Items: items, Total: total}, nil
}

func validateProjectInput(in ProjectInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("project name is required")
	}
	if strings.TrimSpace(in.ClientName) == "" {
		return errors.New("client name is required")
	}
	if in.OwnerID == 0 {
		return errors.New("owner is required")
	}
	if in.StartDate.IsZero() {
		return errors.New("start date is required")
	}
	if in.EndDate != nil && in.EndDate.Before(in.StartDate) {
		return errors.New("end date cannot be before start date")
	}
	return nil
}

func validateProjectUpdate(in ProjectUpdateInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("project name is required")
	}
	if strings.TrimSpace(in.ClientName) == "" {
		return errors.New("client name is required")
	}
	if in.StartDate.IsZero() {
		return errors.New("start date is required")
	}
	if in.EndDate != nil && in.EndDate.Before(in.StartDate) {
		return errors.New("end date cannot be before start date")
	}
	switch in.Status {
	case ProjectPlanning, ProjectActive, ProjectCompleted, ProjectCanceled:
	default:
		return fmt.Errorf("invalid project status %q", in.Status)
	}
	return nil
}

func validateProjectStatusTransition(current, next ProjectStatus) error {
	if current == ProjectCompleted || current == ProjectCanceled {
		if next != current {
			return fmt.Errorf("cannot change project in terminal status (%s)", current)
		}
	}
	return nil
}

func sanitizeProjectFilter(filter ProjectFilter) ProjectFilter {
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
