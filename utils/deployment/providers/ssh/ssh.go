package ssh

import (
	"autoSSL/logger"
	"autoSSL/utils/deployment/config"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"github.com/povsister/scp"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// OutputFormatType 证书输出格式类型
type OutputFormatType string

const (
	OUTPUT_FORMAT_PEM OutputFormatType = "pem"
	OUTPUT_FORMAT_PFX OutputFormatType = "pfx"
	OUTPUT_FORMAT_JKS OutputFormatType = "jks"
	OUTPUT_FORMAT_DER OutputFormatType = "der"
)

// SSLDeployerProviderConfig SSH部署提供者配置
type SSLDeployerProviderConfig struct {
	// SSH连接配置
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Key           string `json:"key"`
	KeyPassphrase string `json:"key_passphrase"`
	SshAuthMethod string `json:"ssh_auth_method"`

	// 跳板机配置
	JumpServers []JumpServerConfig `json:"jump_servers"`

	// 部署配置
	UseSCP                   bool             `json:"use_scp"`
	PreCommand               string           `json:"pre_command"`
	PostCommand              string           `json:"post_command"`
	OutputFormat             OutputFormatType `json:"output_format"`
	OutputCertPath           string           `json:"output_cert_path"`
	OutputServerCertPath     string           `json:"output_server_cert_path"`
	OutputIntermediaCertPath string           `json:"output_intermedia_cert_path"`
	OutputKeyPath            string           `json:"output_key_path"`
	PfxPassword              string           `json:"pfx_password"`
	JksAlias                 string           `json:"jks_alias"`
	JksKeypass               string           `json:"jks_keypass"`
	JksStorepass             string           `json:"jks_storepass"`
}

func (c *SSLDeployerProviderConfig) validate() error {
	if c.Host == "" {
		return fmt.Errorf("SSH主机地址不能为空")
	}
	if c.Username == "" {
		return fmt.Errorf("SSH用户名不能为空")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("SSH端口必须在1-65535之间")
	}
	if c.Password == "" && c.Key == "" {
		return fmt.Errorf("必须提供密码或私钥认证方式")
	}
	return nil
}

// JumpServerConfig 跳板机配置
type JumpServerConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Key      string `json:"key"`
}

// SSHDeployer SSH部署器
type SSHDeployer struct {
	config *SSLDeployerProviderConfig // 改为指针
	logger *logger.Logger
	client *ssh.Client
	mu     sync.Mutex
}

// NewSSLDeployerProvider 创建新的SSH部署提供者实例
func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSHDeployer, error) {
	// 默认值
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 22
	}
	if config.OutputFormat == "" {
		config.OutputFormat = OUTPUT_FORMAT_PEM
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}
	return &SSHDeployer{
		config: config,
		logger: logger.GetLogger(),
	}, nil
}

// Name 返回部署提供者名称
func (s *SSHDeployer) Name() string {
	return "ssh"
}

// DeploymentConfig 部署配置参数结构体
type DeploymentConfig struct {
	RemoteCertPath string `json:"remote_cert_path"`
	RemoteKeyPath  string `json:"remote_key_path"`
	PreCommand     string `json:"pre_command"`
	PostCommand    string `json:"post_command"`
}

