package certificate

import (
	"autoSSL/ent"
	"autoSSL/ent/dnsprovider"
	"autoSSL/ent/user"
	certificateRepo "autoSSL/repository/certificate"
	"autoSSL/service/workflow"
	"context"
	"fmt"
	"log"
	"strings"
)

type CertificateService struct {
	client          *ent.Client
	workflowMgr     *workflow.Manager
	certificateRepo certificateRepo.CertificateRepository
}

// NewCertificateService 创建证书服务实例
func NewCertificateService(client *ent.Client, workflowMgr *workflow.Manager, certificateRepo certificateRepo.CertificateRepository) *CertificateService {
	return &CertificateService{
		client:          client,
		workflowMgr:     workflowMgr,
		certificateRepo: certificateRepo,
	}
}

// GetAllWithPagination 获取用户的证书列表（完整分页信息）
// 返回证书列表、总记录数、总页数等分页信息
func (s *CertificateService) GetAllWithPagination(ctx context.Context, userID int, page int, pageSize int) ([]*ent.Certificate, int, error) {
	return s.certificateRepo.ListByUserID(ctx, userID, page, pageSize)
}

func (s *CertificateService) GetByID(ctx context.Context, userID int, id int) (*ent.Certificate, error) {
	return s.certificateRepo.GetByID(ctx, id, userID)
}

func (s *CertificateService) UpdateCertificate(ctx context.Context, userID int, id int, req workflow.CertificateRequest) (*ent.Certificate, error) {
	return s.certificateRepo.UpdateCertificate(ctx, id, userID, req.Domains, req.IssuedBy, req.AcmeAccountEmail, req.RenewalDaysBefore, req.ValidationType)
}

func (s *CertificateService) Delete(ctx context.Context, userID int, id int) error {
	return s.certificateRepo.Delete(ctx, id, userID)
}

// CreateCertificate 创建证书
func (s *CertificateService) CreateCertificate(ctx context.Context, req workflow.CertificateRequest) (*ent.Certificate, error) {
	dnsProvider, err := s.client.DNSProvider.Query().
		Where(dnsprovider.ID(req.DNSProviderID), dnsprovider.HasUserWith(user.ID(req.UserID))).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("DNS供应商验证失败: %v", err)
	}
	user, err := s.client.User.Query().
		Select(user.FieldEmail).
		Where(user.ID(req.UserID)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("用户验证失败: %v", err)
	}

	hasDeploymentConfig := req.ServerGroupID > 0 || len(req.AccessIDs) > 0
	totalSteps := 4
	if hasDeploymentConfig {
		totalSteps = 7
	}

	domainDesc := strings.Join(req.Domains, ", ")
	if len(domainDesc) > 100 {
		domainDesc = domainDesc[:97] + "..."
	}

	cert, err := s.certificateRepo.Create(ctx, req.Domains, req.IssuedBy, user.Email, req.DNSProviderID, dnsProvider.ProviderType.String(), req.RenewalDaysBefore, req.ValidationType, req.UserID, req.ServerGroupID, req.AccessIDs)
	if err != nil {
		return nil, fmt.Errorf("创建证书记录失败: %v", err)
	}

	log.Printf("成功创建证书记录，ID: %d, 域名: %s, 用户: %d", cert.ID, domainDesc, req.UserID)

	workflow, err := s.workflowMgr.CreateWorkflow(context.Background(), cert.ID, req.Domains, string(workflow.StatusPending), req.UserID, req.ExtraConfig, totalSteps, req.UnifiedNotificationConfig)
	if err != nil {
		return nil, fmt.Errorf("创建工作流记录失败: %v", err)
	}

	if req.ExecuteImmediately {
		if err := s.workflowMgr.StartCertificateWorkflow(workflow.ID, req); err != nil {
			return nil, fmt.Errorf("启动证书工作流失败: %v", err)
		}
	} else {
		log.Printf("Certificate workflow %d is scheduled for later execution", workflow.ID)
	}

	return cert, nil
}

// GetCertificateContent 获取证书内容（证书和私钥）
func (s *CertificateService) GetCertificateContent(ctx context.Context, userID int, id int) (certContent string, privateKeyContent string, err error) {
	certData, privateKey, err := s.certificateRepo.GetCertificateContent(ctx, id, userID)
	if err != nil {
		return "", "", fmt.Errorf("获取证书失败: %v", err)
	}

	cert, err := s.certificateRepo.GetByID(ctx, id, userID)
	if err != nil {
		return "", "", fmt.Errorf("获取证书失败: %v", err)
	}

	if cert.Status != "completed" && cert.Status != "issued" {
		return "", "", fmt.Errorf("证书状态无效: %s", cert.Status)
	}

	return certData, privateKey, nil
}
