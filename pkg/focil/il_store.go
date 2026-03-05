package focil

import (
	"bytes"
	"sync"
)

// ILStore maintains per-validator-per-slot inclusion list state and detects
// equivocations per EIP-7805 §equivocation.
type ILStore struct {
	mu sync.Mutex
	// store: validatorIdx -> slotNum -> first IL received
	store map[uint64]map[uint64]*InclusionList
	// equivocators: validatorIdx -> set of slots where they equivocated
	equivocators map[uint64]map[uint64]bool
}

// NewILStore creates a new ILStore.
func NewILStore() *ILStore {
	return &ILStore{
		store:        make(map[uint64]map[uint64]*InclusionList),
		equivocators: make(map[uint64]map[uint64]bool),
	}
}

// AddIL adds an inclusion list from a validator for a slot. Returns true if
// accepted, false if dropped (equivocator's subsequent ILs are silently dropped).
// If the validator already submitted a different IL for the same slot, marks them
// as an equivocator.
func (s *ILStore) AddIL(validatorIdx, slot uint64, il *InclusionList) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Drop ILs from known equivocators.
	if slots, ok := s.equivocators[validatorIdx]; ok {
		if slots[slot] {
			return false
		}
	}

	if _, ok := s.store[validatorIdx]; !ok {
		s.store[validatorIdx] = make(map[uint64]*InclusionList)
	}

	existing, exists := s.store[validatorIdx][slot]
	if !exists {
		s.store[validatorIdx][slot] = il
		return true
	}

	// Check if the new IL differs from the existing one.
	if !ilsEqual(existing, il) {
		// Mark as equivocator.
		if _, ok := s.equivocators[validatorIdx]; !ok {
			s.equivocators[validatorIdx] = make(map[uint64]bool)
		}
		s.equivocators[validatorIdx][slot] = true
	}
	return true
}

// IsEquivocator returns true if the validator equivocated for the given slot.
func (s *ILStore) IsEquivocator(validatorIdx, slot uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if slots, ok := s.equivocators[validatorIdx]; ok {
		return slots[slot]
	}
	return false
}

// EquivocatorCount returns the number of equivocators for a given slot.
func (s *ILStore) EquivocatorCount(slot uint64) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, slots := range s.equivocators {
		if slots[slot] {
			count++
		}
	}
	return count
}

// ilsEqual compares two ILs for equality by comparing their entries.
func ilsEqual(a, b *InclusionList) bool {
	if a.Slot != b.Slot || a.ProposerIndex != b.ProposerIndex {
		return false
	}
	if len(a.Entries) != len(b.Entries) {
		return false
	}
	for i := range a.Entries {
		if !bytes.Equal(a.Entries[i].Transaction, b.Entries[i].Transaction) {
			return false
		}
	}
	return true
}
