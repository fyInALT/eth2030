package engine

import (
	"sync/atomic"

	"github.com/eth2030/eth2030/log"
	"github.com/eth2030/eth2030/metrics"
)

var priorityLog = log.Default().Module("engine/priority")

// RequestPriority defines the priority level for Engine API requests.
type RequestPriority int

const (
	// PriorityHigh is for time-sensitive operations (FCU, GetPayload).
	PriorityHigh RequestPriority = iota
	// PriorityNormal is for regular operations (NewPayload).
	PriorityNormal
	// PriorityLow is for non-critical operations (status checks).
	PriorityLow
)

// String returns a human-readable priority name.
func (p RequestPriority) String() string {
	switch p {
	case PriorityHigh:
		return "high"
	case PriorityNormal:
		return "normal"
	case PriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

// MethodPriority returns the default priority for an Engine API method.
func MethodPriority(method string) RequestPriority {
	switch method {
	case "engine_forkchoiceUpdatedV1",
		"engine_forkchoiceUpdatedV2",
		"engine_forkchoiceUpdatedV3":
		// FCU is time-sensitive - CL needs quick response.
		return PriorityHigh

	case "engine_getPayloadV1",
		"engine_getPayloadV2",
		"engine_getPayloadV3",
		"engine_getPayloadV4",
		"engine_getPayloadV5",
		"engine_getPayloadV6",
		"engine_getPayloadV7":
		// GetPayload needs to be quick to meet slot deadline.
		return PriorityHigh

	case "engine_newPayloadV1",
		"engine_newPayloadV2",
		"engine_newPayloadV3",
		"engine_newPayloadV4",
		"engine_newPayloadV5":
		// NewPayload involves execution, can take longer.
		return PriorityNormal

	case "engine_getPayloadBodiesByHashV1",
		"engine_getPayloadBodiesByHashV2",
		"engine_getPayloadBodiesByRangeV1",
		"engine_getPayloadBodiesByRangeV2":
		// Body retrieval is not time-critical.
		return PriorityLow

	default:
		return PriorityNormal
	}
}

// preemptionThreshold is the number of concurrent high-priority requests
// that triggers preemption of lower-priority work.
const preemptionThreshold = 2

// PriorityHandler tracks request priorities and provides statistics.
// It can be extended to implement actual priority queuing if needed.
type PriorityHandler struct {
	// Statistics (use atomic for lock-free updates)
	highCount   atomic.Int64
	normalCount atomic.Int64
	lowCount    atomic.Int64

	// Current active requests by priority
	highActive   atomic.Int32
	normalActive atomic.Int32
	lowActive    atomic.Int32
}

// NewPriorityHandler creates a new priority handler.
func NewPriorityHandler() *PriorityHandler {
	return &PriorityHandler{}
}

// StartRequest records the start of a request with the given priority.
func (h *PriorityHandler) StartRequest(priority RequestPriority) {
	h.recordRequest(priority, 1)
}

// EndRequest records the end of a request with the given priority.
func (h *PriorityHandler) EndRequest(priority RequestPriority) {
	h.recordRequest(priority, -1)
}

// recordRequest updates counters for a request start (+1) or end (-1).
func (h *PriorityHandler) recordRequest(priority RequestPriority, delta int32) {
	switch priority {
	case PriorityHigh:
		if delta > 0 {
			h.highCount.Add(int64(delta))
			metrics.EngineRequestHighTotal.Inc()
			h.highActive.Add(delta)
		} else {
			// Prevent negative count
			for {
				current := h.highActive.Load()
				if current <= 0 {
					break
				}
				if h.highActive.CompareAndSwap(current, current+delta) {
					break
				}
			}
		}
		metrics.EngineRequestHighActive.Set(int64(h.highActive.Load()))
	case PriorityNormal:
		if delta > 0 {
			h.normalCount.Add(int64(delta))
			metrics.EngineRequestNormalTotal.Inc()
			h.normalActive.Add(delta)
		} else {
			for {
				current := h.normalActive.Load()
				if current <= 0 {
					break
				}
				if h.normalActive.CompareAndSwap(current, current+delta) {
					break
				}
			}
		}
		metrics.EngineRequestNormalActive.Set(int64(h.normalActive.Load()))
	case PriorityLow:
		if delta > 0 {
			h.lowCount.Add(int64(delta))
			metrics.EngineRequestLowTotal.Inc()
			h.lowActive.Add(delta)
		} else {
			for {
				current := h.lowActive.Load()
				if current <= 0 {
					break
				}
				if h.lowActive.CompareAndSwap(current, current+delta) {
					break
				}
			}
		}
		metrics.EngineRequestLowActive.Set(int64(h.lowActive.Load()))
	}
}

// Stats returns current statistics.
type PriorityStats struct {
	HighTotal    int64
	NormalTotal  int64
	LowTotal     int64
	HighActive   int32
	NormalActive int32
	LowActive    int32
}

// GetStats returns current priority statistics.
func (h *PriorityHandler) GetStats() PriorityStats {
	return PriorityStats{
		HighTotal:    h.highCount.Load(),
		NormalTotal:  h.normalCount.Load(),
		LowTotal:     h.lowCount.Load(),
		HighActive:   h.highActive.Load(),
		NormalActive: h.normalActive.Load(),
		LowActive:    h.lowActive.Load(),
	}
}

// ShouldPreempt returns true if a high-priority request should preempt
// normal/low priority operations. This is triggered when there are
// preemptionThreshold or more concurrent high-priority requests,
// indicating system load that may benefit from priority scheduling.
func (h *PriorityHandler) ShouldPreempt() bool {
	return h.highActive.Load() >= preemptionThreshold
}