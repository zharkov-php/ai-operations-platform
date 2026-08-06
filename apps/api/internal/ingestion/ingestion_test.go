package ingestion

import (
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	valid := Input{ProjectID: "project", WorkloadID: "workload", ExternalCallID: "call-1", Provider: "demo", Model: "model", RequestTimestamp: time.Now().UTC(), Status: "success"}
	if err := Validate(valid); err != nil {
		t.Fatal(err)
	}
	valid.InputTokens = -1
	if err := Validate(valid); err == nil {
		t.Fatal("negative tokens accepted")
	}
}
func TestValidateTimestampOrder(t *testing.T) {
	request := time.Now().UTC()
	response := request.Add(-time.Second)
	input := Input{ProjectID: "project", WorkloadID: "workload", ExternalCallID: "call-1", Provider: "demo", Model: "model", RequestTimestamp: request, ResponseTimestamp: &response, Status: "success"}
	if err := Validate(input); err == nil {
		t.Fatal("response before request accepted")
	}
}
