package budget

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

var (
	ErrInvalidInput      = errors.New("invalid budget input")
	ErrNotFound          = errors.New("budget alert not found")
	ErrInvalidTransition = errors.New("invalid alert transition")
)

type Snapshot struct {
	ProjectID, Currency                                    string
	MonthlyBudget, MonthCost, TodayCost, PriorDailyAverage decimal.Decimal
	Now                                                    time.Time
}
type Candidate struct {
	Type, Severity, DedupeKey string
	ThresholdValue            decimal.Decimal
	Evidence                  map[string]any
}
type ThresholdRule struct {
	Percentage decimal.Decimal
	Severity   string
}

func Detect(s Snapshot, thresholds []ThresholdRule) []Candidate {
	if !s.MonthlyBudget.IsPositive() {
		return nil
	}
	now := s.Now.UTC()
	days := decimal.NewFromInt(int64(now.Day()))
	daysInMonth := decimal.NewFromInt(int64(time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()))
	projection := s.MonthCost.Div(days).Mul(daysInMonth)
	items := []Candidate{}
	for _, threshold := range thresholds {
		percentage := threshold.Percentage
		trigger := s.MonthlyBudget.Mul(percentage).Div(decimal.NewFromInt(100))
		if s.MonthCost.GreaterThanOrEqual(trigger) {
			items = append(items, Candidate{"percentage", threshold.Severity, "percentage:" + percentage.StringFixed(2), percentage, map[string]any{"observed_cost": s.MonthCost.StringFixed(12), "budget": s.MonthlyBudget.StringFixed(12), "currency": s.Currency}})
		}
	}
	if projection.GreaterThan(s.MonthlyBudget) {
		items = append(items, Candidate{"projected_overspend", "warning", "projected_overspend:" + now.Format("2006-01"), projection, map[string]any{"projection": projection.StringFixed(12), "budget": s.MonthlyBudget.StringFixed(12), "method": "elapsed_day_linear"}})
	}
	if s.PriorDailyAverage.IsPositive() && s.TodayCost.GreaterThan(s.PriorDailyAverage.Mul(decimal.NewFromInt(2))) {
		items = append(items, Candidate{"anomaly", "warning", "daily_anomaly:" + now.Format("2006-01-02"), s.TodayCost, map[string]any{"today_cost": s.TodayCost.StringFixed(12), "prior_daily_average": s.PriorDailyAverage.StringFixed(12), "rule": "greater_than_2x_prior_daily_average"}})
	}
	return items
}
func ValidateThreshold(percentage string) (decimal.Decimal, error) {
	value, err := decimal.NewFromString(percentage)
	if err != nil || !value.IsPositive() || value.GreaterThan(decimal.NewFromInt(200)) {
		return decimal.Zero, ErrInvalidInput
	}
	return value, nil
}
func DedupeSummary(created, suppressed int) string {
	return fmt.Sprintf("created=%d suppressed=%d", created, suppressed)
}
