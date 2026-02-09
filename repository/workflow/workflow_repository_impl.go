package repository

import (
	"context"
	"fmt"
	"time"

	"autoSSL/ent"
	"autoSSL/ent/certificate"
	"autoSSL/ent/user"
	"autoSSL/ent/workflow"
	"autoSSL/ent/workflowstep"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

type workflowRepositoryImpl struct {
	client *ent.Client
}

type workflowStepRepositoryImpl struct {
	client *ent.Client
}

func NewWorkflowRepository(client *ent.Client) WorkflowRepository {
	return &workflowRepositoryImpl{client: client}
}

func NewWorkflowStepRepository(client *ent.Client) WorkflowStepRepository {
	return &workflowStepRepositoryImpl{client: client}
}

func (r *workflowRepositoryImpl) GetByID(ctx context.Context, id int, userID int) (*ent.Workflow, error) {
	return r.client.Workflow.Query().
		Where(workflow.ID(id), workflow.HasCertificateWith(certificate.HasUserWith(user.ID(userID)))).
		Only(ctx)
}

func (r *workflowRepositoryImpl) GetByIDWithSteps(ctx context.Context, id int, userID int) (*ent.Workflow, error) {
	return r.client.Workflow.Query().
		Where(workflow.ID(id), workflow.HasCertificateWith(certificate.HasUserWith(user.ID(userID)))).
		WithSteps(func(sq *ent.WorkflowStepQuery) {
			sq.Order(ent.Asc(workflowstep.FieldStepNumber))
		}).
		Only(ctx)
}

func (r *workflowRepositoryImpl) GetDetail(ctx context.Context, id int, userID int) (*ent.Workflow, error) {
	return r.client.Workflow.Query().
		Where(workflow.ID(id), workflow.HasCertificateWith(certificate.HasUserWith(user.ID(userID)))).
		WithCertificate(func(cq *ent.CertificateQuery) {
			cq.WithDNSProvider()
		}).
		WithSteps(func(sq *ent.WorkflowStepQuery) {
			sq.Order(ent.Asc(workflowstep.FieldStepNumber))
		}).
		Only(ctx)
}

func (r *workflowRepositoryImpl) List(ctx context.Context, userID int, page, pageSize int, startDate, endDate, status, domain string) ([]*ent.Workflow, int, error) {
	query := r.client.Workflow.Query().
		Where(workflow.HasCertificateWith(certificate.HasUserWith(user.ID(userID))))

	if startDate != "" {
		if startTime, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where(workflow.CreatedAtGTE(startTime))
		}
	}
	if endDate != "" {
		if endTime, err := time.Parse("2006-01-02", endDate); err == nil {
			endTime = endTime.Add(24 * time.Hour).Add(-time.Second)
			query = query.Where(workflow.CreatedAtLTE(endTime))
		}
	}
	if status != "" {
		query = query.Where(workflow.Status(status))
	}
	if domain != "" {
		query = query.Where(func(s *sql.Selector) {
			switch s.Dialect() {
			case dialect.MySQL:
				s.Where(sql.P(func(b *sql.Builder) {
					b.WriteString("JSON_CONTAINS(domains, ")
					b.Arg(fmt.Sprintf("\"%s\"", domain))
					b.WriteString(")")
				}))
			case dialect.Postgres:
				s.Where(sql.P(func(b *sql.Builder) {
					b.WriteString("EXISTS (SELECT 1 FROM jsonb_array_elements_text(domains) AS elem WHERE LOWER(elem) LIKE ?)")
					b.Arg("%" + domain + "%")
				}))
			default:
				s.Where(sql.P(func(b *sql.Builder) {
					b.WriteString("domains LIKE ?")
					b.Arg("%" + domain + "%")
				}))
			}
		})
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("查询工作流总数失败: %w", err)
	}

	workflows, err := query.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order(ent.Desc(workflow.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, total, fmt.Errorf("查询工作流列表失败: %w", err)
	}

	return workflows, total, nil
}

func (r *workflowRepositoryImpl) Create(ctx context.Context, certID int, domains []string, status string, userID int, extraConfig map[string]interface{}, totalSteps int, notificationConfig map[string]interface{}) (*ent.Workflow, error) {
	if extraConfig == nil {
		extraConfig = make(map[string]interface{})
	}

	workflowCreate := r.client.Workflow.Create().
		SetCertificateID(certID).
		SetDomains(domains).
		SetStatus(status).
		SetUserID(userID).
		SetCurrentStep(0).
		SetTotalSteps(totalSteps).
		SetExtraConfig(extraConfig)

	if notificationConfig != nil {
		workflowCreate.SetNotificationConfig(notificationConfig)
	}

	return workflowCreate.Save(ctx)
}

func (r *workflowRepositoryImpl) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := r.client.Workflow.UpdateOneID(id).SetStatus(status).Save(ctx)
	return err
}

