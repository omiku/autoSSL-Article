package repository

import (
	"context"

	"autoSSL/ent"
	"autoSSL/ent/certificate"
)

type CertificateRepository interface {
	GetByID(ctx context.Context, id int, userID int) (*ent.Certificate, error)
	GetByIDWithDNSProvider(ctx context.Context, id int) (*ent.Certificate, error)
	GetByIDForUpdate(ctx context.Context, id int) (*ent.Certificate, error)
	ListByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.Certificate, int, error)
	UpdateCertificate(ctx context.Context, id int, userID int, domains []string, issuedBy, acmeAccountEmail string, renewalDaysBefore int, validationType string) (*ent.Certificate, error)
	UpdateCertificateContent(ctx context.Context, id int, certData, privateKey string, status certificate.Status, issuedAt, expiredAt any) (*ent.Certificate, error)
	UpdateCertificateStatus(ctx context.Context, id int, status certificate.Status) error
	Create(ctx context.Context, domains []string, issuedBy, acmeAccountEmail string, dnsProviderID int, dnsProviderName string, renewalDaysBefore int, validationType string, userID int, serverGroupID int, accessIDs []int) (*ent.Certificate, error)
	Delete(ctx context.Context, id int, userID int) error
	GetCertificateContent(ctx context.Context, id int, userID int) (string, string, error)
	CountByUserID(ctx context.Context, userID int) (int, error)
	GetRenewalEnabledCertificates(ctx context.Context) ([]*ent.Certificate, error)
}
