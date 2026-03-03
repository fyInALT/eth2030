package engine

import (
	"testing"
)

func TestNewBlockPipeline(t *testing.T) {
	bp := NewBlockPipeline(nil)
	if bp == nil {
		t.Fatal("expected non-nil pipeline")
	}
	if bp.IsRunning() {
		t.Error("new pipeline should not be running")
	}

	metrics := bp.GetMetrics()
	if len(metrics) != 7 {
		t.Errorf("expected 7 stage metrics, got %d", len(metrics))
	}
}

func TestNewBlockPipeline_CustomConfig(t *testing.T) {
	cfg := &PipelineConfig{
		EnableAnonymousIngress: false,
		EnableEncryptedPool:    true,
		EnableBigFOCIL:         false,
		EnableDependencyGraph:  true,
		EnableParallelBuild:    false,
	}
	bp := NewBlockPipeline(cfg)
	if bp == nil {
		t.Fatal("expected non-nil pipeline")
	}

	if bp.StageEnabled(StageIngress) {
		t.Error("ingress should be disabled")
	}
	if !bp.StageEnabled(StageEncrypt) {
		t.Error("encrypt should be enabled")
	}
	if bp.StageEnabled(StageFOCIL) {
		t.Error("focil should be disabled")
	}
	if !bp.StageEnabled(StagePartition) {
		t.Error("partition should be enabled")
	}
	if bp.StageEnabled(StageBuild) {
		t.Error("build should be disabled")
	}
	// Merge and propose are always enabled.
	if !bp.StageEnabled(StageMerge) {
		t.Error("merge should always be enabled")
	}
	if !bp.StageEnabled(StagePropose) {
		t.Error("propose should always be enabled")
	}
}

func TestBlockPipeline_StartStop(t *testing.T) {
	bp := NewBlockPipeline(nil)

	if bp.IsRunning() {
		t.Error("should not be running before start")
	}

	if err := bp.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if !bp.IsRunning() {
		t.Error("should be running after start")
	}

	if err := bp.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if bp.IsRunning() {
		t.Error("should not be running after stop")
	}
}

func TestBlockPipeline_ProcessSlot(t *testing.T) {
	bp := NewBlockPipeline(nil)
	_ = bp.Start()

	result := bp.ProcessSlot(42)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.Slot != 42 {
		t.Errorf("expected slot 42, got %d", result.Slot)
	}

	// All 7 stages should have results.
	if len(result.StageResults) != 7 {
		t.Errorf("expected 7 stage results, got %d", len(result.StageResults))
	}

	// No stages should be skipped with default config.
	for stage, sr := range result.StageResults {
		if sr.Skipped {
			t.Errorf("stage %s should not be skipped", stage)
		}
	}
}

func TestBlockPipeline_ProcessSlot_NotStarted(t *testing.T) {
	bp := NewBlockPipeline(nil)

	result := bp.ProcessSlot(1)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Success {
		t.Error("expected failure when not started")
	}
	if result.Error == "" {
		t.Error("expected error message")
	}
}

func TestBlockPipeline_ProcessSlot_DisabledStages(t *testing.T) {
	cfg := &PipelineConfig{
		EnableAnonymousIngress: false,
		EnableEncryptedPool:    false,
		EnableBigFOCIL:         false,
		EnableDependencyGraph:  false,
		EnableParallelBuild:    false,
	}
	bp := NewBlockPipeline(cfg)
	_ = bp.Start()

	result := bp.ProcessSlot(1)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}

	// Check disabled stages are skipped.
	skippedStages := []PipelineStage{StageIngress, StageEncrypt, StageFOCIL, StagePartition, StageBuild}
	for _, stage := range skippedStages {
		sr, ok := result.StageResults[stage]
		if !ok {
			t.Errorf("missing result for stage %s", stage)
			continue
		}
		if !sr.Skipped {
			t.Errorf("stage %s should be skipped", stage)
		}
	}

	// Merge and propose should not be skipped.
	for _, stage := range []PipelineStage{StageMerge, StagePropose} {
		sr, ok := result.StageResults[stage]
		if !ok {
			t.Errorf("missing result for stage %s", stage)
			continue
		}
		if sr.Skipped {
			t.Errorf("stage %s should not be skipped", stage)
		}
	}
}

func TestBlockPipeline_GetMetrics(t *testing.T) {
	bp := NewBlockPipeline(nil)
	_ = bp.Start()

	// Process multiple slots.
	bp.ProcessSlot(1)
	bp.ProcessSlot(2)
	bp.ProcessSlot(3)

	metrics := bp.GetMetrics()
	if len(metrics) != 7 {
		t.Errorf("expected 7 metrics, got %d", len(metrics))
	}

	// Each enabled stage should have 3 executions.
	for stage, m := range metrics {
		if m.Executions != 3 {
			t.Errorf("stage %s: expected 3 executions, got %d", stage, m.Executions)
		}
	}
}

func TestPipelineStage_String(t *testing.T) {
	tests := []struct {
		stage PipelineStage
		want  string
	}{
		{StageIngress, "ingress"},
		{StageEncrypt, "encrypt"},
		{StageFOCIL, "focil"},
		{StagePartition, "partition"},
		{StageBuild, "build"},
		{StageMerge, "merge"},
		{StagePropose, "propose"},
		{PipelineStage(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.stage.String(); got != tt.want {
			t.Errorf("PipelineStage(%d).String() = %q, want %q", tt.stage, got, tt.want)
		}
	}
}

func TestStageMetrics_AvgLatency(t *testing.T) {
	m := &StageMetrics{
		Executions:  0,
		TotalTimeNs: 1000,
	}
	if m.AvgLatency() != 0 {
		t.Error("expected 0 latency with 0 executions")
	}

	m.Executions = 4
	m.TotalTimeNs = 1000
	if got := m.AvgLatency(); got != 250 {
		t.Errorf("expected avg latency 250, got %d", got)
	}
}

func TestBlockPipeline_StageEnabled(t *testing.T) {
	bp := NewBlockPipeline(DefaultPipelineConfig())

	allStages := []PipelineStage{
		StageIngress, StageEncrypt, StageFOCIL,
		StagePartition, StageBuild, StageMerge, StagePropose,
	}
	for _, s := range allStages {
		if !bp.StageEnabled(s) {
			t.Errorf("stage %s should be enabled with default config", s)
		}
	}
}
