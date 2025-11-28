package workspace

import (
	"context"
	"errors"
	"math"
	"time"

	"gorm.io/gorm"
)

// TimeEntriesPage wraps paginated time entries.
type TimeEntriesPage struct {
	Items []TimeEntry
	Total int64
}

// TimeEntryFilter holds filters for list queries.
type TimeEntryFilter struct {
	TaskID   *uint
	UserID   *uint
	Approved *bool
	Page     int
	PageSize int
}

// TimeEntryInput holds data to log time.
type TimeEntryInput struct {
	TaskID    uint
	UserID    uint
	EntryDate time.Time
	Hours     float64
	Notes     string
}

// TimeEntryUpdateInput updates entry before approval.
type TimeEntryUpdateInput struct {
	EntryDate time.Time
	Hours     float64
	Notes     string
}

// TimeEntryService orchestrates time tracking flows.
type TimeEntryService struct {
	db *gorm.DB
}

func NewTimeEntryService(db *gorm.DB) *TimeEntryService {
	return &TimeEntryService{db: db}
}

func (s *TimeEntryService) LogTime(ctx context.Context, in TimeEntryInput) (*TimeEntry, error) {
	if err := validateTimeEntryInput(in); err != nil {
		return nil, err
	}
	task, err := s.loadTask(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}
	if err := validateEntryAgainstTask(in.EntryDate, task); err != nil {
		return nil, err
	}

	entry := &TimeEntry{
		TaskID:    in.TaskID,
		UserID:    in.UserID,
		EntryDate: in.EntryDate,
		Hours:     in.Hours,
		Notes:     in.Notes,
	}
	if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *TimeEntryService) UpdateEntry(ctx context.Context, id uint, in TimeEntryUpdateInput) (*TimeEntry, error) {
	if err := validateTimeEntryUpdateInput(in); err != nil {
		return nil, err
	}

	var entry TimeEntry
	if err := s.db.WithContext(ctx).First(&entry, id).Error; err != nil {
		return nil, err
	}
	if entry.ApprovedAt != nil {
		return nil, errors.New("cannot update approved entry")
	}

	task, err := s.loadTask(ctx, entry.TaskID)
	if err != nil {
		return nil, err
	}
	if err := validateEntryAgainstTask(in.EntryDate, task); err != nil {
		return nil, err
	}

	entry.EntryDate = in.EntryDate
	entry.Hours = in.Hours
	entry.Notes = in.Notes
	if err := s.db.WithContext(ctx).Save(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *TimeEntryService) ApproveEntry(ctx context.Context, id uint, approverID uint) (*TimeEntry, error) {
	if approverID == 0 {
		return nil, errors.New("approver is required")
	}
	var entry TimeEntry
	if err := s.db.WithContext(ctx).First(&entry, id).Error; err != nil {
		return nil, err
	}
	if entry.ApprovedAt != nil {
		return &entry, nil
	}
	now := time.Now().UTC()
	entry.ApprovedAt = &now
	entry.ApprovedBy = &approverID
	if err := s.db.WithContext(ctx).Save(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *TimeEntryService) GetEntry(ctx context.Context, id uint) (*TimeEntry, error) {
	var entry TimeEntry
	if err := s.db.WithContext(ctx).First(&entry, id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *TimeEntryService) ListEntries(ctx context.Context, filter TimeEntryFilter) (TimeEntriesPage, error) {
	filter = sanitizeTimeEntryFilter(filter)

	tx := s.db.WithContext(ctx).Model(&TimeEntry{})
	if filter.TaskID != nil {
		tx = tx.Where("task_id = ?", *filter.TaskID)
	}
	if filter.UserID != nil {
		tx = tx.Where("user_id = ?", *filter.UserID)
	}
	if filter.Approved != nil {
		if *filter.Approved {
			tx = tx.Where("approved_at IS NOT NULL")
		} else {
			tx = tx.Where("approved_at IS NULL")
		}
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return TimeEntriesPage{}, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	var entries []TimeEntry
	if err := tx.Order("entry_date DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Find(&entries).Error; err != nil {
		return TimeEntriesPage{}, err
	}

	return TimeEntriesPage{Items: entries, Total: total}, nil
}

func (s *TimeEntryService) loadTask(ctx context.Context, id uint) (*Task, error) {
	var task Task
	if err := s.db.WithContext(ctx).
		Preload("Project").
		First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func validateTimeEntryInput(in TimeEntryInput) error {
	if in.TaskID == 0 {
		return errors.New("task is required")
	}
	if in.UserID == 0 {
		return errors.New("user is required")
	}
	if err := validateHourValue(in.Hours); err != nil {
		return err
	}
	if in.EntryDate.IsZero() {
		return errors.New("entry date is required")
	}
	return nil
}

func validateTimeEntryUpdateInput(in TimeEntryUpdateInput) error {
	if err := validateHourValue(in.Hours); err != nil {
		return err
	}
	if in.EntryDate.IsZero() {
		return errors.New("entry date is required")
	}
	return nil
}

func validateHourValue(hours float64) error {
	if math.IsNaN(hours) || math.IsInf(hours, 0) {
		return errors.New("hours must be a number")
	}
	if hours <= 0 {
		return errors.New("hours must be greater than zero")
	}
	if hours > 24 {
		return errors.New("hours cannot exceed 24 in a single entry")
	}
	return nil
}

func sanitizeTimeEntryFilter(filter TimeEntryFilter) TimeEntryFilter {
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

func validateEntryAgainstTask(entryDate time.Time, task *Task) error {
	if entryDate.Before(task.Project.StartDate) {
		return errors.New("entry date cannot be before project start")
	}
	if task.Project.EndDate != nil && entryDate.After(*task.Project.EndDate) {
		return errors.New("entry date cannot be after project end")
	}
	if entryDate.After(time.Now().Add(24 * time.Hour)) {
		return errors.New("entry date cannot be in the far future")
	}
	if task.Project.Status == ProjectCanceled {
		return errors.New("cannot log time for canceled project")
	}
	return nil
}
