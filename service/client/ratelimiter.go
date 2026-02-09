package client

import (
	"sync"
	"time"
)

// RateLimiter 管理不同证书提供商的速率限制
// Let's Encrypt: 单账户每3小时300份证书，36秒恢复1份
// ZeroSSL: ACME端点每秒最多7个请求，其他操作无限制
const (
	// Let's Encrypt 限制
	LetsEncryptCertLimitWindow   = 3 * time.Hour
	LetsEncryptCertLimitMax      = 300
	LetsEncryptCertLimitRecovery = 36 * time.Second // 每36秒恢复1份

	// ZeroSSL ACME限制
	ZeroSSLRateLimitWindow = 1 * time.Second
	ZeroSSLRateLimitMax    = 7 // 每秒最多7个请求

	// 新账户限制：每用户每24小时最多3个新ACME账户（通用）
	AccountLimitWindow = 24 * time.Hour
	AccountLimitMax    = 3
)

// RateLimiter 全局速率限制器实例
var (
	GlobalRateLimiter = NewRateLimiter()
)

// RateLimiter 结构体管理速率限制
type RateLimiter struct {
	mu sync.RWMutex

	// 证书申请限制（按提供商区分）
	letsEncryptRequests []time.Time // Let's Encrypt证书申请
	zeroSSLRequests     []time.Time // ZeroSSL请求（ACME端点限制）

	// 账户创建限制
	accountCreations map[string][]time.Time // 用户ID -> 账户创建时间列表
	accountMu        sync.RWMutex
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		letsEncryptRequests: make([]time.Time, 0),
		zeroSSLRequests:     make([]time.Time, 0),
		accountCreations:    make(map[string][]time.Time),
	}
}

// AllowCertRequest 检查是否允许新的证书申请（按提供商区分）
// provider: 证书提供商（letsencrypt, zerossl）
// 返回是否允许，以及需要等待的时间（如果不允许）
func (rl *RateLimiter) AllowCertRequest(provider string) (bool, time.Duration) {
	switch provider {
	case "letsencrypt":
		return rl.allowLetsEncryptCertRequest()
	case "zerossl":
		return rl.allowZeroSSLRequest()
	default:
		// 默认使用Let's Encrypt限制
		return rl.allowLetsEncryptCertRequest()
	}
}

// allowLetsEncryptCertRequest 检查Let's Encrypt证书申请限制
func (rl *RateLimiter) allowLetsEncryptCertRequest() (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	
	// 清理过期的请求记录
	windowStart := now.Add(-LetsEncryptCertLimitWindow)
	validRequests := make([]time.Time, 0)
	for _, t := range rl.letsEncryptRequests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	rl.letsEncryptRequests = validRequests

	// 检查是否达到限制
	if len(rl.letsEncryptRequests) >= LetsEncryptCertLimitMax {
		// 计算下一个可用时间
		oldest := rl.letsEncryptRequests[0]
		waitTime := oldest.Add(LetsEncryptCertLimitWindow).Sub(now)
		return false, waitTime
	}

	return true, 0
}

// allowZeroSSLRequest 检查ZeroSSL ACME端点限制
func (rl *RateLimiter) allowZeroSSLRequest() (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	
	// 清理过期的请求记录
	windowStart := now.Add(-ZeroSSLRateLimitWindow)
	validRequests := make([]time.Time, 0)
	for _, t := range rl.zeroSSLRequests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	rl.zeroSSLRequests = validRequests

	// 检查是否达到限制
	if len(rl.zeroSSLRequests) >= ZeroSSLRateLimitMax {
		// 计算下一个可用时间
		oldest := rl.zeroSSLRequests[0]
		waitTime := oldest.Add(ZeroSSLRateLimitWindow).Sub(now)
		return false, waitTime
	}

	return true, 0
}

// RecordCertRequest 记录一次证书申请（按提供商区分）
// provider: 证书提供商（letsencrypt, zerossl）
func (rl *RateLimiter) RecordCertRequest(provider string) {
	switch provider {
	case "letsencrypt":
		rl.recordLetsEncryptCertRequest()
	case "zerossl":
		rl.recordZeroSSLRequest()
	default:
		rl.recordLetsEncryptCertRequest()
	}
}

