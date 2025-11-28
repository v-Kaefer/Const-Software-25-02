package workspace

import (
	"context"
	"testing"
	"time"
)

func TestTimeEntryService_LogAndApprove(t *testing.T) {
	db := newWorkspaceTestDB(t)
	projectSvc := NewProjectService(db)
	taskSvc := NewTaskService(db)
	timeSvc := NewTimeEntryService(db)
	ctx := context.Background()

	start := time.Now().Add(-48 * time.Hour).UTC()
	end := time.Now().Add(48 * time.Hour).UTC()
	project, err := projectSvc.CreateProject(ctx, ProjectInput{
		Name:        "Integração",
		ClientName:  "Banco X",
		Description: "Integração PIX",
		StartDate:   start,
		EndDate:     &end,
		OwnerID:     3,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	task, err := taskSvc.CreateTask(ctx, TaskInput{
		ProjectID:   project.ID,
		Title:       "Endpoints",
		Description: "Implementar endpoints",
		AssigneeID:  99,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	// Invalid hours
	_, err = timeSvc.LogTime(ctx, TimeEntryInput{
		TaskID:    task.ID,
		UserID:    99,
		EntryDate: start,
		Hours:     30,
	})
	if err == nil {
		t.Fatal("expected error for invalid hours")
	}

	entry, err := timeSvc.LogTime(ctx, TimeEntryInput{
		TaskID:    task.ID,
		UserID:    99,
		EntryDate: time.Now().UTC(),
		Hours:     4.5,
		Notes:     "Implementação inicial",
	})
	if err != nil {
		t.Fatalf("log time: %v", err)
	}

	if entry.ApprovedAt != nil {
		t.Fatal("entry should not be approved yet")
	}

	updated, err := timeSvc.ApproveEntry(ctx, entry.ID, 3)
	if err != nil {
		t.Fatalf("approve entry: %v", err)
	}
	if updated.ApprovedAt == nil || updated.ApprovedBy == nil {
		t.Fatal("entry should contain approval fields")
	}
}

func TestTimeEntryService_ApproveTwice(t *testing.T) {
	db := newWorkspaceTestDB(t)
	projectSvc := NewProjectService(db)
	taskSvc := NewTaskService(db)
	timeSvc := NewTimeEntryService(db)
	ctx := context.Background()

	project, err := projectSvc.CreateProject(ctx, ProjectInput{
		Name:        "Projeto",
		ClientName:  "Cliente",
		Description: "Desc",
		StartDate:   time.Now().UTC().Add(-time.Hour),
		EndDate:     ptrTime(time.Now().UTC().Add(24 * time.Hour)),
		OwnerID:     1,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	task, err := taskSvc.CreateTask(ctx, TaskInput{
		ProjectID:   project.ID,
		Title:       "Task",
		Description: "",
		AssigneeID:  1,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	entry, err := timeSvc.LogTime(ctx, TimeEntryInput{
		TaskID:    task.ID,
		UserID:    1,
		EntryDate: time.Now().UTC(),
		Hours:     1,
	})
	if err != nil {
		t.Fatalf("log time: %v", err)
	}

	first, err := timeSvc.ApproveEntry(ctx, entry.ID, 2)
	if err != nil {
		t.Fatalf("approve entry first: %v", err)
	}
	second, err := timeSvc.ApproveEntry(ctx, entry.ID, 3)
	if err != nil {
		t.Fatalf("approve entry second: %v", err)
	}
	if second.ApprovedBy == nil || *second.ApprovedBy != 2 {
		t.Fatalf("approval should not change approver, got %+v", second.ApprovedBy)
	}
	if first.ApprovedAt == nil || second.ApprovedAt == nil || !first.ApprovedAt.Equal(*second.ApprovedAt) {
		t.Fatalf("expected same approval timestamp")
	}
}

func TestTimeEntryService_LogTimeDateValidation(t *testing.T) {
	db := newWorkspaceTestDB(t)
	projectSvc := NewProjectService(db)
	taskSvc := NewTaskService(db)
	timeSvc := NewTimeEntryService(db)
	ctx := context.Background()

	start := time.Now().UTC()
	project, err := projectSvc.CreateProject(ctx, ProjectInput{
		Name:        "Projeto",
		ClientName:  "Cliente",
		Description: "Desc",
		StartDate:   start,
		EndDate:     ptrTime(start.Add(4 * time.Hour)),
		OwnerID:     1,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	task, err := taskSvc.CreateTask(ctx, TaskInput{
		ProjectID:   project.ID,
		Title:       "Task",
		Description: "",
		AssigneeID:  1,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	_, err = timeSvc.LogTime(ctx, TimeEntryInput{
		TaskID:    task.ID,
		UserID:    1,
		EntryDate: start.Add(10 * time.Hour),
		Hours:     1,
	})
	if err == nil {
		t.Fatal("expected error when entry date after project end")
	}
}
