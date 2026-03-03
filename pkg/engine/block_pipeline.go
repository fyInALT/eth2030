// block_pipeline.go orchestrates the block building pipeline:
// anonymous ingress -> encrypted pool -> Big FOCIL -> dependency partition -> parallel build -> merge -> propose.
package engine

import (
	"errors"
	"sync"
	"time"
)

// Pipeline errors.
var (
	ErrPipelineNotStarted = errors.New("pipeline: not started")
	ErrPipelineStopped    = errors.New("pipeline: stopped")
)

// PipelineStage identifies a stage in the block building pipeline.
type PipelineStage uint8

const (
	StageIngress   PipelineStage = iota // Anonymous transport ingress
	StageEncrypt                        // Encrypted mempool commit
	StageFOCIL                          // Big FOCIL inclusion list
	StagePartition                      // Dependency graph partitioning
	StageBuild                          // Parallel sub-block building
	StageMerge                          // Merge sub-blocks
	StagePropose                        // Propose final block
)

// String returns the name of the pipeline stage.
func (s PipelineStage) String() string {
	switch s {
	case StageIngress:
		return "ingress"
	case StageEncrypt:
		return "encrypt"
	case StageFOCIL:
		return "focil"
	case StagePartition:
		return "partition"
	case StageBuild:
		return "build"
	case StageMerge:
		return "merge"
	case StagePropose:
		return "propose"
	default:
		return "unknown"
	}
}

// PipelineConfig configures which pipeline stages are enabled.
type PipelineConfig struct {
	EnableAnonymousIngress bool
	EnableEncryptedPool    bool
	EnableBigFOCIL         bool
	EnableDependencyGraph  bool
	EnableParallelBuild    bool
}

// DefaultPipelineConfig returns a config with all stages enabled.
func DefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		EnableAnonymousIngress: true,
		EnableEncryptedPool:    true,
		EnableBigFOCIL:         true,
		EnableDependencyGraph:  true,
		EnableParallelBuild:    true,
	}
}

// StageMetrics holds metrics for a single pipeline stage.
type StageMetrics struct {
	Stage       PipelineStage
	Executions  uint64
	TotalTimeNs uint64
	TxProcessed uint64
	Errors      uint64
	LastRun     time.Time
}

// AvgLatency returns the average latency per execution in nanoseconds.
func (m *StageMetrics) AvgLatency() uint64 {
	if m.Executions == 0 {
		return 0
	}
	return m.TotalTimeNs / m.Executions
}

// PipelineResult holds the result of a pipeline execution for a slot.
type PipelineResult struct {
	Slot          uint64
	TxCount       int
	GroupCount    int
	StageResults  map[PipelineStage]StageResult
	TotalDuration time.Duration
	Success       bool
	Error         string
}

// StageResult holds the result of a single stage execution.
type StageResult struct {
	Stage    PipelineStage
	Duration time.Duration
	TxCount  int
	Skipped  bool
	Error    string
}

// BlockPipeline orchestrates the full block building pipeline.
type BlockPipeline struct {
	mu      sync.RWMutex
	config  *PipelineConfig
	metrics map[PipelineStage]*StageMetrics
	running bool
}

// NewBlockPipeline creates a new block building pipeline.
func NewBlockPipeline(config *PipelineConfig) *BlockPipeline {
	if config == nil {
		config = DefaultPipelineConfig()
	}

	metrics := make(map[PipelineStage]*StageMetrics)
	for _, stage := range []PipelineStage{
		StageIngress, StageEncrypt, StageFOCIL,
		StagePartition, StageBuild, StageMerge, StagePropose,
	} {
		metrics[stage] = &StageMetrics{Stage: stage}
	}

	return &BlockPipeline{
		config:  config,
		metrics: metrics,
	}
}

// Start initializes the pipeline.
func (bp *BlockPipeline) Start() error {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.running = true
	return nil
}

// Stop shuts down the pipeline.
func (bp *BlockPipeline) Stop() error {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.running = false
	return nil
}

// ProcessSlot runs the pipeline for a given slot.
// Each stage is executed in order, with disabled stages skipped.
func (bp *BlockPipeline) ProcessSlot(slot uint64) *PipelineResult {
	bp.mu.RLock()
	if !bp.running {
		bp.mu.RUnlock()
		return &PipelineResult{
			Slot:    slot,
			Success: false,
			Error:   ErrPipelineNotStarted.Error(),
		}
	}
	bp.mu.RUnlock()

	start := time.Now()
	result := &PipelineResult{
		Slot:         slot,
		StageResults: make(map[PipelineStage]StageResult),
	}

	stages := []struct {
		stage   PipelineStage
		enabled bool
	}{
		{StageIngress, bp.config.EnableAnonymousIngress},
		{StageEncrypt, bp.config.EnableEncryptedPool},
		{StageFOCIL, bp.config.EnableBigFOCIL},
		{StagePartition, bp.config.EnableDependencyGraph},
		{StageBuild, bp.config.EnableParallelBuild},
		{StageMerge, true},
		{StagePropose, true},
	}

	totalTxs := 0
	groups := 1

	for _, s := range stages {
		stageStart := time.Now()
		sr := StageResult{Stage: s.stage}

		if !s.enabled {
			sr.Skipped = true
			result.StageResults[s.stage] = sr
			continue
		}

		// Each stage processes in sequence.
		// In production, each stage would call into the actual subsystem.
		sr.TxCount = totalTxs
		sr.Duration = time.Since(stageStart)
		result.StageResults[s.stage] = sr

		bp.mu.Lock()
		if m := bp.metrics[s.stage]; m != nil {
			m.Executions++
			m.TotalTimeNs += uint64(sr.Duration.Nanoseconds())
			m.TxProcessed += uint64(sr.TxCount)
			m.LastRun = time.Now()
			if sr.Error != "" {
				m.Errors++
			}
		}
		bp.mu.Unlock()
	}

	result.TxCount = totalTxs
	result.GroupCount = groups
	result.TotalDuration = time.Since(start)
	result.Success = true
	return result
}

// GetMetrics returns a copy of the metrics for all stages.
func (bp *BlockPipeline) GetMetrics() map[PipelineStage]*StageMetrics {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	result := make(map[PipelineStage]*StageMetrics, len(bp.metrics))
	for stage, m := range bp.metrics {
		cp := *m
		result[stage] = &cp
	}
	return result
}

// IsRunning returns true if the pipeline is active.
func (bp *BlockPipeline) IsRunning() bool {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.running
}

// StageEnabled returns true if the given stage is enabled in the config.
func (bp *BlockPipeline) StageEnabled(stage PipelineStage) bool {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	switch stage {
	case StageIngress:
		return bp.config.EnableAnonymousIngress
	case StageEncrypt:
		return bp.config.EnableEncryptedPool
	case StageFOCIL:
		return bp.config.EnableBigFOCIL
	case StagePartition:
		return bp.config.EnableDependencyGraph
	case StageBuild:
		return bp.config.EnableParallelBuild
	default:
		return true
	}
}
