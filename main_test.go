package main

import "testing"

func TestSummarizeAvailabilityAndBudget(t *testing.T) {
	result := Summarize(1000, 999, 0.999)
	if result.Availability != 0.999 {
		t.Fatalf("availability = %v", result.Availability)
	}
	if result.ErrorBudgetRemaining != 0 {
		t.Fatalf("remaining budget = %v", result.ErrorBudgetRemaining)
	}
	if Summarize(1000, 1000, 0.999).ErrorBudgetRemaining != 1 {
		t.Fatal("perfect availability should preserve the full budget")
	}
	miss := Summarize(100, 98, 0.999)
	if miss.ErrorBudgetRemaining != 0 {
		t.Fatalf("budget should be exhausted: %v", miss.ErrorBudgetRemaining)
	}
}

func TestValidate(t *testing.T) {
	if Validate(Config{URL: "http://localhost", TargetAvailability: 0.99}) != nil {
		t.Fatal("valid config rejected")
	}
	if Validate(Config{URL: "http://localhost", TargetAvailability: 1.2}) == nil {
		t.Fatal("invalid target accepted")
	}
}
