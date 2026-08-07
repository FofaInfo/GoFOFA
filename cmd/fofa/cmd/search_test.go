package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestPipelineProcessReturnsQueryError(t *testing.T) {
	oldWorkers := workers
	oldRatePerSecond := ratePerSecond
	oldTemplate := template
	t.Cleanup(func() {
		workers = oldWorkers
		ratePerSecond = oldRatePerSecond
		template = oldTemplate
	})

	workers = 1
	ratePerSecond = 100
	template = "{}"
	wantErr := errors.New("query failed")
	err := pipelineProcess(func(string) error {
		return wantErr
	}, strings.NewReader("port=80\n"))
	if !errors.Is(err, wantErr) {
		t.Fatalf("pipeline error = %v, want %v", err, wantErr)
	}
}
