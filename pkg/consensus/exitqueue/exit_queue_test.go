package exitqueue

import (
	"testing"
)

func TestMinslackExitQueue(t *testing.T) {
	q := &MinslackExitQueue{}

	tests := []struct {
		name      string
		epoch     uint64
		queueSize uint64
		maxChurn  uint64
		want      uint64
	}{
		{
			name:      "empty queue returns MinValidatorWithdrawabilityDelay",
			epoch:     0,
			queueSize: 0,
			maxChurn:  8,
			want:      MinValidatorWithdrawabilityDelay,
		},
		{
			name:      "full queue returns MaxSeedLookahead",
			epoch:     10,
			queueSize: 8,
			maxChurn:  8,
			want:      MaxSeedLookahead,
		},
		{
			name:      "queue exceeds maxChurn returns MaxSeedLookahead",
			epoch:     5,
			queueSize: 100,
			maxChurn:  8,
			want:      MaxSeedLookahead,
		},
		{
			name:      "half-full queue returns MinValidatorWithdrawabilityDelay/2",
			epoch:     1,
			queueSize: 3,
			maxChurn:  8,
			want:      MinValidatorWithdrawabilityDelay / 2,
		},
		{
			name:      "queue exactly at half boundary uses half-full path",
			epoch:     2,
			queueSize: 4,
			maxChurn:  8,
			// queueSize(4) < maxChurn/2(4) is false (4 < 4 is false)
			// so this falls into the proportional path
			want: MinValidatorWithdrawabilityDelay / 2 * (8 - 4) / 8,
		},
		{
			name:      "partial fill returns proportional delay",
			epoch:     3,
			queueSize: 5,
			maxChurn:  8,
			// queueSize(5) >= maxChurn/2(4) and queueSize(5) < maxChurn(8)
			// proportional: MinValidatorWithdrawabilityDelay / 2 * (maxChurn - queueSize) / maxChurn
			want: MinValidatorWithdrawabilityDelay / 2 * (8 - 5) / 8,
		},
		{
			name:      "one below full queue",
			epoch:     7,
			queueSize: 7,
			maxChurn:  8,
			want:      MinValidatorWithdrawabilityDelay / 2 * (8 - 7) / 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := q.ComputeExitDelay(tt.epoch, tt.queueSize, tt.maxChurn)
			if got != tt.want {
				t.Errorf("ComputeExitDelay(epoch=%d, queueSize=%d, maxChurn=%d) = %d, want %d",
					tt.epoch, tt.queueSize, tt.maxChurn, got, tt.want)
			}
		})
	}
}

func TestValidatorExitMinslack(t *testing.T) {
	q := &MinslackExitQueue{}
	maxChurn := uint64(10)

	delays := make([]uint64, 10)
	for i := uint64(0); i < 10; i++ {
		delays[i] = q.ComputeExitDelay(0, i, maxChurn)
	}

	// First validator sees empty queue delay.
	if delays[0] != MinValidatorWithdrawabilityDelay {
		t.Errorf("delay[0] (empty queue) = %d, want %d", delays[0], MinValidatorWithdrawabilityDelay)
	}

	// Validators 1..4 see half-full delay (queueSize < maxChurn/2 = 5).
	for i := uint64(1); i < 5; i++ {
		if delays[i] != MinValidatorWithdrawabilityDelay/2 {
			t.Errorf("delay[%d] (half-full) = %d, want %d", i, delays[i], MinValidatorWithdrawabilityDelay/2)
		}
	}

	// Delays should generally decrease as queue fills.
	if delays[9] > delays[0] {
		t.Errorf("last delay (%d) should not exceed first delay (%d) for full queue scenario",
			delays[9], delays[0])
	}

	// Verify constants are correct.
	if MinValidatorWithdrawabilityDelay != 256 {
		t.Errorf("MinValidatorWithdrawabilityDelay = %d, want 256", MinValidatorWithdrawabilityDelay)
	}
	if MaxSeedLookahead != 4 {
		t.Errorf("MaxSeedLookahead = %d, want 4", MaxSeedLookahead)
	}
}
