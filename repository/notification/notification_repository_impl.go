package repository

import (
	"context"

	"autoSSL/ent"
	"autoSSL/ent/notificationconfig"
	"autoSSL/ent/notificationtemplate"
	"autoSSL/ent/predicate"
	"autoSSL/ent/user"

	"entgo.io/ent/dialect/sql"
)

type notificationConfigRepositoryImpl struct {
	client *ent.Client
}

type notificationTemplateRepositoryImpl struct {
	client *ent.Client
}

func NewNotificationConfigRepository(client *ent.Client) NotificationConfigRepository {
	return &notificationConfigRepositoryImpl{client: client}
}

func NewNotificationTemplateRepository(client *ent.Client) NotificationTemplateRepository {
	return &notificationTemplateRepositoryImpl{client: client}
}

func (r *notificationConfigRepositoryImpl) GetByUserID(ctx context.Context, userID int, isActive bool) ([]*ent.NotificationConfig, error) {
	query := r.client.NotificationConfig.Query().
		Where(notificationconfig.HasUserWith(user.ID(userID)))

	if isActive {
		query = query.Where(notificationconfig.IsActive(true))
	}

	return query.All(ctx)
}

func (r *notificationConfigRepositoryImpl) GetByUserIDAndType(ctx context.Context, userID int, configType notificationconfig.Type) (*ent.NotificationConfig, error) {
	return r.client.NotificationConfig.Query().
		Where(
			notificationconfig.HasUserWith(user.ID(userID)),
			predicate.NotificationConfig(sql.FieldEQ(notificationconfig.FieldType, configType)),
		).
		Only(ctx)
}

func (r *notificationConfigRepositoryImpl) Create(ctx context.Context, userID int, configType string, config map[string]interface{}, isActive bool) (*ent.NotificationConfig, error) {
	return r.client.NotificationConfig.Create().
		SetType(notificationconfig.Type(configType)).
		SetConfig(config).
		SetIsActive(isActive).
		SetUserID(userID).
		Save(ctx)
}

func (r *notificationConfigRepositoryImpl) Update(ctx context.Context, id int, userID int, configType string, config map[string]interface{}, isActive bool) (*ent.NotificationConfig, error) {
	return r.client.NotificationConfig.UpdateOneID(id).
		Where(notificationconfig.HasUserWith(user.ID(userID))).
		SetType(notificationconfig.Type(configType)).
		SetConfig(config).
		SetIsActive(isActive).
		Save(ctx)
}

func (r *notificationConfigRepositoryImpl) Delete(ctx context.Context, id int, userID int) error {
	return r.client.NotificationConfig.DeleteOneID(id).
		Where(notificationconfig.HasUserWith(user.ID(userID))).
		Exec(ctx)
}

func (r *notificationConfigRepositoryImpl) CountByUserID(ctx context.Context, userID int) (int, error) {
	return r.client.NotificationConfig.Query().
		Where(notificationconfig.HasUserWith(user.ID(userID))).
		Count(ctx)
}

func (r *notificationTemplateRepositoryImpl) GetByID(ctx context.Context, id int) (*ent.NotificationTemplate, error) {
	return r.client.NotificationTemplate.Get(ctx, id)
}

func (r *notificationTemplateRepositoryImpl) GetByType(ctx context.Context, templateType string, isSystem bool) ([]*ent.NotificationTemplate, error) {
	query := r.client.NotificationTemplate.Query().
		Where(notificationtemplate.Type(templateType))

	if isSystem {
		query = query.Where(notificationtemplate.IsSystemDefault(true))
	}

	return query.All(ctx)
}

func (r *notificationTemplateRepositoryImpl) GetSystemTemplates(ctx context.Context) ([]*ent.NotificationTemplate, error) {
	return r.client.NotificationTemplate.Query().
		Where(notificationtemplate.IsSystemDefault(true)).
		All(ctx)
}

func (r *notificationTemplateRepositoryImpl) Create(ctx context.Context, templateType, name, subject, body string, isSystem bool) (*ent.NotificationTemplate, error) {
	return r.client.NotificationTemplate.Create().
		SetType(templateType).
		SetName(name).
		SetSubjectTemplate(subject).
		SetContentTemplate(body).
		SetIsSystemDefault(isSystem).
		SetIsActive(true).
		Save(ctx)
}

func (r *notificationTemplateRepositoryImpl) Update(ctx context.Context, id int, templateType, name, subject, body string) (*ent.NotificationTemplate, error) {
	return r.client.NotificationTemplate.UpdateOneID(id).
		SetType(templateType).
		SetName(name).
		SetSubjectTemplate(subject).
		SetContentTemplate(body).
		Save(ctx)
}

func (r *notificationTemplateRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.client.NotificationTemplate.DeleteOneID(id).Exec(ctx)
}

func (r *notificationTemplateRepositoryImpl) CountByType(ctx context.Context, templateType string) (int, error) {
	return r.client.NotificationTemplate.Query().
		Where(notificationtemplate.Type(templateType)).
		Count(ctx)
}
