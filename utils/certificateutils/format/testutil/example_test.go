package testutil_test

import (
	"testing"
	
	"autoSSL/utils/certificateutils/format/testutil"
)

func TestExampleUsage(t *testing.T) {
	// 示例1: 监控单个函数的内存使用
	t.Run("MonitorFunction", func(t *testing.T) {
		testutil.MonitorFunction(t, "GenerateKey", func() {
			key, err := testutil.SafeGenerateKey(1024)
			if err != nil {
				t.Fatal(err)
			}
			_ = key // 使用key
		}, 1024*1024) // 限制1MB内存增长
	})

	// 示例2: 串行运行测试
	t.Run("SerialTests", func(t *testing.T) {
		tests := map[string]func(*testing.T){
			"Test1": func(t *testing.T) {
				// 测试1代码
				data := testutil.GenerateTestData(512)
				_ = data
			},
			"Test2": func(t *testing.T) {
				// 测试2代码
				key, err := testutil.SafeGenerateKey(1024)
				if err != nil {
					t.Fatal(err)
				}
				_ = key
			},
		}
		testutil.RunSerialTests(t, tests)
	})

	// 示例3: 直接检查内存
	t.Run("DirectMemoryCheck", func(t *testing.T) {
		monitor := testutil.NewMemoryMonitor()
		
		// 执行需要监控的代码
		for i := 0; i < 100; i++ {
			data := testutil.GenerateTestData(128)
			_ = data
		}
		
		monitor.CheckMemory(t, 512*1024) // 限制512KB内存增长
	})
}