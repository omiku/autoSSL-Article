package client

import (
	"testing"
	"time"
)

// 测试用常量
const (
	testAccountLimitMax    = 3
	testAccountLimitWindow = 24 * time.Hour
	testCertLimitWindow    = 3 * time.Hour
)

func TestRateLimiter(t *testing.T) {
	// 测试证书申请限制
	rl := NewRateLimiter()
	userID := "test-user"

	// 测试允许Let's Encrypt证书申请
	allowed, waitTime := rl.AllowCertRequest("letsencrypt")
	if !allowed {
		t.Errorf("Expected to allow cert request, got wait time: %v", waitTime)
	}
	
	// 记录Let's Encrypt证书申请
	rl.RecordCertRequest("letsencrypt")
	
	// 获取Let's Encrypt统计信息
	certCount := rl.GetStats("letsencrypt")
	if certCount != 1 {
		t.Errorf("Expected 1 cert request, got %d", certCount)
	}

	// 测试允许ZeroSSL请求
	allowed, waitTime = rl.AllowCertRequest("zerossl")
	if !allowed {
		t.Errorf("Expected to allow ZeroSSL request, got wait time: %v", waitTime)
	}
	
	// 记录ZeroSSL请求
	rl.RecordCertRequest("zerossl")
	
	// 获取ZeroSSL统计信息
	zeroSSLCount := rl.GetStats("zerossl")
	if zeroSSLCount != 1 {
		t.Errorf("Expected 1 ZeroSSL request, got %d", zeroSSLCount)
	}

	// 测试用户ACME账户创建限制
	for i := 0; i < testAccountLimitMax; i++ {
		if allowed, _ := rl.AllowAccountCreation(userID); !allowed {
			t.Errorf("第%d次账户创建应该被允许", i+1)
		}
		rl.RecordAccountCreation(userID)
	}

	// 第4次应该被拒绝
	if allowed, _ := rl.AllowAccountCreation(userID); allowed {
		t.Error("第4次账户创建应该被拒绝")
	}

	// 测试不同用户互不影响
	otherUser := "other-user"
	if allowed, _ := rl.AllowAccountCreation(otherUser); !allowed {
		t.Error("其他用户应该允许创建账户")
	}
}

func TestRateLimiterWindow(t *testing.T) {
	// 测试时间窗口清理
	rl := NewRateLimiter()
	
	// 记录一个过去的Let's Encrypt请求
	oldTime := time.Now().Add(-4 * time.Hour)
	rl.letsEncryptRequests = append(rl.letsEncryptRequests, oldTime)
	
	// 应该清理掉旧记录
	allowed, _ := rl.AllowCertRequest("letsencrypt")
	if !allowed {
		t.Error("Expected to allow cert request after old records are cleaned")
	}
	
	// 获取Let's Encrypt统计信息
	certCount := rl.GetStats("letsencrypt")
	if certCount != 0 {
		t.Errorf("Expected 0 cert requests after cleanup, got %d", certCount)
	}
	
	// 记录一个过去的ZeroSSL请求
	rl.zeroSSLRequests = append(rl.zeroSSLRequests, oldTime)
	
	// 应该清理掉旧记录
	allowed, _ = rl.AllowCertRequest("zerossl")
	if !allowed {
		t.Error("Expected to allow ZeroSSL request after old records are cleaned")
	}
	
	// 获取ZeroSSL统计信息
	zeroSSLCount := rl.GetStats("zerossl")
	if zeroSSLCount != 0 {
		t.Errorf("Expected 0 ZeroSSL requests after cleanup, got %d", zeroSSLCount)
	}
}