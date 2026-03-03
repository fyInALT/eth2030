package encrypted

import (
	"encoding/binary"
	"errors"
	"math/big"
	"sync"

	"github.com/eth2030/eth2030/crypto"
)

// VDF timer errors.
var (
	ErrVDFTimerNilPuzzle       = errors.New("vdf_timer: nil puzzle")
	ErrVDFTimerNotReady        = errors.New("vdf_timer: VDF puzzle not yet solved")
	ErrVDFTimerInvalidSolution = errors.New("vdf_timer: invalid VDF solution")
	ErrVDFTimerSlotMismatch    = errors.New("vdf_timer: solution slot mismatch")
)

// VDFPuzzle represents a slot-bound VDF puzzle that must be solved
// before encrypted transactions can be decrypted.
type VDFPuzzle struct {
	Slot       uint64
	Input      []byte
	Difficulty uint64
	Solved     bool
	Solution   []byte
}

// VDFDecryptionTimer ensures encrypted txs can only be decrypted
// after a VDF delay tied to the slot number.
type VDFDecryptionTimer struct {
	mu         sync.Mutex
	puzzles    map[uint64]*VDFPuzzle // slot -> puzzle
	difficulty uint64
}

// NewVDFDecryptionTimer creates a timer with the given difficulty.
func NewVDFDecryptionTimer(difficulty uint64) *VDFDecryptionTimer {
	if difficulty == 0 {
		difficulty = 1 << 16 // default ~65K iterations
	}
	return &VDFDecryptionTimer{
		puzzles:    make(map[uint64]*VDFPuzzle),
		difficulty: difficulty,
	}
}

// GenerateSlotPuzzle creates a VDF puzzle for a given slot.
// The puzzle input is derived deterministically from the slot number.
func (t *VDFDecryptionTimer) GenerateSlotPuzzle(slot uint64) *VDFPuzzle {
	t.mu.Lock()
	defer t.mu.Unlock()

	if puzzle, exists := t.puzzles[slot]; exists {
		return puzzle
	}

	// Derive puzzle input from slot: H("vdf-slot" || slot_bytes).
	var slotBuf [8]byte
	binary.BigEndian.PutUint64(slotBuf[:], slot)
	input := crypto.Keccak256(
		[]byte("vdf-slot-puzzle"),
		slotBuf[:],
	)

	puzzle := &VDFPuzzle{
		Slot:       slot,
		Input:      input,
		Difficulty: t.difficulty,
	}
	t.puzzles[slot] = puzzle
	return puzzle
}

// SubmitSolution submits a VDF solution for a slot puzzle.
// Returns true if the solution is valid.
func (t *VDFDecryptionTimer) SubmitSolution(slot uint64, solution []byte) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	puzzle, exists := t.puzzles[slot]
	if !exists {
		return false, ErrVDFTimerNilPuzzle
	}
	if puzzle.Solved {
		return true, nil // already solved
	}

	// Verify the solution against the puzzle difficulty.
	if !verifySolution(puzzle, solution) {
		return false, ErrVDFTimerInvalidSolution
	}

	puzzle.Solved = true
	puzzle.Solution = make([]byte, len(solution))
	copy(puzzle.Solution, solution)
	return true, nil
}

// IsSlotReady returns true if the VDF puzzle for the slot has been solved.
func (t *VDFDecryptionTimer) IsSlotReady(slot uint64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	puzzle, exists := t.puzzles[slot]
	if !exists {
		return false
	}
	return puzzle.Solved
}

// GetDecryptionKey derives a decryption key share from the VDF solution.
// Can only be called after the puzzle is solved.
func (t *VDFDecryptionTimer) GetDecryptionKey(slot uint64) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	puzzle, exists := t.puzzles[slot]
	if !exists {
		return nil, ErrVDFTimerNilPuzzle
	}
	if !puzzle.Solved {
		return nil, ErrVDFTimerNotReady
	}

	// Derive key from solution: H("vdf-decrypt-key" || slot || solution).
	var slotBuf [8]byte
	binary.BigEndian.PutUint64(slotBuf[:], slot)
	key := crypto.Keccak256(
		[]byte("vdf-decrypt-key"),
		slotBuf[:],
		puzzle.Solution,
	)
	return key, nil
}

// PruneBefore removes puzzles for slots before the given slot.
func (t *VDFDecryptionTimer) PruneBefore(slot uint64) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	pruned := 0
	for s := range t.puzzles {
		if s < slot {
			delete(t.puzzles, s)
			pruned++
		}
	}
	return pruned
}

// PuzzleCount returns the number of tracked puzzles.
func (t *VDFDecryptionTimer) PuzzleCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.puzzles)
}

// VDFTimedDecryptor wraps a ThresholdDecryptor with VDF timing enforcement.
// Decryption is only allowed after the VDF puzzle for the slot is solved.
type VDFTimedDecryptor struct {
	timer     *VDFDecryptionTimer
	decryptor *ThresholdDecryptor
	slot      uint64
}

// NewVDFTimedDecryptor creates a timed decryptor for a specific slot.
func NewVDFTimedDecryptor(timer *VDFDecryptionTimer, decryptor *ThresholdDecryptor, slot uint64) *VDFTimedDecryptor {
	return &VDFTimedDecryptor{
		timer:     timer,
		decryptor: decryptor,
		slot:      slot,
	}
}

// TryDecrypt attempts to decrypt using the threshold decryptor, but only
// if the VDF puzzle for this slot has been solved.
func (vtd *VDFTimedDecryptor) TryDecrypt() ([]byte, error) {
	if !vtd.timer.IsSlotReady(vtd.slot) {
		return nil, ErrVDFTimerNotReady
	}
	return vtd.decryptor.TryDecrypt()
}

// Slot returns the slot this decryptor is bound to.
func (vtd *VDFTimedDecryptor) Slot() uint64 {
	return vtd.slot
}

// --- Internal helpers ---

// verifySolution checks that the VDF solution is valid for the puzzle.
// The solution must produce a hash below a target derived from difficulty.
func verifySolution(puzzle *VDFPuzzle, solution []byte) bool {
	if len(solution) == 0 {
		return false
	}

	// Compute verification hash.
	var diffBuf [8]byte
	binary.BigEndian.PutUint64(diffBuf[:], puzzle.Difficulty)
	verifyHash := crypto.Keccak256(
		[]byte("vdf-verify"),
		puzzle.Input,
		solution,
		diffBuf[:],
	)

	// Check that the hash meets difficulty: interpret as big.Int < target.
	hashInt := new(big.Int).SetBytes(verifyHash)
	// Target: 2^256 / difficulty.
	maxVal := new(big.Int).Lsh(big.NewInt(1), 256)
	target := new(big.Int).Div(maxVal, new(big.Int).SetUint64(puzzle.Difficulty))
	return hashInt.Cmp(target) < 0
}

// SolveVDFPuzzle is a helper that brute-forces a VDF puzzle solution (for testing).
func SolveVDFPuzzle(puzzle *VDFPuzzle) []byte {
	for nonce := uint64(0); nonce < 1<<32; nonce++ {
		var nonceBuf [8]byte
		binary.BigEndian.PutUint64(nonceBuf[:], nonce)
		candidate := crypto.Keccak256(
			[]byte("vdf-solution-candidate"),
			puzzle.Input,
			nonceBuf[:],
		)
		if verifySolution(puzzle, candidate) {
			return candidate
		}
	}
	return nil
}