// recordLetsEncryptCertRequest 记录Let's Encrypt证书申请
func (rl *RateLimiter) recordLetsEncryptCertRequest() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	
	// 清理过期的请求
	windowStart := now.Add(-LetsEncryptCertLimitWindow)
	validRequests := make([]time.Time, 0)
	for _, t := range rl.letsEncryptRequests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	
	// 添加新请求
	validRequests = append(validRequests, now)
	rl.letsEncryptRequests = validRequests
}

// recordZeroSSLRequest 记录ZeroSSL请求
func (rl *RateLimiter) recordZeroSSLRequest() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	
	// 清理过期的请求
	windowStart := now.Add(-ZeroSSLRateLimitWindow)
	validRequests := make([]time.Time, 0)
	for _, t := range rl.zeroSSLRequests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	
	// 添加新请求
	validRequests = append(validRequests, now)
	rl.zeroSSLRequests = validRequests
}

// AllowAccountCreation 检查用户是否允许创建新ACME账户
func (rl *RateLimiter) AllowAccountCreation(userID string) (bool, time.Duration) {
	rl.accountMu.Lock()
	defer rl.accountMu.Unlock()
	
	now := time.Now()
	
	// 获取用户的账户创建记录
	userCreations, exists := rl.accountCreations[userID]
	if !exists {
		return true, 0
	}
	
	// 清理过期的记录
	windowStart := now.Add(-AccountLimitWindow)
	validCreations := make([]time.Time, 0)
	for _, t := range userCreations {
		if t.After(windowStart) {
			validCreations = append(validCreations, t)
		}
	}
	
	// 更新记录
	rl.accountCreations[userID] = validCreations
	
	// 检查是否达到限制
	if len(validCreations) >= AccountLimitMax {
		// 计算下一个可用时间
		oldest := validCreations[0]
		waitTime := oldest.Add(AccountLimitWindow).Sub(now)
		return false, waitTime
	}
	
	return true, 0
}

// RecordAccountCreation 记录用户的账户创建
func (rl *RateLimiter) RecordAccountCreation(userID string) {
	rl.accountMu.Lock()
	defer rl.accountMu.Unlock()
	
	now := time.Now()
	
	// 清理过期的记录
	windowStart := now.Add(-AccountLimitWindow)
	validCreations := make([]time.Time, 0)
	if existing, exists := rl.accountCreations[userID]; exists {
		for _, t := range existing {
			if t.After(windowStart) {
				validCreations = append(validCreations, t)
			}
		}
	}
	
	// 添加新记录
	validCreations = append(validCreations, now)
	rl.accountCreations[userID] = validCreations
}

// GetStats 获取当前速率限制状态（按提供商区分）
func (rl *RateLimiter) GetStats(provider string) (requestCount int) {
	switch provider {
	case "letsencrypt":
		return rl.getLetsEncryptStats()
	case "zerossl":
		return rl.getZeroSSLStats()
	default:
		return rl.getLetsEncryptStats()
	}
}

// getLetsEncryptStats 获取Let's Encrypt证书申请统计
func (rl *RateLimiter) getLetsEncryptStats() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	now := time.Now()
	
	// 计算有效的证书申请数量
	certStart := now.Add(-LetsEncryptCertLimitWindow)
	count := 0
	for _, t := range rl.letsEncryptRequests {
		if t.After(certStart) {
			count++
		}
	}
	
	return count
}

// getZeroSSLStats 获取ZeroSSL请求统计
func (rl *RateLimiter) getZeroSSLStats() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	now := time.Now()
	
	// 计算有效的ZeroSSL请求数量
	requestStart := now.Add(-ZeroSSLRateLimitWindow)
	count := 0
	for _, t := range rl.zeroSSLRequests {
		if t.After(requestStart) {
			count++
		}
	}
	
	return count
}

// GetAccountStats 获取用户的账户创建统计
func (rl *RateLimiter) GetAccountStats(userID string) int {
	rl.accountMu.RLock()
	defer rl.accountMu.RUnlock()
	
	now := time.Now()
	
	// 计算用户的有效账户创建数量
	userCreations, exists := rl.accountCreations[userID]
	if !exists {
		return 0
	}
	
	windowStart := now.Add(-AccountLimitWindow)
	count := 0
	for _, t := range userCreations {
		if t.After(windowStart) {
			count++
		}
	}
	
	return count
}