// DeployCertificate 部署证书到SSH目标服务器
func (s *SSHDeployer) DeployCertificate(ctx context.Context, certContent, keyContent string, workflowConfig map[string]interface{}) error {
	s.logger.Debug("开始SSL证书部署",
		zap.String("host", s.config.Host),
		zap.Int("port", s.config.Port),
		zap.String("username", s.config.Username),
		zap.Int("jumpServers", len(s.config.JumpServers)))

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	defer s.closeConnection()

	// 解析部署配置
	deployConfig := DeploymentConfig{}
	if err := config.PopulateConfig(workflowConfig, &deployConfig); err != nil {
		s.logger.Debug("解析部署配置失败",
			zap.String("host", s.config.Host),
			zap.Int("port", s.config.Port),
			zap.Error(err))
		return fmt.Errorf("[CONFIG_ERROR] 解析部署配置失败 | 主机: %s:%d | 错误: %w", s.config.Host, s.config.Port, err)
	}
	if deployConfig.RemoteCertPath == "" {
		s.logger.Debug("remote_cert_path配置缺失",
			zap.String("host", s.config.Host),
			zap.Int("port", s.config.Port))
		return fmt.Errorf("[CONFIG_ERROR] remote_cert_path配置缺失 | 主机: %s:%d", s.config.Host, s.config.Port)
	}
	if deployConfig.RemoteKeyPath == "" {
		s.logger.Debug("remote_key_path配置缺失",
			zap.String("host", s.config.Host),
			zap.Int("port", s.config.Port))
		return fmt.Errorf("[CONFIG_ERROR] remote_key_path配置缺失 | 主机: %s:%d", s.config.Host, s.config.Port)
	}

	client, err := s.ensureConnection(ctx)
	if err != nil {
		return err
	}
	s.logger.Info("SSH连接成功",
		zap.String("remoteAddress", client.RemoteAddr().String()))

	if deployConfig.PreCommand != "" {
		s.logger.Info("执行前置命令",
			zap.String("host", s.config.Host),
			zap.String("command", deployConfig.PreCommand))
		if err := s.executeCommandWithTimeout(ctx, client, deployConfig.PreCommand); err != nil {
			s.logger.Debug("执行前置命令失败",
				zap.String("host", s.config.Host),
				zap.String("command", deployConfig.PreCommand),
				zap.Error(err))
			return fmt.Errorf("[PRE_COMMAND_ERROR] 执行前置命令失败 | 主机: %s | 命令: %s | 错误: %v", s.config.Host, deployConfig.PreCommand, err)
		}
	}

	s.logger.Info("开始部署证书文件",
		zap.String("host", s.config.Host),
		zap.String("remotePath", deployConfig.RemoteCertPath))
	if err := s.deployFileWithFallback(ctx, client, certContent, deployConfig.RemoteCertPath, 0644); err != nil {
		s.logger.Debug("部署证书文件失败",
			zap.String("host", s.config.Host),
			zap.String("remotePath", deployConfig.RemoteCertPath),
			zap.Error(err))
		return fmt.Errorf("[CERT_DEPLOY_ERROR] 部署证书文件失败 | 主机: %s | 路径: %s | 错误: %w", s.config.Host, deployConfig.RemoteCertPath, err)
	}

	s.logger.Info("开始部署私钥文件",
		zap.String("host", s.config.Host),
		zap.String("remotePath", deployConfig.RemoteKeyPath))
	if err := s.deployFileWithFallback(ctx, client, keyContent, deployConfig.RemoteKeyPath, 0600); err != nil {
		s.logger.Debug("部署私钥文件失败",
			zap.String("host", s.config.Host),
			zap.String("remotePath", deployConfig.RemoteKeyPath),
			zap.Error(err))
		return fmt.Errorf("[KEY_DEPLOY_ERROR] 部署私钥文件失败 | 主机: %s | 路径: %s | 错误: %w", s.config.Host, deployConfig.RemoteKeyPath, err)
	}

	if deployConfig.PostCommand != "" {
		s.logger.Info("执行后置命令",
			zap.String("host", s.config.Host),
			zap.String("command", deployConfig.PostCommand))
		if err := s.executeCommandWithTimeout(ctx, client, deployConfig.PostCommand); err != nil {
			s.logger.Debug("执行后置命令失败",
				zap.String("host", s.config.Host),
				zap.String("command", deployConfig.PostCommand),
				zap.Error(err))
			return fmt.Errorf("[POST_COMMAND_ERROR] 执行后置命令失败 | 主机: %s | 命令: %s | 错误: %v", s.config.Host, deployConfig.PostCommand, err)
		}
	}

	s.logger.Info("SSL证书部署完成",
		zap.String("host", s.config.Host),
		zap.Int("port", s.config.Port))
	return nil
}

// ---------- 以下为内部辅助方法 ----------

func (s *SSHDeployer) ensureConnection(ctx context.Context) (*ssh.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查现有连接是否有效
	if s.client != nil {
		if s.isConnectionAlive() {
			s.logger.Debug("现有SSH连接有效")
			return s.client, nil
		}
		// 关闭无效连接
		s.logger.Debug("关闭无效的SSH连接")
		s.client.Close()
		s.client = nil
	}

	client, err := s.createSSHConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("[SSH_RECONNECT_ERROR] SSH连接重建失败: %w", err)
	}
	s.client = client
	s.logger.Info("SSH连接已重新建立",
		zap.String("host", s.config.Host),
		zap.Int("port", s.config.Port))
	return s.client, nil
}

func (s *SSHDeployer) isConnectionAlive() bool {
	if s.client == nil {
		return false
	}
	session, err := s.client.NewSession()
	if err != nil {
		s.logger.Debug("连接检查失败",
			zap.String("host", s.config.Host),
			zap.Int("port", s.config.Port),
			zap.Error(err))
		return false
	}
	session.Close()
	return true
}

func (s *SSHDeployer) closeConnection() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		s.logger.Info("关闭SSH连接",
			zap.String("host", s.config.Host),
			zap.Int("port", s.config.Port))
		s.client.Close()
		s.client = nil
	}
}

