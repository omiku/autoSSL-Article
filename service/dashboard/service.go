package dashboard

import (
	"context"
	"fmt"
	"time"

	"autoSSL/ent"
	"autoSSL/ent/access"
	"autoSSL/ent/certificate"
	"autoSSL/ent/dnsprovider"
	"autoSSL/ent/servergroup"
	"autoSSL/ent/user"
	"autoSSL/ent/workflow"
)

// Service 仪表板服务层
type Service struct {
	client *ent.Client
}

// NewService 创建仪表板服务实例
func NewService(client *ent.Client) *Service {
	return &Service{client: client}
}

// CountStats 数量统计数据结构
type CountStats struct {
	TotalCertificates int `json:"total_certificates"`
	TotalAccesses     int `json:"total_accesses"`
	TotalServerGroups int `json:"total_server_groups"`
	TotalDNSProviders int `json:"total_dns_providers"`
}

// DeploymentInfo 部署信息数据结构
type DeploymentInfo struct {
	CertificateName string    `json:"certificate_name"`
	DeploymentType  string    `json:"deployment_type"` // "server_group" 或 "access"
	TargetName      string    `json:"target_name"`
	LastUpdated     time.Time `json:"last_updated"`
	Status          string    `json:"status"`
}

// PendingWorkflow 待开始工作流数据结构
type PendingWorkflow struct {
	CertificateName string    `json:"certificate_name"`
	DeploymentType  string    `json:"deployment_type"` // "server_group" 或 "access"
	TargetName      string    `json:"target_name"`
	CreatedAt       time.Time `json:"created_at"`
}

// GetCountStats 获取用户数量统计数据
func (s *Service) GetCountStats(ctx context.Context, userID int) (*CountStats, error) {
	stats := &CountStats{}

	// 获取总证书数
	certCount, err := s.client.Certificate.Query().
		Where(certificate.HasUserWith(user.IDEQ(userID))).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalCertificates = certCount

	// 获取服务组数量
	serverGroupCount, err := s.client.ServerGroup.Query().
		Where(servergroup.HasUserWith(user.IDEQ(userID))).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalServerGroups = serverGroupCount

	// 获取DNS供应商数量
	dnsProviderCount, err := s.client.DNSProvider.Query().
		Where(dnsprovider.HasUserWith(user.IDEQ(userID))).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalDNSProviders = dnsProviderCount

	// 获取access数量
	accessCount, err := s.client.Access.Query().
		Where(access.HasUserWith(user.IDEQ(userID))).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalAccesses = accessCount

	return stats, nil
}

// GetRecentDeployments 获取最近部署状态
func (s *Service) GetRecentDeployments(ctx context.Context, userID int) ([]*DeploymentInfo, error) {
	// 获取最近5条工作流记录（按更新时间排序）
	workflows, err := s.client.Workflow.Query().
		Where(
			workflow.HasCertificateWith(
				certificate.HasUserWith(user.IDEQ(userID)),
			),
		).
		WithCertificate().
		Order(ent.Desc(workflow.FieldUpdatedAt)).
		Limit(5).
		All(ctx)
	if err != nil {
		return nil, err
	}

	var result []*DeploymentInfo
	for _, wf := range workflows {
		info := &DeploymentInfo{
			CertificateName: getDomainDescription(wf.Edges.Certificate.Domains),
			LastUpdated:     wf.UpdatedAt,
			Status:          string(wf.Status),
		}

		// 获取证书关联的服务组和访问
		cert := wf.Edges.Certificate

		// 检查服务器组
		if cert.ServerGroupID > 0 {
			serverGroup, err := s.client.ServerGroup.Get(ctx, cert.ServerGroupID)
			if err == nil && serverGroup != nil {
				info.DeploymentType = "server_group"
				info.TargetName = serverGroup.Name
			}
		}

		// 如果没有服务器组，检查访问配置
		if info.DeploymentType == "" && len(cert.AccessIds) > 0 && cert.AccessIds[0] > 0 {
			access, err := s.client.Access.Get(ctx, cert.AccessIds[0])
			if err == nil && access != nil {
				info.DeploymentType = "access"
				info.TargetName = access.Name
			}
		}

		result = append(result, info)
	}

	return result, nil
}

// getDomainDescription 从域名列表生成描述
func getDomainDescription(domains []string) string {
	if len(domains) == 0 {
		return "未指定域名"
	}
	if len(domains) == 1 {
		return domains[0]
	}

	// 多域名时，返回第一个域名加上数量
	description := domains[0]
	if len(domains) > 1 {
		description += fmt.Sprintf(" (+%d个域名)", len(domains)-1)
	}
	return description
}

// GetPendingWorkflows 获取待开始工作流
func (s *Service) GetPendingWorkflows(ctx context.Context, userID int) ([]*PendingWorkflow, error) {
	// 获取最近5条待开始的工作流（状态为pending）
	workflows, err := s.client.Workflow.Query().
		Where(
			workflow.HasCertificateWith(
				certificate.HasUserWith(user.IDEQ(userID)),
			),
			workflow.StatusEQ("pending"),
		).
		WithCertificate().
		Order(ent.Desc(workflow.FieldCreatedAt)).
		Limit(5).
		All(ctx)
	if err != nil {
		return nil, err
	}

	var result []*PendingWorkflow
	for _, wf := range workflows {
		pending := &PendingWorkflow{
			CertificateName: getDomainDescription(wf.Edges.Certificate.Domains),
			CreatedAt:       wf.CreatedAt,
		}

		// 判断部署类型
		cert := wf.Edges.Certificate

		if cert.ServerGroupID > 0 {
			serverGroup, err := s.client.ServerGroup.Get(ctx, cert.ServerGroupID)
			if err == nil && serverGroup != nil {
				pending.DeploymentType = "server_group"
				pending.TargetName = serverGroup.Name
			}
		} else if len(cert.AccessIds) > 0 && cert.AccessIds[0] > 0 {
			access, err := s.client.Access.Get(ctx, cert.AccessIds[0])
			if err == nil && access != nil {
				pending.DeploymentType = "access"
				pending.TargetName = access.Name
			}
		}

		result = append(result, pending)
	}

	return result, nil
}
