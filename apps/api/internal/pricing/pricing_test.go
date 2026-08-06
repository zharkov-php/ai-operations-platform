package pricing

import (
	"testing"
	"time"
)

func TestCalculateCostExactly(t *testing.T) {
	cached := "1"
	price := Price{InputPrice: "5", OutputPrice: "15", CachedInputPrice: &cached, UnitSize: 1000000}
	cost, err := Calculate(price, Usage{InputTokens: 1000, OutputTokens: 200, CachedInputTokens: 500})
	if err != nil {
		t.Fatal(err)
	}
	if cost.Total.String() != "0.0085" {
		t.Fatalf("total=%s", cost.Total)
	}
	if cost.Input.String() != "0.005" || cost.Output.String() != "0.003" || cost.CachedInput.String() != "0.0005" {
		t.Fatalf("cost=%+v", cost)
	}
}
func TestCalculatePreservesDecimalPrecision(t *testing.T) {
	price := Price{InputPrice: "0.000001", OutputPrice: "0", UnitSize: 1000000}
	cost, err := Calculate(price, Usage{InputTokens: 1})
	if err != nil {
		t.Fatal(err)
	}
	if cost.Total.String() != "0.000000000001" {
		t.Fatalf("total=%s", cost.Total)
	}
}
func TestCalculateRejectsUnavailableCachedPricing(t *testing.T) {
	_, err := Calculate(Price{InputPrice: "1", OutputPrice: "1", UnitSize: 1000}, Usage{CachedInputTokens: 1})
	if err == nil {
		t.Fatal("expected cached pricing error")
	}
}
func TestValidatePricing(t *testing.T) {
	from := time.Now().UTC()
	to := from.Add(time.Hour)
	valid := PriceInput{Provider: "demo", Model: "small", InputPrice: "1.25", OutputPrice: "2.5", Currency: "USD", UnitSize: 1000000, EffectiveFrom: from, EffectiveTo: &to, SourceNote: "Illustrative"}
	if err := Validate(valid); err != nil {
		t.Fatal(err)
	}
	valid.InputPrice = "NaN"
	if err := Validate(valid); err == nil {
		t.Fatal("invalid decimal accepted")
	}
	valid.InputPrice = "1"
	valid.Currency = "123"
	if err := Validate(valid); err == nil {
		t.Fatal("invalid currency accepted")
	}
}