func (s *SSHDeployer) createSSHConnection(ctx context.Context) (*ssh.Client, error) {
	s.logger.Info("建立新的SSH连接",
		zap.String("host", s.config.Host),
		zap.Int("port", s.config.Port))
	cfg, err := s.buildSSHClientConfig()
	if err != nil {
		return nil, fmt.Errorf("构建SSH配置失败: %w", err)
	}
	if len(s.config.JumpServers) > 0 {
		return s.createJumpConnection(ctx, cfg)
	}
	return s.createDirectConnection(ctx, cfg)
}

func (s *SSHDeployer) buildSSHClientConfig() (*ssh.ClientConfig, error) {
	var auths []ssh.AuthMethod
	if s.config.Password != "" {
		auths = append(auths, ssh.Password(s.config.Password))
	}
	if s.config.Key != "" {
		var signer ssh.Signer
		var err error
		if s.config.KeyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(s.config.Key), []byte(s.config.KeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(s.config.Key))
		}
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}
	if len(auths) == 0 {
		return nil, fmt.Errorf("未配置认证方式(password或private_key)")
	}
	return &ssh.ClientConfig{
		User:    s.config.Username,
		Auth:    auths,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil // 生产环境应验证
		},
	}, nil
}

func (s *SSHDeployer) createDirectConnection(ctx context.Context, cfg *ssh.ClientConfig) (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	var client *ssh.Client
	var err error
	done := make(chan struct{})
	go func() {
		client, err = ssh.Dial("tcp", addr, cfg)
		close(done)
	}()
	select {
	case <-done:
		return client, err
	case <-ctx.Done():
		return nil, fmt.Errorf("连接超时: %v", ctx.Err())
	}
}

func (s *SSHDeployer) createJumpConnection(ctx context.Context, cfg *ssh.ClientConfig) (*ssh.Client, error) {
	var current *ssh.Client
	for i, j := range s.config.JumpServers {
		jCfg := &ssh.ClientConfig{
			User:            j.Username,
			Auth:            []ssh.AuthMethod{},
			Timeout:         30 * time.Second,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		if j.Key != "" {
			signer, err := ssh.ParsePrivateKey([]byte(j.Key))
			if err != nil {
				return nil, fmt.Errorf("跳板机私钥解析失败: %w", err)
			}
			jCfg.Auth = append(jCfg.Auth, ssh.PublicKeys(signer))
		} else {
			jCfg.Auth = append(jCfg.Auth, ssh.Password(j.Password))
		}
		addr := fmt.Sprintf("%s:%d", j.Host, j.Port)
		if i == 0 {
			c, err := ssh.Dial("tcp", addr, jCfg)
			if err != nil {
				return nil, fmt.Errorf("连接跳板机失败: %w", err)
			}
			current = c
		} else {
			conn, err := current.Dial("tcp", addr)
			if err != nil {
				return nil, fmt.Errorf("跳板机连接失败: %w", err)
			}
			cc, chans, reqs, err := ssh.NewClientConn(conn, addr, jCfg)
			if err != nil {
				return nil, fmt.Errorf("跳板机SSH握手失败: %w", err)
			}
			current.Close()
			current = ssh.NewClient(cc, chans, reqs)
		}
	}
	finalConn, err := current.Dial("tcp", fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))
	if err != nil {
		return nil, fmt.Errorf("连接目标服务器失败: %w", err)
	}
	finalCC, chans, reqs, err := ssh.NewClientConn(finalConn, fmt.Sprintf("%s:%d", s.config.Host, s.config.Port), cfg)
	if err != nil {
		return nil, fmt.Errorf("目标服务器SSH握手失败: %w", err)
	}
	return ssh.NewClient(finalCC, chans, reqs), nil
}

func (s *SSHDeployer) executeCommandWithTimeout(ctx context.Context, client *ssh.Client, cmd string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("[SSH_ERROR] 创建会话失败 | 远程地址: %s | 错误: %v", client.RemoteAddr(), err)
	}
	defer session.Close()

	done := make(chan error, 1)
	go func() {
		output, err := session.CombinedOutput(cmd)
		if err != nil {
			err = fmt.Errorf(`[SSH_COMMAND_ERROR] 命令执行失败 | 远程地址: %s | 命令: %s | 输出: %s | 错误: %v`,
				client.RemoteAddr(), cmd, string(output), err)
		}
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("[SSH_TIMEOUT] 命令执行超时 | 远程地址: %s | 命令: %s", client.RemoteAddr(), cmd)
	}
}

