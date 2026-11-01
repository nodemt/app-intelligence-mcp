package astro

import "testing"

func TestKeywordPerformanceArgumentsWithAppID(t *testing.T) {
	arguments := keywordPerformanceArguments(
		"photo editor", "es", "iphone", "123456789",
	)

	value, exists := arguments["appId"]
	if !exists {
		t.Fatal("appId is missing")
	}

	if value != "123456789" {
		t.Errorf("appId = %v, want 123456789", value)
	}

	if _, exists := arguments["appID"]; exists {
		t.Error("unexpected appID key; Astro expects appId")
	}
}

func TestKeywordPerformanceArgumentsWithoutAppID(t *testing.T) {
	arguments := keywordPerformanceArguments(
		"photo editor", "es", "iphone", "",
	)

	if _, exists := arguments["appId"]; exists {
		t.Error("appId must be omitted when empty")
	}
}
