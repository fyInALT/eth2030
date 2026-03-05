package eftest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/eth2030/eth2030/geth"
)

// testdataDir returns the path to go-ethereum's state test fixtures.
func testdataDir() string {
	return filepath.Join("..", "..", "..", "refs", "go-ethereum", "tests", "testdata", "GeneralStateTests")
}

func TestGethRunnerSmoke(t *testing.T) {
	// Run a single known fixture to verify the geth runner works.
	fixturePath := filepath.Join(testdataDir(), "stExample", "add11.json")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("go-ethereum test fixtures not available")
	}

	tests, err := LoadGethTests(fixturePath)
	if err != nil {
		t.Fatalf("LoadGethTests: %v", err)
	}

	var passed, failed, skipped int
	for _, test := range tests {
		for _, sub := range test.Subtests() {
			if !geth.EFTestForkSupported(sub.Fork) {
				skipped++
				continue
			}
			result := test.RunWithGeth(sub)
			if result.Passed {
				passed++
			} else {
				failed++
				if result.Error != nil {
					t.Logf("FAIL %s/%s[%d]: %v", test.Name, sub.Fork, sub.Index, result.Error)
				}
			}
		}
	}
	t.Logf("smoke: %d passed, %d failed, %d skipped", passed, failed, skipped)
	if passed == 0 {
		t.Error("expected at least one passing test")
	}
}

func TestGethCategorySummary(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full EF test suite in short mode")
	}

	dir := testdataDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("go-ethereum test fixtures not available")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	type catResult struct {
		passed  int
		failed  int
		skipped int
	}

	var (
		mu         sync.Mutex
		catMap     = make(map[string]*catResult)
		totPassed  int
		totFailed  int
		totSkipped int
	)

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		catMap[name] = &catResult{}
		names = append(names, name)
	}
	sort.Strings(names)

	// Run each category directory in parallel goroutines so totals are
	// available before we print the summary (t.Run+t.Parallel subtests
	// only run after the parent returns, making result collection impossible).
	var wg sync.WaitGroup
	for _, name := range names {
		name := name
		wg.Add(1)
		go func() {
			defer wg.Done()
			catDir := filepath.Join(dir, name)
			files, _ := filepath.Glob(filepath.Join(catDir, "*.json"))
			var passed, failed, skipped int
			for _, file := range files {
				tests, loadErr := LoadGethTests(file)
				if loadErr != nil {
					continue
				}
				for _, test := range tests {
					for _, sub := range test.Subtests() {
						if !geth.EFTestForkSupported(sub.Fork) {
							skipped++
							continue
						}
						func() {
							defer func() {
								if r := recover(); r != nil {
									failed++
								}
							}()
							result := test.RunWithGeth(sub)
							if result.Passed {
								passed++
							} else {
								failed++
							}
						}()
					}
				}
			}
			mu.Lock()
			catMap[name].passed = passed
			catMap[name].failed = failed
			catMap[name].skipped = skipped
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Collect totals.
	mu.Lock()
	for _, n := range names {
		cat := catMap[n]
		totPassed += cat.passed
		totFailed += cat.failed
		totSkipped += cat.skipped
	}
	mu.Unlock()

	t.Log("")
	t.Log("=== GETH-BACKED EF STATE TEST RESULTS ===")
	t.Log(strings.Repeat("-", 70))
	t.Logf("%-40s %6s %6s %6s %6s", "CATEGORY", "PASS", "FAIL", "SKIP", "RATE")
	t.Log(strings.Repeat("-", 70))

	for _, n := range names {
		cat := catMap[n]
		total := cat.passed + cat.failed
		rate := 0.0
		if total > 0 {
			rate = float64(cat.passed) / float64(total) * 100
		}
		t.Logf("%-40s %6d %6d %6d %5.1f%%", n, cat.passed, cat.failed, cat.skipped, rate)
	}

	t.Log(strings.Repeat("-", 70))
	total := totPassed + totFailed
	rate := 0.0
	if total > 0 {
		rate = float64(totPassed) / float64(total) * 100
	}
	t.Logf("%-40s %6d %6d %6d %5.1f%%", "TOTAL", totPassed, totFailed, totSkipped, rate)
	t.Log("")
	t.Logf("SUMMARY: %d/%d passing (%.1f%%)", totPassed, total, rate)

	if totPassed == 0 {
		t.Error("expected at least some passing tests")
	}
}

func TestGethRunnerSingleCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// Test a single category for quick validation.
	catDir := filepath.Join(testdataDir(), "stExample")
	if _, err := os.Stat(catDir); os.IsNotExist(err) {
		t.Skip("test fixtures not available")
	}

	files, _ := filepath.Glob(filepath.Join(catDir, "*.json"))
	var passed, failed int
	for _, file := range files {
		tests, err := LoadGethTests(file)
		if err != nil {
			continue
		}
		for _, test := range tests {
			for _, sub := range test.Subtests() {
				if !geth.EFTestForkSupported(sub.Fork) {
					continue
				}
				result := test.RunWithGeth(sub)
				if result.Passed {
					passed++
				} else {
					failed++
					t.Logf("FAIL %s/%s[%d]: %v", test.Name, sub.Fork, sub.Index, result.Error)
				}
			}
		}
	}
	t.Logf("stExample: %d passed, %d failed", passed, failed)
}

func TestGethForkConfigs(t *testing.T) {
	forks := []string{
		"Frontier", "Homestead", "EIP150", "EIP158",
		"Byzantium", "Constantinople", "Istanbul",
		"Berlin", "London", "Merge", "Shanghai", "Cancun", "Prague",
	}
	for _, fork := range forks {
		config, err := geth.EFTestChainConfig(fork)
		if err != nil {
			t.Errorf("EFTestChainConfig(%s): %v", fork, err)
			continue
		}
		if config == nil {
			t.Errorf("EFTestChainConfig(%s) returned nil", fork)
		}
	}

	_, err := geth.EFTestChainConfig("UnknownFork")
	if err == nil {
		t.Error("expected error for unsupported fork")
	}
}

func TestGethMakePreState(t *testing.T) {
	accounts := map[string]geth.PreAccount{
		"0x1000000000000000000000000000000000000001": {
			Balance: hexToBigInt("0x1000"),
			Nonce:   1,
			Code:    []byte{0x60, 0x00},
		},
	}

	state, err := geth.MakePreState(accounts)
	if err != nil {
		t.Fatalf("MakePreState: %v", err)
	}
	defer state.Close()

	if state.StateDB == nil {
		t.Fatal("StateDB is nil")
	}

	// Verify the account exists.
	addr := gethcommon.HexToAddress("0x1000000000000000000000000000000000000001")
	if !state.StateDB.Exist(addr) {
		t.Error("account does not exist")
	}
	if state.StateDB.GetNonce(addr) != 1 {
		t.Errorf("nonce = %d, want 1", state.StateDB.GetNonce(addr))
	}
}

var _ = fmt.Sprintf // keep fmt imported