func (s *SSHDeployer) deployFileWithFallback(ctx context.Context, client *ssh.Client, content, remotePath string, fileMode os.FileMode) error {
	remoteAddr := client.RemoteAddr().String()
	contentSize := len(content)
	s.config.UseSCP = true
	// 先尝试 SCP
	if s.config.UseSCP {
		if err := s.scpSendContentWithPermission(ctx, client, content, remotePath, fileMode); err == nil {
			s.logger.Info("SCP传输成功",
				zap.String("remoteAddress", remoteAddr),
				zap.String("remotePath", remotePath),
				zap.Int("fileSize", contentSize))
			s.mu.Lock()
			client.Close()
			s.client = nil //提前关闭置空，以便重新生成链接
			s.mu.Unlock()
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		} else {
			s.logger.Debug("SCP传输失败，回退到SFTP",
				zap.String("remoteAddress", remoteAddr),
				zap.Error(err))
		}
	}
	// 回退 SFTP
	if err := s.sftpSendContentWithPermission(ctx, client, content, remotePath, fileMode); err != nil {
		s.logger.Debug("文件部署失败",
			zap.String("remoteAddress", remoteAddr),
			zap.String("remotePath", remotePath),
			zap.Error(err))
		return fmt.Errorf("[DEPLOY_FAILED] 文件部署失败 | 远程地址: %s | 远程路径: %s | 错误: %w", remoteAddr, remotePath, err)
	}
	s.logger.Info("SFTP回退成功",
		zap.String("remoteAddress", remoteAddr),
		zap.String("remotePath", remotePath),
		zap.Int("fileSize", contentSize))
	return nil
}

func (s *SSHDeployer) scpSendContentWithPermission(ctx context.Context, client *ssh.Client, content, remotePath string, fileMode os.FileMode) error {
	remoteAddr := client.RemoteAddr().String()
	scpClient, err := scp.NewClientFromExistingSSH(client, &scp.ClientOption{})
	if err != nil {
		s.logger.Debug("创建SCP客户端失败",
			zap.String("remoteAddress", remoteAddr),
			zap.Error(err))
		return fmt.Errorf("[SCP_ERROR] 创建SCP客户端失败 | 远程地址: %s | 错误: %v", remoteAddr, err)
	}
	defer scpClient.Close()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err = scpClient.CopyToRemote(strings.NewReader(content), remotePath, &scp.FileTransferOption{
		Context: ctx,
		Timeout: 3 * time.Minute,
		Perm:    fileMode,
	})
	if err != nil {
		s.logger.Debug("SCP传输失败",
			zap.String("remoteAddress", remoteAddr),
			zap.String("remotePath", remotePath),
			zap.Error(err))
		return fmt.Errorf("[SCP_TRANSFER_ERROR] SCP传输失败 | 远程地址: %s | 远程路径: %s | 错误: %v", remoteAddr, remotePath, err)
	}
	return nil
}

func (s *SSHDeployer) sftpSendContentWithPermission(ctx context.Context, client *ssh.Client, content, remotePath string, fileMode os.FileMode) error {
	remoteAddr := client.RemoteAddr().String()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		s.logger.Debug("创建SFTP客户端失败",
			zap.String("remoteAddress", remoteAddr),
			zap.Error(err))
		return fmt.Errorf("[SFTP_ERROR] 创建SFTP客户端失败 | 远程地址: %s | 错误: %v", remoteAddr, err)
	}
	defer sftpClient.Close()

	// 创建目录（如果不存在）
	dir := filepath.Dir(remotePath)
	if dir != "." && dir != "/" {
		if err := sftpClient.MkdirAll(dir); err != nil {
			s.logger.Debug("创建目录失败",
				zap.String("remoteAddress", remoteAddr),
				zap.String("directory", dir),
				zap.Error(err))
			return fmt.Errorf("[SFTP_MKDIR_ERROR] 创建目录失败 | 远程地址: %s | 目录: %s | 错误: %v", remoteAddr, dir, err)
		}
	}

	// 创建文件
	file, err := sftpClient.Create(remotePath)
	if err != nil {
		s.logger.Debug("创建文件失败",
			zap.String("remoteAddress", remoteAddr),
			zap.String("remotePath", remotePath),
			zap.Error(err))
		return fmt.Errorf("[SFTP_CREATE_ERROR] 创建文件失败 | 远程地址: %s | 远程路径: %s | 错误: %v", remoteAddr, remotePath, err)
	}
	defer file.Close()

	// 设置权限
	if err := file.Chmod(fileMode); err != nil {
		s.logger.Debug("设置权限失败",
			zap.String("remoteAddress", remoteAddr),
			zap.String("remotePath", remotePath),
			zap.Error(err))
		return fmt.Errorf("[SFTP_CHMOD_ERROR] 设置权限失败 | 远程地址: %s | 远程路径: %s | 错误: %v", remoteAddr, remotePath, err)
	}

	// 写入内容
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	_, err = file.Write([]byte(content))
	if err != nil {
		s.logger.Debug("写入内容失败",
			zap.String("remoteAddress", remoteAddr),
			zap.String("remotePath", remotePath),
			zap.Error(err))
		return fmt.Errorf("[SFTP_WRITE_ERROR] 写入内容失败 | 远程地址: %s | 远程路径: %s | 错误: %v", remoteAddr, remotePath, err)
	}
	return nil
}