func (r *workflowRepositoryImpl) UpdateProgress(ctx context.Context, id int, currentStep int) error {
	_, err := r.client.Workflow.UpdateOneID(id).
		SetCurrentStep(currentStep).
		Save(ctx)
	return err
}

func (r *workflowRepositoryImpl) UpdateStatusWithCompletion(ctx context.Context, id int, status string, completedAt any) error {
	update := r.client.Workflow.UpdateOneID(id).SetStatus(status)
	if completedAt != nil {
		update = update.SetCompletedAt(completedAt.(time.Time))
	}
	_, err := update.Save(ctx)
	return err
}

func (r *workflowRepositoryImpl) Cancel(ctx context.Context, id int, userID int) error {
	_, err := r.client.Workflow.Query().
		Where(workflow.ID(id), workflow.HasCertificateWith(certificate.HasUserWith(user.ID(userID)))).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("无权限取消工作流: %w", err)
	}

	_, err = r.client.Workflow.UpdateOneID(id).
		SetStatus(string(StatusCancelled)).
		SetCompletedAt(time.Now()).
		Save(ctx)
	return err
}

func (r *workflowRepositoryImpl) ResetForRetry(ctx context.Context, id int) error {
	_, err := r.client.Workflow.UpdateOneID(id).
		SetStatus(string(StatusProcessing)).
		SetCurrentStep(0).
		Save(ctx)
	return err
}

func (r *workflowRepositoryImpl) Count(ctx context.Context, userID int, status string) (int, error) {
	query := r.client.Workflow.Query().
		Where(workflow.HasCertificateWith(certificate.HasUserWith(user.ID(userID))))
	if status != "" {
		query = query.Where(workflow.Status(status))
	}
	return query.Count(ctx)
}

func (r *workflowStepRepositoryImpl) GetByWorkflowID(ctx context.Context, workflowID int) ([]*ent.WorkflowStep, error) {
	return r.client.WorkflowStep.Query().
		Where(workflowstep.HasWorkflowWith(workflow.ID(workflowID))).
		Order(ent.Asc(workflowstep.FieldStepNumber)).
		All(ctx)
}

func (r *workflowStepRepositoryImpl) UpdateStatus(ctx context.Context, stepID int, status workflowstep.Status, startedAt, completedAt any) error {
	update := r.client.WorkflowStep.UpdateOneID(stepID)
	if startedAt != nil {
		update = update.SetStartedAt(startedAt.(time.Time))
	}
	if completedAt != nil {
		update = update.SetCompletedAt(completedAt.(time.Time))
	}
	_, err := update.SetStatus(status).Save(ctx)
	return err
}

func (r *workflowStepRepositoryImpl) UpdateLog(ctx context.Context, stepID int, log string) error {
	_, err := r.client.WorkflowStep.UpdateOneID(stepID).SetLog(log).Save(ctx)
	return err
}

func (r *workflowStepRepositoryImpl) Create(ctx context.Context, workflowID, certID, stepNumber int, name string, userIDs []int) (*ent.WorkflowStep, error) {
	stepCreate := r.client.WorkflowStep.Create().
		SetWorkflowID(workflowID).
		SetCertificateID(certID).
		SetStepNumber(stepNumber).
		SetName(name).
		SetStatus(workflowstep.StatusPending)

	if len(userIDs) > 0 {
		stepCreate.AddUserIDs(userIDs...)
	}

	return stepCreate.Save(ctx)
}

func (r *workflowStepRepositoryImpl) BatchUpdateStatus(ctx context.Context, stepIDs []int, status workflowstep.Status, startedAt, completedAt any) error {
	if len(stepIDs) == 0 {
		return nil
	}

	update := r.client.WorkflowStep.Update().
		Where(workflowstep.IDIn(stepIDs...))

	if startedAt != nil {
		update = update.SetStartedAt(startedAt.(time.Time))
	}
	if completedAt != nil {
		update = update.SetCompletedAt(completedAt.(time.Time))
	}

	_, err := update.SetStatus(status).Save(ctx)
	return err
}

func (r *workflowStepRepositoryImpl) GetByWorkflowAndStepNumber(ctx context.Context, workflowID, stepNumber int) (*ent.WorkflowStep, error) {
	return r.client.WorkflowStep.Query().
		Where(
			workflowstep.HasWorkflowWith(workflow.ID(workflowID)),
			workflowstep.StepNumber(stepNumber),
		).
		Only(ctx)
}
