package portfolio

import "testing"

func TestValidateProject(t *testing.T) {
	tests := []struct {
		name  string
		input ProjectInput
		valid bool
	}{{"valid", ProjectInput{Name: "API", Slug: "api", Environment: "production", MonthlyBudget: "1200.25", Currency: "USD"}, true}, {"bad slug", ProjectInput{Name: "API", Slug: "Bad Slug", Environment: "production", MonthlyBudget: "1", Currency: "USD"}, false}, {"negative budget", ProjectInput{Name: "API", Slug: "api", Environment: "production", MonthlyBudget: "-0.01", Currency: "USD"}, false}, {"binary float syntax", ProjectInput{Name: "API", Slug: "api", Environment: "production", MonthlyBudget: "NaN", Currency: "USD"}, false}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateProject(test.input)
			if (err == nil) != test.valid {
				t.Fatalf("error=%v valid=%v", err, test.valid)
			}
		})
	}
}
func TestValidateWorkload(t *testing.T) {
	valid := WorkloadInput{ProjectID: "project", Name: "Classifier", Slug: "classifier", Type: "feature", Owner: "Support", Criticality: "medium", QualityRequirement: "standard", PrivacyClassification: "internal"}
	if err := ValidateWorkload(valid); err != nil {
		t.Fatal(err)
	}
	valid.Type = "unknown"
	if err := ValidateWorkload(valid); err == nil {
		t.Fatal("invalid type accepted")
	}
}
func TestNormalizePagination(t *testing.T) {
	opts := normalize(ListOptions{Limit: 1000, Offset: -2})
	if opts.Limit != 100 || opts.Offset != 0 {
		t.Fatalf("options=%+v", opts)
	}
	defaults := normalize(ListOptions{})
	if defaults.Limit != 25 {
		t.Fatalf("defaults=%+v", defaults)
	}
}
