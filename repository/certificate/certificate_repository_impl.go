package repository

import (
	"context"
	"fmt"
	"time"

	"autoSSL/ent"
	"autoSSL/ent/certificate"
	"autoSSL/ent/user"
)

type certificateRepositoryImpl struct {
	client *ent.Client
}

func NewCertificateRepository(client *ent.Client) CertificateRepository {
	return &certificateRepositoryImpl{client: client}
}

func (r *certificateRepositoryImpl) GetByID(ctx context.Context, id int, userID int) (*ent.Certificate, error) {
	return r.client.Certificate.Query().
		Where(certificate.ID(id), certificate.HasUserWith(user.ID(userID))).
		Only(ctx)
}

func (r *certificateRepositoryImpl) GetByIDWithDNSProvider(ctx context.Context, id int) (*ent.Certificate, error) {
	return r.client.Certificate.Query().
		Where(certificate.ID(id)).
		WithDNSProvider().
		Only(ctx)
}

func (r *certificateRepositoryImpl) GetByIDForUpdate(ctx context.Context, id int) (*ent.Certificate, error) {
	return r.client.Certificate.Get(ctx, id)
}

func (r *certificateRepositoryImpl) ListByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.Certificate, int, error) {
	totalCount, err := r.client.Certificate.Query().
		Where(certificate.HasUserWith(user.ID(userID))).
		Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取总记录数失败: %v", err)
	}

	certificates, err := r.client.Certificate.Query().
		Where(certificate.HasUserWith(user.ID(userID))).
		Order(ent.Desc(certificate.FieldCreatedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取分页数据失败: %v", err)
	}

	return certificates, totalCount, nil
}

func (r *certificateRepositoryImpl) UpdateCertificate(ctx context.Context, id int, userID int, domains []string, issuedBy, acmeAccountEmail string, renewalDaysBefore int, validationType string) (*ent.Certificate, error) {
	update := r.client.Certificate.UpdateOneID(id).
		Where(certificate.HasUserWith(user.ID(userID))).
		SetDomains(domains).
		SetIssuedBy(issuedBy).
		SetAcmeAccountEmail(acmeAccountEmail).
		SetRenewalDaysBefore(renewalDaysBefore).
		SetValidationType(certificate.ValidationType(validationType))

	return update.Save(ctx)
}

func (r *certificateRepositoryImpl) UpdateCertificateContent(ctx context.Context, id int, certData, privateKey string, status certificate.Status, issuedAt, expiredAt any) (*ent.Certificate, error) {
	update := r.client.Certificate.UpdateOneID(id).
		SetCertificate(certData).
		SetPrivateKey(privateKey).
		SetStatus(status)

	if issuedAt != nil {
		update = update.SetIssuedAt(issuedAt.(time.Time))
	}
	if expiredAt != nil {
		update = update.SetExpiredAt(expiredAt.(time.Time))
	}

	return update.Save(ctx)
}

func (r *certificateRepositoryImpl) UpdateCertificateStatus(ctx context.Context, id int, status certificate.Status) error {
	_, err := r.client.Certificate.UpdateOneID(id).SetStatus(status).Save(ctx)
	return err
}

func (r *certificateRepositoryImpl) Create(ctx context.Context, domains []string, issuedBy, acmeAccountEmail string, dnsProviderID int, dnsProviderName string, renewalDaysBefore int, validationType string, userID int, serverGroupID int, accessIDs []int) (*ent.Certificate, error) {
	certCreate := r.client.Certificate.Create().
		SetDomains(domains).
		SetIssuedBy(issuedBy).
		SetAcmeAccountEmail(acmeAccountEmail).
		SetDNSProviderName(dnsProviderName).
		SetRenewalDaysBefore(renewalDaysBefore).
		SetValidationType(certificate.ValidationType(validationType)).
		SetStatus(certificate.StatusPending).
		SetUserID(userID).
		SetCertificate("temp").
		SetPrivateKey("temp")

	if dnsProviderID > 0 {
		certCreate.SetDNSProviderID(dnsProviderID)
	}
	if serverGroupID > 0 {
		certCreate.SetServerGroupID(serverGroupID)
	}
	if len(accessIDs) > 0 {
		certCreate.SetAccessIds(accessIDs)
	}

	return certCreate.Save(ctx)
}

func (r *certificateRepositoryImpl) Delete(ctx context.Context, id int, userID int) error {
	return r.client.Certificate.DeleteOneID(id).
		Where(certificate.HasUserWith(user.ID(userID))).
		Exec(ctx)
}

func (r *certificateRepositoryImpl) GetCertificateContent(ctx context.Context, id int, userID int) (string, string, error) {
	cert, err := r.client.Certificate.Query().
		Where(certificate.ID(id), certificate.HasUserWith(user.ID(userID))).
		Only(ctx)
	if err != nil {
		return "", "", fmt.Errorf("获取证书失败: %v", err)
	}

	return cert.Certificate, cert.PrivateKey, nil
}

func (r *certificateRepositoryImpl) CountByUserID(ctx context.Context, userID int) (int, error) {
	return r.client.Certificate.Query().
		Where(certificate.HasUserWith(user.ID(userID))).
		Count(ctx)
}

func (r *certificateRepositoryImpl) GetRenewalEnabledCertificates(ctx context.Context) ([]*ent.Certificate, error) {
	return r.client.Certificate.Query().
		Where(certificate.RenewalEnabled(true)).
		WithUser().
		WithDNSProvider().
		All(ctx)
}
