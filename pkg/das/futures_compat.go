package das

// futures_compat.go re-exports types from das/futures for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/das/futures"
)

// Futures type aliases.
type (
	FutureType         = futures.FutureType
	FutureStatus       = futures.FutureStatus
	BlobFutureContract = futures.BlobFutureContract
	BlobFuturesMarket  = futures.BlobFuturesMarket
	BlobFuture         = futures.BlobFuture
	FuturesMarket      = futures.FuturesMarket
	OrderSide          = futures.OrderSide
	OrderStatus        = futures.OrderStatus
	FuturesOrder       = futures.FuturesOrder
	MarginAccount      = futures.MarginAccount
	FuturesOrderBook   = futures.FuturesOrderBook
	FuturesPool        = futures.FuturesPool
)

// Futures constants.
const (
	ShortDatedFuture       = futures.ShortDatedFuture
	LongDatedFuture        = futures.LongDatedFuture
	ShortDatedMaxSlots     = futures.ShortDatedMaxSlots
	LongDatedMaxSlots      = futures.LongDatedMaxSlots
	FutureActive           = futures.FutureActive
	FutureSettled          = futures.FutureSettled
	FutureExpired          = futures.FutureExpired
	FutureCancelled        = futures.FutureCancelled
	MinMarginBasisPoints   = futures.MinMarginBasisPoints
	LiquidationBasisPoints = futures.LiquidationBasisPoints
	BasisPointsDenom       = futures.BasisPointsDenom
	MaxOrdersPerBook       = futures.MaxOrdersPerBook
)

// Futures error variables.
var (
	ErrBlobFutureNotFound    = futures.ErrBlobFutureNotFound
	ErrBlobFutureNotActive   = futures.ErrBlobFutureNotActive
	ErrBlobFutureInvalidSlot = futures.ErrBlobFutureInvalidSlot
	ErrBlobFutureBadExpiry   = futures.ErrBlobFutureBadExpiry
	ErrBlobFutureBadPrice    = futures.ErrBlobFutureBadPrice
	ErrBlobFutureBadIndex    = futures.ErrBlobFutureBadIndex
	ErrBlobFutureDuplicate   = futures.ErrBlobFutureDuplicate
	ErrFutureNotFound        = futures.ErrFutureNotFound
	ErrFutureExpired         = futures.ErrFutureExpired
	ErrFutureSettled         = futures.ErrFutureSettled
	ErrInvalidExpiry         = futures.ErrInvalidExpiry
	ErrInvalidPrice          = futures.ErrInvalidPrice
	ErrOrderBookEmpty        = futures.ErrOrderBookEmpty
	ErrOrderNotFound         = futures.ErrOrderNotFound
	ErrOrderAlreadyFilled    = futures.ErrOrderAlreadyFilled
	ErrInsufficientMargin    = futures.ErrInsufficientMargin
	ErrMarginBelowMin        = futures.ErrMarginBelowMin
	ErrMarginAccountExists   = futures.ErrMarginAccountExists
	ErrMarginAccountMissing  = futures.ErrMarginAccountMissing
	ErrPoolSlotRange         = futures.ErrPoolSlotRange
	ErrPoolNoLiquidity       = futures.ErrPoolNoLiquidity
)

// Futures function wrappers.
func NewBlobFuturesMarket(currentSlot uint64) *BlobFuturesMarket {
	return futures.NewBlobFuturesMarket(currentSlot)
}
func ComputeSettlementPrice(committed, actual types.Hash, price *big.Int) *big.Int {
	return futures.ComputeSettlementPrice(committed, actual, price)
}
func ValidateFutureContract(f *BlobFutureContract, currentSlot uint64) error {
	return futures.ValidateFutureContract(f, currentSlot)
}
func NewFuturesMarket(currentSlot uint64) *FuturesMarket {
	return futures.NewFuturesMarket(currentSlot)
}
func PriceFuture(currentSlot, expirySlot uint64, blobCount uint64) *big.Int {
	return futures.PriceFuture(currentSlot, expirySlot, blobCount)
}
func NewFuturesOrderBook(targetSlot uint64) *FuturesOrderBook {
	return futures.NewFuturesOrderBook(targetSlot)
}
func NewFuturesPool(startSlot, endSlot uint64) (*FuturesPool, error) {
	return futures.NewFuturesPool(startSlot, endSlot)
}
