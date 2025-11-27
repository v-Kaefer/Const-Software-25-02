package workspace

import (
	"context"
	"testing"
	"time"
)

func TestTaskService_CreateTaskValidations(t *testing.T) {
	db := newWorkspaceTestDB(t)
	projectSvc := NewProjectService(db)
	taskSvc := NewTaskService(db)
	ctx := context.Background()

	start := time.Now().UTC()
	end := start.Add(48 * time.Hour)
	project, err := projectSvc.CreateProject(ctx, ProjectInput{
		Name:        "App Mobile",
		ClientName:  "Retail Inc",
		Description: "Aplicativo mobile",
		StartDate:   start,
		EndDate:     &end,
		OwnerID:     1,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	// Due date outside project window must fail.
	due := end.Add(24 * time.Hour)
	_, err = taskSvc.CreateTask(ctx, TaskInput{
		ProjectID:   project.ID,
		Title:       "Design",
		Description: "Criar telas",
		AssigneeID:  99,
		DueDate:     &due,
	})
	if err == nil {
		t.Fatal("expected error for due date outside project window")
	}

	dueValid := start.Add(24 * time.Hour)
	task, err := taskSvc.CreateTask(ctx, TaskInput{
		ProjectID:   project.ID,
		Title:       "Design",
		Description: "Criar telas",
		AssigneeID:  99,
		DueDate:     &dueValid,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.Status != TaskTodo {
		t.Fatalf("expected todo status, got %s", task.Status)
	}
}

func TestTaskService_UpdateValidations(t *testing.T) {
	db := newWorkspaceTestDB(t)
	projectSvc := NewProjectService(db)
	taskSvc := NewTaskService(db)
	ctx := context.Background()

	start := time.Now().UTC()
	project, err := projectSvc.CreateProject(ctx, ProjectInput{
		Name:        "API",
		ClientName:  "FinTech",
		Description: "API financeira",
		StartDate:   start,
		EndDate:     nil,
		OwnerID:     5,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	task, err := taskSvc.CreateTask(ctx, TaskInput{
		ProjectID:   project.ID,
		Title:       "Modelagem",
		Description: "Modelagem dados",
		AssigneeID:  77,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	_, err = taskSvc.UpdateTask(ctx, task.ID, TaskUpdateInput{
		Title:       "Modelagem",
		Description: "Atualizado",
		Status:      TaskStatus("invalid"),
		AssigneeID:  77,
	})
	if err == nil {
		t.Fatal("expected validation error for invalid status")
	}

	updated, err := taskSvc.UpdateTask(ctx, task.ID, TaskUpdateInput{
		Title:       "Modelagem",
		Description: "Atualizado",
		Status:      TaskInProgress,
		AssigneeID:  77,
	})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if updated.Status != TaskInProgress {
		t.Fatalf("unexpected status %s", updated.Status)
	}
}

func TestTaskService_ListFilters(t *testing.T) {
	db := newWorkspaceTestDB(t)
	projectSvc := NewProjectService(db)
	taskSvc := NewTaskService(db)
	ctx := context.Background()

	project, err := projectSvc.CreateProject(ctx, ProjectInput{
		Name:        "Projeto",
		ClientName:  "Cliente",
		Description: "Desc",
		StartDate:   time.Now().UTC(),
		OwnerID:     1,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = taskSvc.CreateTask(ctx, TaskInput{
		ProjectID:   project.ID,
		Title:       "A",
		Description: "",
		AssigneeID:  1,
	})
	if err != nil {
		t.Fatalf("create task A: %v", err)
	}
	taskB, err := taskSvc.CreateTask(ctx, TaskInput{
		ProjectID:   project.ID,
		Title:       "B",
		Description: "",
		AssigneeID:  2,
	})
	if err != nil {
		t.Fatalf("create task B: %v", err)
	}
	if _, err := taskSvc.UpdateTask(ctx, taskB.ID, TaskUpdateInput{
		Title:       "B",
		Description: "",
		Status:      TaskInProgress,
		AssigneeID:  2,
	}); err != nil {
		t.Fatalf("update task B: %v", err)
	}

	filter := TaskFilter{
		ProjectID:  project.ID,
		Status:     []TaskStatus{TaskInProgress},
		AssigneeID: ptrUint(2),
		Page:       1,
		PageSize:   10,
	}

	page, err := taskSvc.ListTasks(ctx, filter)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("expected 1 task, got total=%d len=%d", page.Total, len(page.Items))
	}
	if page.Items[0].Status != TaskInProgress {
		t.Fatalf("unexpected status %s", page.Items[0].Status)
	}
}
