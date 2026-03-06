package prover

import (
	"fmt"
	"testing"
)

func BenchmarkValidationFrameCircuit(b *testing.B) {
	for _, count := range []int{1, 10, 50, 100} {
		frames := make([][]byte, count)
		for i := range frames {
			frames[i] = []byte{byte(i%255 + 1), 0xAA, 0xBB, 0xCC}
		}
		b.Run(fmt.Sprintf("frames=%d", count), func(b *testing.B) {
			prover := NewSTARKValidationFrameProver()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = prover.ProveAllValidationFrames(frames)
			}
		})
	}
}
