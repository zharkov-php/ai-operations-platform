package budget

import (
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestDetectThresholdProjectionAndAnomaly(t *testing.T) {
	snapshot := Snapshot{ProjectID: "p", Currency: "USD", MonthlyBudget: decimal.NewFromInt(1000), MonthCost: decimal.NewFromInt(600), TodayCost: decimal.NewFromInt(30), PriorDailyAverage: decimal.NewFromInt(10), Now: time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)}
	items := Detect(snapshot, []ThresholdRule{{decimal.NewFromInt(50), "info"}, {decimal.NewFromInt(75), "critical"}})
	if len(items) != 3 || items[0].Type != "percentage" || items[1].Type != "projected_overspend" || items[2].Type != "anomaly" {
		t.Fatalf("items=%+v", items)
	}
}
func TestDetectSkipsProjectsWithoutBudget(t *testing.T) {
	if items := Detect(Snapshot{Now: time.Now()}, nil); len(items) != 0 {
		t.Fatalf("items=%+v", items)
	}
}
func TestThresholdValidation(t *testing.T) {
	for _, value := range []string{"0", "-1", "201", "bad"} {
		if _, err := ValidateThreshold(value); err == nil {
			t.Fatalf("value=%s", value)
		}
	}
}
