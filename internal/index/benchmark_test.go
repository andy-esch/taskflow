package index

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	taskflowv1 "github.com/andy-esch/taskflow/contracts/gen/go/contracts/proto/taskflow/v1"
)

func generateDummyIndex(count int) *Index {
	tasks := make([]*taskflowv1.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = &taskflowv1.Task{
			Id:       fmt.Sprintf("task-%d", i),
			Title:    fmt.Sprintf("Task %d", i),
			Status:   taskflowv1.Status_STATUS_READY_TO_START,
			Tier:     2,
			Priority: taskflowv1.Priority_PRIORITY_HIGH,
			Effort:   "2-4 hours",
			Tags:     []string{"backend", "frontend", "infrastructure"},
			Project:  "TaskFlow V1",
		}
	}
	return &Index{
		Version:     "1.0",
		LastUpdated: time.Now().Format(time.RFC3339),
		Tasks:       tasks,
	}
}

func BenchmarkIndexLoad(b *testing.B) {
	// 1. Setup: Create a large JSON blob
	idx := generateDummyIndex(10000)
	data, err := json.Marshal(idx)
	if err != nil {
		b.Fatalf("Failed to marshal: %v", err)
	}

	b.Logf("JSON Size: %.2f MB", float64(len(data))/(1024*1024))

	b.ResetTimer()

	// 2. Measure: Unmarshal repeatedly
	for i := 0; i < b.N; i++ {
		var loaded Index
		if err := json.Unmarshal(data, &loaded); err != nil {
			b.Fatalf("Failed to unmarshal: %v", err)
		}
	}
}
