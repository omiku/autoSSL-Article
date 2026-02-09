package config

import (
	"testing"
)

type TestConfig struct {
	Domains []string `json:"domains"`
	ZoneId  string   `json:"zoneId"`
	Count   int      `json:"count"`
}

func TestSliceConversion(t *testing.T) {
	// 测试 []interface{} 到 []string 的转换
	config := map[string]interface{}{
		"domains": []interface{}{"example.com", "*.example.com", "test.com"},
		"zoneId":  "zone-123456",
		"count":   5,
	}

	var testConfig TestConfig
	err := PopulateConfig(config, &testConfig)
	if err != nil {
		t.Fatalf("PopulateConfig failed: %v", err)
	}

	// 验证结果
	if len(testConfig.Domains) != 3 {
		t.Errorf("Expected 3 domains, got %d", len(testConfig.Domains))
	}

	expectedDomains := []string{"example.com", "*.example.com", "test.com"}
	for i, domain := range testConfig.Domains {
		if domain != expectedDomains[i] {
			t.Errorf("Expected domain[%d] = %s, got %s", i, expectedDomains[i], domain)
		}
	}

	if testConfig.ZoneId != "zone-123456" {
		t.Errorf("Expected zoneId = zone-123456, got %s", testConfig.ZoneId)
	}

	if testConfig.Count != 5 {
		t.Errorf("Expected count = 5, got %d", testConfig.Count)
	}
}

func TestEmptySlice(t *testing.T) {
	// 测试空切片
	config := map[string]interface{}{
		"domains": []interface{}{},
		"zoneId":  "zone-123456",
	}

	var testConfig TestConfig
	err := PopulateConfig(config, &testConfig)
	if err != nil {
		t.Fatalf("PopulateConfig failed: %v", err)
	}

	if len(testConfig.Domains) != 0 {
		t.Errorf("Expected empty domains, got %v", testConfig.Domains)
	}
}

func TestStringSliceDirect(t *testing.T) {
	// 测试直接的 []string 类型
	config := map[string]interface{}{
		"domains": []string{"example.com", "test.com"},
		"zoneId":  "zone-123456",
	}

	var testConfig TestConfig
	err := PopulateConfig(config, &testConfig)
	if err != nil {
		t.Fatalf("PopulateConfig failed: %v", err)
	}

	if len(testConfig.Domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(testConfig.Domains))
	}
}