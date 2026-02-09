package repository

import (
	"context"

	"autoSSL/ent"
	"autoSSL/ent/workflowstep"
)

type WorkflowRepository interface {
	GetByID(ctx context.Context, id int, userID int) (*ent.Workflow, error)
	GetByIDWithSteps(ctx context.Context, id int, userID int) (*ent.Workflow, error)
	GetDetail(ctx context.Context, id int, userID int) (*ent.Workflow, error)
	List(ctx context.Context, userID int, page, pageSize int, startDate, endDate, status, domain string) ([]*ent.Workflow, int, error)
	Create(ctx context.Context, certID int, domains []string, status string, userID int, extraConfig map[string]interface{}, totalSteps int, notificationConfig map[string]interface{}) (*ent.Workflow, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	UpdateProgress(ctx context.Context, id int, currentStep int) error
	UpdateStatusWithCompletion(ctx context.Context, id int, status string, completedAt any) error
	Cancel(ctx context.Context, id int, userID int) error
	ResetForRetry(ctx context.Context, id int) error
	Count(ctx context.Context, userID int, status string) (int, error)
}

type WorkflowStepRepository interface {
	GetByWorkflowID(ctx context.Context, workflowID int) ([]*ent.WorkflowStep, error)
	UpdateStatus(ctx context.Context, stepID int, status workflowstep.Status, startedAt, completedAt any) error
	UpdateLog(ctx context.Context, stepID int, log string) error
	Create(ctx context.Context, workflowID, certID, stepNumber int, name string, userIDs []int) (*ent.WorkflowStep, error)
	BatchUpdateStatus(ctx context.Context, stepIDs []int, status workflowstep.Status, startedAt, completedAt any) error
	GetByWorkflowAndStepNumber(ctx context.Context, workflowID, stepNumber int) (*ent.WorkflowStep, error)
}
