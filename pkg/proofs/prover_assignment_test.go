package proofs

import (
	"sync"
	"testing"
)

func TestProverPoolRegisterAndAssign(t *testing.T) {
	pool := NewProverPool(3)
	for i := range 5 {
		id := proverID(i)
		if err := pool.RegisterProver(id, 10); err != nil {
			t.Fatalf("RegisterProver(%s): %v", id, err)
		}
	}
	if pool.Size() != 5 {
		t.Fatalf("expected 5 provers, got %d", pool.Size())
	}

	result, err := pool.AssignProvers(100, 3)
	if err != nil {
		t.Fatalf("AssignProvers: %v", err)
	}
	if len(result.ProverIDs) != 3 {
		t.Fatalf("expected 3 assigned, got %d", len(result.ProverIDs))
	}
	if result.BlockNum != 100 {
		t.Fatalf("expected block 100, got %d", result.BlockNum)
	}
}

func TestProverPoolReputationIncreasesOnSuccess(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProver("p1", 10); err != nil {
		t.Fatal(err)
	}

	before, _ := pool.GetReputation("p1")
	if err := pool.RecordSuccess("p1", 1); err != nil {
		t.Fatal(err)
	}
	after, _ := pool.GetReputation("p1")
	if after <= before {
		t.Fatalf("reputation should increase: before=%f, after=%f", before, after)
	}
}

func TestProverPoolReputationDecreasesOnFailure(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProver("p1", 10); err != nil {
		t.Fatal(err)
	}

	before, _ := pool.GetReputation("p1")
	if err := pool.RecordFailure("p1", 1); err != nil {
		t.Fatal(err)
	}
	after, _ := pool.GetReputation("p1")
	if after >= before {
		t.Fatalf("reputation should decrease: before=%f, after=%f", before, after)
	}
}

func TestProverPoolInsufficientProvers(t *testing.T) {
	pool := NewProverPool(5)
	if err := pool.RegisterProver("p1", 10); err != nil {
		t.Fatal(err)
	}

	_, err := pool.AssignProvers(1, 3)
	if err == nil {
		t.Fatal("should fail with insufficient provers")
	}
}

