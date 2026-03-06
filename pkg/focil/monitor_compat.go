package focil

// monitor_compat.go re-exports types from focil/monitor for backward compatibility.

import "github.com/eth2030/eth2030/focil/monitor"

// Monitor type aliases.
type (
	MonitorConfig        = monitor.MonitorConfig
	InclusionItem        = monitor.InclusionItem
	SlotComplianceReport = monitor.SlotComplianceReport
	InclusionMonitor     = monitor.InclusionMonitor
	ComplianceMetrics    = monitor.ComplianceMetrics
	MempoolMonitor       = monitor.MempoolMonitor
	FairnessAnalyzer     = monitor.FairnessAnalyzer
	CensorshipIndicator  = monitor.CensorshipIndicator
)

// Monitor error variables.
var (
	ErrMonitorSlotNotFound   = monitor.ErrMonitorSlotNotFound
	ErrMonitorNoBuilders     = monitor.ErrMonitorNoBuilders
	ErrMonitorBuilderUnknown = monitor.ErrMonitorBuilderUnknown
)

// Monitor function wrappers.
func DefaultMonitorConfig() MonitorConfig { return monitor.DefaultMonitorConfig() }
func NewInclusionMonitor(config *MonitorConfig) *InclusionMonitor {
	return monitor.NewInclusionMonitor(config)
}
func NewMempoolMonitor() *MempoolMonitor     { return monitor.NewMempoolMonitor() }
func NewFairnessAnalyzer() *FairnessAnalyzer { return monitor.NewFairnessAnalyzer() }
func NewCensorshipIndicator(threshold float64) *CensorshipIndicator {
	return monitor.NewCensorshipIndicator(threshold)
}
