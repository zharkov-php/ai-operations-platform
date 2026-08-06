package apikey

import "testing"

func TestValidateScopes(t *testing.T) {
	if err := Validate("ingestion", []string{"ingest:calls"}); err != nil {
		t.Fatal(err)
	}
	if err := Validate("bad", []string{"root:everything"}); err == nil {
		t.Fatal("unknown scope accepted")
	}
	if err := Validate("duplicate", []string{"ingest:calls", "ingest:calls"}); err == nil {
		t.Fatal("duplicate scope accepted")
	}
}