func TestProverPoolCapacityLimits(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProver("p1", 2); err != nil {
		t.Fatal(err)
	}
	if err := pool.RegisterProver("p2", 10); err != nil {
		t.Fatal(err)
	}

	// Assign p1 twice to fill its capacity.
	if _, err := pool.AssignProvers(1, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.AssignProvers(2, 1); err != nil {
		t.Fatal(err)
	}

	// After 2 assignments, p1 is at capacity. Next assignment should still
	// succeed by picking p2.
	result, err := pool.AssignProvers(3, 1)
	if err != nil {
		t.Fatalf("should succeed with available prover: %v", err)
	}
	if len(result.ProverIDs) != 1 {
		t.Fatal("expected 1 assigned")
	}
}

func TestProverPoolTimeDecay(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProver("p1", 10); err != nil {
		t.Fatal(err)
	}

	// Boost reputation.
	for range 5 {
		if err := pool.RecordSuccess("p1", 1); err != nil {
			t.Fatal(err)
		}
	}

	before, _ := pool.GetReputation("p1")
	if err := pool.DecayAllReputations(0.5); err != nil {
		t.Fatal(err)
	}
	after, _ := pool.GetReputation("p1")
	if after >= before {
		t.Fatalf("decay should reduce score: before=%f, after=%f", before, after)
	}
}

func TestProverPoolByzantineProverDeprioritized(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProver("good", 10); err != nil {
		t.Fatal(err)
	}
	if err := pool.RegisterProver("bad", 10); err != nil {
		t.Fatal(err)
	}

	// Good prover succeeds.
	for range 5 {
		if err := pool.RecordSuccess("good", 1); err != nil {
			t.Fatal(err)
		}
	}
	// Bad prover fails repeatedly.
	for range 10 {
		if err := pool.RecordFailure("bad", 1); err != nil {
			t.Fatal(err)
		}
	}

	// Assign 1: should pick 'good' first.
	result, err := pool.AssignProvers(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.ProverIDs[0] != "good" {
		t.Fatalf("expected 'good' prover, got %s", result.ProverIDs[0])
	}
}

func TestProverPoolReassignmentWhenAtCapacity(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProver("p1", 1); err != nil {
		t.Fatal(err)
	}
	if err := pool.RegisterProver("p2", 1); err != nil {
		t.Fatal(err)
	}

	// Assign p1 (it will be picked first if same reputation).
	_, err := pool.AssignProvers(1, 1)
	if err != nil {
		t.Fatal(err)
	}

	// Now only p2 is available.
	result, err := pool.AssignProvers(2, 1)
	if err != nil {
		t.Fatal(err)
	}
	// The second prover should be whichever one was not assigned first.
	if result.ProverIDs[0] == "" {
		t.Fatal("should have assigned a prover")
	}
}

func TestProverPoolMultipleBlocksSameProvers(t *testing.T) {
	pool := NewProverPool(1)
	for i := range 3 {
		if err := pool.RegisterProver(proverID(i), 100); err != nil {
			t.Fatal(err)
		}
	}

	for block := uint64(1); block <= 10; block++ {
		result, err := pool.AssignProvers(block, 2)
		if err != nil {
			t.Fatalf("block %d: %v", block, err)
		}
		if len(result.ProverIDs) != 2 {
			t.Fatalf("block %d: expected 2, got %d", block, len(result.ProverIDs))
		}
	}
}

func TestProverPoolGeographicDiversity(t *testing.T) {
	pool := NewProverPool(1)
	// Register provers in different regions.
	if err := pool.RegisterProverWithRegion("us1", 10, "us-east"); err != nil {
		t.Fatal(err)
	}
	if err := pool.RegisterProverWithRegion("eu1", 10, "eu-west"); err != nil {
		t.Fatal(err)
	}
	if err := pool.RegisterProverWithRegion("ap1", 10, "ap-southeast"); err != nil {
		t.Fatal(err)
	}
	// Give them all the same reputation boost.
	for _, id := range []string{"us1", "eu1", "ap1"} {
		for range 3 {
			if err := pool.RecordSuccess(id, 1); err != nil {
				t.Fatal(err)
			}
		}
	}

	result, err := pool.AssignProvers(1, 3)
	if err != nil {
		t.Fatal(err)
	}
	// Verify all 3 regions are represented.
	regions := make(map[string]bool)
	for _, r := range result.Regions {
		regions[r] = true
	}
	if len(regions) != 3 {
		t.Fatalf("expected 3 different regions, got %d: %v", len(regions), regions)
	}
}

func TestProverPoolZeroCapacityRejected(t *testing.T) {
	pool := NewProverPool(1)
	err := pool.RegisterProver("p1", 0)
	if err != ErrZeroCapacity {
		t.Fatalf("expected ErrZeroCapacity, got %v", err)
	}
	err = pool.RegisterProver("p2", -1)
	if err != ErrZeroCapacity {
		t.Fatalf("expected ErrZeroCapacity for negative, got %v", err)
	}
}

func TestProverPoolReputationFloor(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProver("p1", 10); err != nil {
		t.Fatal(err)
	}

	// Fail many times to push reputation to floor.
	for range 100 {
		if err := pool.RecordFailure("p1", 1); err != nil {
			t.Fatal(err)
		}
	}

	rep, _ := pool.GetReputation("p1")
	if rep < minReputation {
		t.Fatalf("reputation should not drop below %f, got %f", minReputation, rep)
	}
	if rep != minReputation {
		t.Fatalf("reputation should be at floor %f, got %f", minReputation, rep)
	}
}

func TestProverPoolDuplicateRegistration(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProver("p1", 5); err != nil {
		t.Fatal(err)
	}
	err := pool.RegisterProver("p1", 10)
	if err != ErrProverAlreadyExists {
		t.Fatalf("expected ErrProverAlreadyExists, got %v", err)
	}
}

func TestProverPoolExactMinProvers(t *testing.T) {
	pool := NewProverPool(3)
	for i := range 3 {
		if err := pool.RegisterProver(proverID(i), 10); err != nil {
			t.Fatal(err)
		}
	}

	result, err := pool.AssignProvers(1, 3)
	if err != nil {
		t.Fatalf("should succeed with exactly min provers: %v", err)
	}
	if len(result.ProverIDs) != 3 {
		t.Fatalf("expected 3, got %d", len(result.ProverIDs))
	}
}

func TestProverPoolConcurrentAssignment(t *testing.T) {
	pool := NewProverPool(1)
	for i := range 20 {
		if err := pool.RegisterProver(proverID(i), 100); err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup
	errs := make(chan error, 50)

	for block := uint64(0); block < 50; block++ {
		wg.Add(1)
		go func(bn uint64) {
			defer wg.Done()
			_, err := pool.AssignProvers(bn, 3)
			if err != nil {
				errs <- err
			}
		}(block)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent assignment error: %v", err)
	}
}

func TestProverPoolEmptyAssign(t *testing.T) {
	pool := NewProverPool(1)
	_, err := pool.AssignProvers(1, 1)
	if err != ErrPoolEmpty {
		t.Fatalf("expected ErrPoolEmpty, got %v", err)
	}
}

func TestProverPoolInvalidCountAssign(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProver("p1", 10); err != nil {
		t.Fatal(err)
	}
	_, err := pool.AssignProvers(1, 0)
	if err != ErrInvalidCount {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}
	_, err = pool.AssignProvers(1, -1)
	if err != ErrInvalidCount {
		t.Fatalf("expected ErrInvalidCount for negative, got %v", err)
	}
}

func TestReputationScorerDecayInvalid(t *testing.T) {
	rs := NewReputationScorer()
	rs.Init("p1")
	if err := rs.Decay(0); err != ErrInvalidDecayFactor {
		t.Fatalf("expected ErrInvalidDecayFactor for 0, got %v", err)
	}
	if err := rs.Decay(-0.5); err != ErrInvalidDecayFactor {
		t.Fatalf("expected ErrInvalidDecayFactor for negative, got %v", err)
	}
	if err := rs.Decay(1.5); err != ErrInvalidDecayFactor {
		t.Fatalf("expected ErrInvalidDecayFactor for >1, got %v", err)
	}
	// 1.0 should be valid (no change).
	if err := rs.Decay(1.0); err != nil {
		t.Fatalf("decay(1.0) should succeed: %v", err)
	}
}

func TestReputationScorerAttempts(t *testing.T) {
	rs := NewReputationScorer()
	rs.Init("p1")
	if rs.Attempts("p1") != 0 {
		t.Fatal("initial attempts should be 0")
	}
	rs.RecordSuccess("p1")
	rs.RecordFailure("p1")
	if rs.Attempts("p1") != 2 {
		t.Fatalf("expected 2 attempts, got %d", rs.Attempts("p1"))
	}
}

func TestProverPoolRecordUnregistered(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RecordSuccess("nonexistent", 1); err != ErrProverNotRegistered {
		t.Fatalf("expected ErrProverNotRegistered, got %v", err)
	}
	if err := pool.RecordFailure("nonexistent", 1); err != ErrProverNotRegistered {
		t.Fatalf("expected ErrProverNotRegistered, got %v", err)
	}
}

func TestProverPoolGetProver(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProverWithRegion("p1", 5, "us-east"); err != nil {
		t.Fatal(err)
	}
	p, err := pool.GetProver("p1")
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "p1" || p.Capacity != 5 || p.Region != "us-east" {
		t.Fatalf("unexpected prover info: %+v", p)
	}

	_, err = pool.GetProver("nonexistent")
	if err != ErrProverNotRegistered {
		t.Fatalf("expected ErrProverNotRegistered, got %v", err)
	}
}

func TestProverPoolReputationCap(t *testing.T) {
	pool := NewProverPool(1)
	if err := pool.RegisterProver("p1", 10); err != nil {
		t.Fatal(err)
	}
	// Many successes to hit the cap.
	for range 100 {
		if err := pool.RecordSuccess("p1", 1); err != nil {
			t.Fatal(err)
		}
	}
	rep, _ := pool.GetReputation("p1")
	if rep > maxReputation {
		t.Fatalf("reputation should not exceed %f, got %f", maxReputation, rep)
	}
}

// proverID generates a deterministic prover ID string for testing.
func proverID(i int) string {
	return "prover-" + string(rune('A'+i))
}
