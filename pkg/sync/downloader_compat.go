package sync

// downloader_compat.go re-exports types from sync/downloader for backward compatibility.

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/sync/downloader"
)

// Downloader type aliases.
type (
	HeaderData            = downloader.HeaderData
	BlockData             = downloader.BlockData
	HeaderSource          = downloader.HeaderSource
	BodySource            = downloader.BodySource
	PeerID                = downloader.PeerID
	FetchRequest          = downloader.FetchRequest
	FetchResponse         = downloader.FetchResponse
	HeaderFetcher         = downloader.HeaderFetcher
	BodyFetcher           = downloader.BodyFetcher
	BlockDownloaderConfig = downloader.BlockDownloaderConfig
	DownloadTask          = downloader.DownloadTask
	BlockDownloader       = downloader.BlockDownloader
	BDLPeerInfo           = downloader.BDLPeerInfo
	BDLConfig             = downloader.BDLConfig
	BDLProgress           = downloader.BDLProgress
	BodyDownloader        = downloader.BodyDownloader
	HDLPeerInfo           = downloader.HDLPeerInfo
	HDLConfig             = downloader.HDLConfig
	HDLProgress           = downloader.HDLProgress
	HeaderDownloader      = downloader.HeaderDownloader
	DownloadConfig        = downloader.DownloadConfig
	PeerInfo              = downloader.PeerInfo
	DownloadProgress      = downloader.DownloadProgress
	ChainDownloader       = downloader.ChainDownloader
	SkeletonConfig        = downloader.SkeletonConfig
	SkeletonAnchor        = downloader.SkeletonAnchor
	GapSegment            = downloader.GapSegment
	ReceiptTask           = downloader.ReceiptTask
	ThrottleState         = downloader.ThrottleState
	SkeletonChain         = downloader.SkeletonChain
	BlkAnnouncePeer       = downloader.BlkAnnouncePeer
	BlkAnnounceMetrics    = downloader.BlkAnnounceMetrics
	BlkAnnounce           = downloader.BlkAnnounce
)

// Downloader error variables.
var (
	ErrBDLClosed         = downloader.ErrBDLClosed
	ErrBDLRunning        = downloader.ErrBDLRunning
	ErrBDLNoPeers        = downloader.ErrBDLNoPeers
	ErrBDLTxRootMismatch = downloader.ErrBDLTxRootMismatch
	ErrBDLWdRootMismatch = downloader.ErrBDLWdRootMismatch
	ErrBDLBodyMissing    = downloader.ErrBDLBodyMissing
	ErrBDLRetryExhausted = downloader.ErrBDLRetryExhausted
)

// Downloader function wrappers.
func NewHeaderFetcher(batchSize int) *HeaderFetcher { return downloader.NewHeaderFetcher(batchSize) }
func NewBodyFetcher(batchSize int) *BodyFetcher     { return downloader.NewBodyFetcher(batchSize) }
func DefaultBDLConfig() BDLConfig                   { return downloader.DefaultBDLConfig() }
func NewBodyDownloader(cfg BDLConfig, fb downloader.BodySource) *BodyDownloader {
	return downloader.NewBodyDownloader(cfg, fb)
}
func DefaultHDLConfig() HDLConfig { return downloader.DefaultHDLConfig() }
func NewHeaderDownloader(cfg HDLConfig, src downloader.HeaderSource) *HeaderDownloader {
	return downloader.NewHeaderDownloader(cfg, src)
}
func DefaultDownloadConfig() *DownloadConfig { return downloader.DefaultDownloadConfig() }
func NewChainDownloader(cfg *DownloadConfig) *ChainDownloader {
	return downloader.NewChainDownloader(cfg)
}
func DefaultSkeletonConfig() SkeletonConfig { return downloader.DefaultSkeletonConfig() }
func NewSkeletonChain(config SkeletonConfig) *SkeletonChain {
	return downloader.NewSkeletonChain(config)
}
func NewBlkAnnounce(hasBlock func(hash types.Hash) bool) *BlkAnnounce {
	return downloader.NewBlkAnnounce(hasBlock)
}
