package repository

import (
	"context"

	"autoSSL/ent"
	"autoSSL/ent/notificationconfig"
)

type NotificationConfigRepository interface {
	GetByUserID(ctx context.Context, userID int, isActive bool) ([]*ent.NotificationConfig, error)
	GetByUserIDAndType(ctx context.Context, userID int, configType notificationconfig.Type) (*ent.NotificationConfig, error)
	Create(ctx context.Context, userID int, configType string, config map[string]interface{}, isActive bool) (*ent.NotificationConfig, error)
	Update(ctx context.Context, id int, userID int, configType string, config map[string]interface{}, isActive bool) (*ent.NotificationConfig, error)
	Delete(ctx context.Context, id int, userID int) error
	CountByUserID(ctx context.Context, userID int) (int, error)
}

type NotificationTemplateRepository interface {
	GetByID(ctx context.Context, id int) (*ent.NotificationTemplate, error)
	GetByType(ctx context.Context, templateType string, isSystem bool) ([]*ent.NotificationTemplate, error)
	GetSystemTemplates(ctx context.Context) ([]*ent.NotificationTemplate, error)
	Create(ctx context.Context, templateType, name, subject, body string, isSystem bool) (*ent.NotificationTemplate, error)
	Update(ctx context.Context, id int, templateType, name, subject, body string) (*ent.NotificationTemplate, error)
	Delete(ctx context.Context, id int) error
	CountByType(ctx context.Context, templateType string) (int, error)
}
