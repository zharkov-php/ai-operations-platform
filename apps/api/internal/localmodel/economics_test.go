package localmodel

import (
	"errors"
	"testing"
)

func testConfig() Configuration {
	return Configuration{HardwareName: "Illustrative GPU workstation", PurchaseCost: "3600", UsefulLifetimeMonths: 36, MonthlyElectricity: "40", MonthlyMaintenance: "60", AvailableMemoryGB: "48", EstimatedRequestsPerSecond: "2", Utilization: "0.25", SupportedModel: "illustrative-local-8b", ContextLimit: 32768, Currency: "USD", BenchmarkSource: "user estimate; not measured"}
}

func TestCompareAmortizationCapacityAndBreakEven(t *testing.T) {
	result, err := Compare(testConfig(), ComparisonInput{MonthlyRequests: 100000, HostedCostPerRequest: "0.01", RequiredMemoryGB: "16", RequiredContextTokens: 8000})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Supported || result.HardwareAmortization != "100.000000000000" || result.LocalMonthlyCost != "200.000000000000" || result.MonthlyCapacity != "1296000" || result.EstimatedMonthlySavings != "800.000000000000" || *result.LocalCostPerRequest != "0.002000000000" || *result.EstimatedBreakEvenMonths != "4.0000" {
		t.Fatalf("result=%+v", result)
	}
}

func TestCompareRejectsZeroVolumeWithoutDivision(t *testing.T) {
	result, err := Compare(testConfig(), ComparisonInput{MonthlyRequests: 0, HostedCostPerRequest: "0.01", RequiredMemoryGB: "16", RequiredContextTokens: 8000})
	if err != nil {
		t.Fatal(err)
	}
	if result.Supported || result.LocalCostPerRequest != nil || result.EstimatedBreakEvenMonths != nil || result.Constraints[0] != "monthly_volume_is_zero" {
		t.Fatalf("result=%+v", result)
	}
}

func TestCompareReportsUnsupportedConstraints(t *testing.T) {
	result, err := Compare(testConfig(), ComparisonInput{MonthlyRequests: 2000000, HostedCostPerRequest: "0.01", RequiredMemoryGB: "80", RequiredContextTokens: 50000})
	if err != nil {
		t.Fatal(err)
	}
	if result.Supported || len(result.Constraints) != 3 {
		t.Fatalf("result=%+v", result)
	}
}

func TestCompareRejectsInvalidUtilization(t *testing.T) {
	config := testConfig()
	config.Utilization = "1.1"
	if _, err := Compare(config, ComparisonInput{}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err=%v", err)
	}
}
