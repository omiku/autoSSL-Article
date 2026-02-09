package platform_notification

import (
	"context"
	"fmt"

	"autoSSL/logger"
	"autoSSL/service/global_notification"

	"go.uber.org/zap"
)

// Service 平台通知服务
type Service struct {
	globalService *global_notification.Service
	log           *logger.Logger
}

// NewService 创建平台通知服务
func NewService(globalService *global_notification.Service) *Service {
	return &Service{
		globalService: globalService,
		log:           logger.GetLogger(),
	}
}

// SendRegistrationEmail 发送注册邮件
func (s *Service) SendRegistrationEmail(ctx context.Context, email, username, verificationLink string) error {
	// 检查全局邮件配置是否有效
	config, err := s.globalService.GetGlobalEmailConfig(ctx)
	if err != nil {
		return fmt.Errorf("获取全局邮件配置失败: %w", err)
	}

	if !config.IsValid() {
		return fmt.Errorf("全局邮件配置无效")
	}

	subject := "欢迎注册平台 - 邮箱验证"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">欢迎注册平台</h2>
			<p>亲爱的 %s，</p>
			<p>感谢您注册我们的平台！请点击下面的链接完成邮箱验证：</p>
			<div style="margin: 20px 0; text-align: center;">
				<a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">验证邮箱</a>
			</div>
			<p style="color: #666; font-size: 12px;">如果链接无法点击，请复制以下地址到浏览器打开：</p>
			<p style="color: #666; font-size: 12px; word-break: break-all;">%s</p>
			<p>此链接将在24小时内有效。</p>
			<p>如有疑问，请联系管理员。</p>
		</div>
	`, username, verificationLink, verificationLink)

	return s.globalService.SendEmailWithGlobalConfig(ctx, []string{email}, subject, body)
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *Service) SendPasswordResetEmail(ctx context.Context, email, username, resetLink string) error {
	config, err := s.globalService.GetGlobalEmailConfig(ctx)
	if err != nil {
		return fmt.Errorf("获取全局邮件配置失败: %w", err)
	}

	if !config.IsValid() {
		return fmt.Errorf("全局邮件配置无效")
	}

	subject := "密码重置请求"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">密码重置</h2>
			<p>亲爱的 %s，</p>
			<p>我们收到了您的密码重置请求。请点击下面的链接重置密码：</p>
			<div style="margin: 20px 0; text-align: center;">
				<a href="%s" style="background-color: #dc3545; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">重置密码</a>
			</div>
			<p style="color: #666; font-size: 12px;">如果链接无法点击，请复制以下地址到浏览器打开：</p>
			<p style="color: #666; font-size: 12px; word-break: break-all;">%s</p>
			<p>此链接将在1小时内有效。</p>
			<p>如果您没有发起此请求，请忽略此邮件。</p>
		</div>
	`, username, resetLink, resetLink)

	return s.globalService.SendEmailWithGlobalConfig(ctx, []string{email}, subject, body)
}

// SendRegistrationSMS 发送注册短信
func (s *Service) SendRegistrationSMS(ctx context.Context, phone, username, verificationCode string) error {
	config, err := s.globalService.GetGlobalSMSConfig(ctx)
	if err != nil {
		return fmt.Errorf("获取全局短信配置失败: %w", err)
	}

	if !config.IsValid() {
		return fmt.Errorf("全局短信配置无效")
	}

	// 这里需要根据具体的短信服务提供商实现发送逻辑
	// 以下是一个示例实现
	message := fmt.Sprintf("【%s】欢迎注册平台，您的验证码是：%s，请在5分钟内完成验证。", config.SignName, verificationCode)

	// 实际调用短信服务API
	// 这里需要集成具体的短信服务商SDK
	s.log.Info("发送注册短信",
		zap.String("phone", phone),
		zap.String("message", message),
		zap.String("template", config.TemplateCode),
	)

	// TODO: 集成具体的短信服务商API
	return nil
}

// SendPasswordResetSMS 发送密码重置短信
func (s *Service) SendPasswordResetSMS(ctx context.Context, phone, username, resetCode string) error {
	config, err := s.globalService.GetGlobalSMSConfig(ctx)
	if err != nil {
		return fmt.Errorf("获取全局短信配置失败: %w", err)
	}

	if !config.IsValid() {
		return fmt.Errorf("全局短信配置无效")
	}

	message := fmt.Sprintf("【%s】密码重置验证码：%s，请在5分钟内完成验证。如非本人操作，请忽略此短信。", config.SignName, resetCode)

	s.log.Info("发送密码重置短信",
		zap.String("phone", phone),
		zap.String("message", message),
		zap.String("template", config.TemplateCode),
	)

	// TODO: 集成具体的短信服务商API
	return nil
}

// SendSystemNotification 发送系统通知
func (s *Service) SendSystemNotification(ctx context.Context, email, subject, message string) error {
	return s.globalService.SendEmailWithGlobalConfig(ctx, []string{email}, subject, message)
}
