package bal

import (
	"bytes"
	"errors"
	"fmt"
)

// ErrBALSizeExceeded is returned when the BAL item count exceeds the
// block-gas-limit / ITEM_COST budget per EIP-7928.
var ErrBALSizeExceeded = errors.New("bal: item count exceeds block gas limit / ITEM_COST")

// BALItemCost is the cost per BAL item in gas units (EIP-7928 §constants).
const BALItemCost uint64 = 2000

// ValidateBALOrdering verifies that AccessEntries in the BAL are in strict
// ascending lexicographic order by (Address, AccessIndex) per EIP-7928 §ordering.
// Returns an error describing the first violation found.
func ValidateBALOrdering(bal *BlockAccessList) error {
	for i := 1; i < len(bal.Entries); i++ {
		prev := &bal.Entries[i-1]
		curr := &bal.Entries[i]
		cmp := bytes.Compare(prev.Address[:], curr.Address[:])
		if cmp > 0 {
			return fmt.Errorf("bal ordering: entry %d address %x > entry %d address %x",
				i-1, prev.Address, i, curr.Address)
		}
		if cmp == 0 && prev.AccessIndex >= curr.AccessIndex {
			return fmt.Errorf("bal ordering: duplicate (addr=%x, accessIndex=%d) at entries %d and %d",
				curr.Address, curr.AccessIndex, i-1, i)
		}
	}
	return nil
}
