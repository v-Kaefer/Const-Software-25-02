package workspace

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newWorkspaceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Project{}, &Task{}, &TimeEntry{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestProjectService_CreateAndList(t *testing.T) {
	db := newWorkspaceTestDB(t)
	service := NewProjectService(db)
	ctx := context.Background()

	start := time.Now().UTC()
	project, err := service.CreateProject(ctx, ProjectInput{
		Name:        "Portal",
		ClientName:  "ACME",
		Description: "Portal corporativo",
		StartDate:   start,
		EndDate:     nil,
		OwnerID:     1,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if project.Status != ProjectPlanning {
		t.Fatalf("expected planning status, got %s", project.Status)
	}

	page, err := service.ListProjects(ctx, ProjectFilter{Client: "acme"})
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if page.Total != 1 {
		t.Fatalf("expected total=1, got %d", page.Total)
	}
	if len(page.Items) != 1 || page.Items[0].ClientName != "ACME" {
		t.Fatalf("unexpected projects returned: %+v", page.Items)
	}
}

func TestProjectService_StatusTransition(t *testing.T) {
	db := newWorkspaceTestDB(t)
	service := NewProjectService(db)
	ctx := context.Background()

	start := time.Now().UTC()
	end := start.Add(24 * time.Hour)
	project, err := service.CreateProject(ctx, ProjectInput{
		Name:        "ERP",
		ClientName:  "Big Corp",
		Description: "Entrega ERP",
		StartDate:   start,
		EndDate:     &end,
		OwnerID:     10,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if _, err := service.UpdateProject(ctx, project.ID, ProjectUpdateInput{
		Name:        "ERP",
		ClientName:  "Big Corp",
		Description: "Entrega ERP",
		Status:      ProjectCompleted,
		StartDate:   start,
		EndDate:     &end,
	}); err != nil {
		t.Fatalf("complete project: %v", err)
	}

	_, err = service.UpdateProject(ctx, project.ID, ProjectUpdateInput{
		Name:        "ERP",
		ClientName:  "Big Corp",
		Description: "Entrega ERP",
		Status:      ProjectActive,
		StartDate:   start,
		EndDate:     &end,
	})
	if err == nil {
		t.Fatal("expected error when re-opening completed project")
	}
}

func TestProjectService_ListFiltersAndPagination(t *testing.T) {
	db := newWorkspaceTestDB(t)
	service := NewProjectService(db)
	ctx := context.Background()

	start := time.Now().UTC()
	for i := 0; i < 3; i++ {
		in := ProjectInput{
			Name:        fmt.Sprintf("Projeto %d", i+1),
			ClientName:  "Cliente A",
			Description: "Teste",
			StartDate:   start.Add(time.Duration(i) * time.Hour),
			OwnerID:     1,
		}
		if i == 2 {
			in.OwnerID = 2
		}
		if _, err := service.CreateProject(ctx, in); err != nil {
			t.Fatalf("create project %d: %v", i, err)
		}
	}

	// Atualiza status de um projeto
	_, err := service.UpdateProject(ctx, 1, ProjectUpdateInput{
		Name:        "Projeto 1",
		ClientName:  "Cliente A",
		Description: "Teste",
		Status:      ProjectActive,
		StartDate:   start,
	})
	if err != nil {
		t.Fatalf("update project status: %v", err)
	}

	filter := ProjectFilter{
		Status:   []ProjectStatus{ProjectActive},
		OwnerID:  ptrUint(1),
		Page:     1,
		PageSize: 2,
	}
	page, err := service.ListProjects(ctx, filter)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("expected 1 project, got total=%d len=%d", page.Total, len(page.Items))
	}
	if page.Items[0].Status != ProjectActive {
		t.Fatalf("expected active status, got %s", page.Items[0].Status)
	}
}

func TestProjectService_DeleteCascade(t *testing.T) {
	db := newWorkspaceTestDB(t)
	projectSvc := NewProjectService(db)
	taskSvc := NewTaskService(db)
	timeSvc := NewTimeEntryService(db)
	ctx := context.Background()

	start := time.Now().UTC()
	project, err := projectSvc.CreateProject(ctx, ProjectInput{
		Name:        "Projeto",
		ClientName:  "Cliente",
		StartDate:   start,
		EndDate:     ptrTime(start.Add(48 * time.Hour)),
		OwnerID:     1,
		Description: "teste",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	task, err := taskSvc.CreateTask(ctx, TaskInput{
		ProjectID:   project.ID,
		Title:       "Tarefa",
		Description: "desc",
		AssigneeID:  1,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if _, err := timeSvc.LogTime(ctx, TimeEntryInput{
		TaskID:    task.ID,
		UserID:    1,
		EntryDate: start.Add(time.Hour),
		Hours:     2,
	}); err != nil {
		t.Fatalf("log time: %v", err)
	}

	if err := projectSvc.DeleteProject(ctx, project.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}

	var taskCount int64
	if err := db.Model(&Task{}).Count(&taskCount).Error; err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if taskCount != 0 {
		t.Fatalf("expected tasks deleted, got %d", taskCount)
	}

	var entryCount int64
	if err := db.Model(&TimeEntry{}).Count(&entryCount).Error; err != nil {
		t.Fatalf("count time entries: %v", err)
	}
	if entryCount != 0 {
		t.Fatalf("expected time entries deleted, got %d", entryCount)
	}
}

func ptrUint(v uint) *uint { return &v }

func ptrTime(v time.Time) *time.Time { return &v }
