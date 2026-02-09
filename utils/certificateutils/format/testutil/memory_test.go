// Package testutil 提供测试工具函数，帮助避免内存和并发问题
package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"runtime"
	"testing"
	"time"
)

// MemoryMonitor 监控内存使用情况
type MemoryMonitor struct {
	before runtime.MemStats
	after  runtime.MemStats
}

// NewMemoryMonitor 创建新的内存监控器
func NewMemoryMonitor() *MemoryMonitor {
	var m MemoryMonitor
	runtime.GC()
	runtime.ReadMemStats(&m.before)
	return &m
}

// CheckMemory 检查内存使用情况
func (m *MemoryMonitor) CheckMemory(t *testing.T, maxIncrease int64) {
	runtime.GC()
	runtime.ReadMemStats(&m.after)
	
	increase := int64(m.after.Alloc - m.before.Alloc)
	if increase > maxIncrease {
		t.Errorf("内存使用过高: 增加了 %d bytes (最大允许: %d bytes)", increase, maxIncrease)
	}
}

// RunSerialTests 串行运行测试用例
func RunSerialTests(t *testing.T, tests map[string]func(*testing.T)) {
	for name, testFunc := range tests {
		t.Run(name, testFunc)
		
		// 每个测试后强制GC，避免累积
		runtime.GC()
		time.Sleep(10 * time.Millisecond) // 给GC一点时间
	}
}

// SafeGenerateKey 生成小密钥用于测试
func SafeGenerateKey(bits int) (*rsa.PrivateKey, error) {
	// 在测试中使用较小的密钥以减少内存使用
	if bits > 1024 {
		bits = 1024
	}
	return rsa.GenerateKey(rand.Reader, bits)
}

// GenerateTestData 生成小测试数据集
func GenerateTestData(size int) []byte {
	if size > 1024 {
		size = 1024
	}
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// MemorySnapshot 获取当前内存快照
func MemorySnapshot() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("Alloc=%dMiB TotalAlloc=%dMiB Sys=%dMiB NumGC=%d",
		m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

// MonitorFunction 监控函数内存使用
func MonitorFunction(t *testing.T, name string, fn func(), maxMemory int64) {
	monitor := NewMemoryMonitor()
	fn()
	monitor.CheckMemory(t, maxMemory)
	fmt.Printf("%s 内存使用: %s\n", name, MemorySnapshot())
}

// TestMemoryUsage 测试内存使用的示例
func TestMemoryUsage(t *testing.T) {
	monitor := NewMemoryMonitor()
	
	// 这里放置要测试的代码
	data := GenerateTestData(512)
	_ = data
	
	monitor.CheckMemory(t, 1024*1024) // 最多1MB
	fmt.Printf("测试完成，内存快照: %s\n", MemorySnapshot())
}