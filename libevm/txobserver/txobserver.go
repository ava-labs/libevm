// Copyright (C) 2019-2026, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package txobserver lets an embedder watch a transaction's life inside this
// node -- arrival over RPC, the mempool verdict, promotion, eviction -- without
// libevm knowing anything about where the record goes.
//
// Why a registered hook and not a direct call: avalanchego depends on libevm,
// so libevm cannot import avalanchego's observer without an import cycle. The
// embedder sets Record at init; until it does, every call here is a nil check
// and nothing else.
//
// This is instrumentation for a node under study. It is off unless someone
// registers a sink, and it costs one nil comparison when off.
package txobserver

import "github.com/ava-labs/libevm/common"

// Sources. Kept as plain strings so libevm does not need to agree with the
// embedder on an enum.
const (
	// SrcAPIIn is a transaction arriving through this node's own RPC.
	SrcAPIIn = "api_in"
	// SrcPoolAdd is a transaction offered to the mempool. The verdict -- ok, or
	// the rejection reason -- rides in note.
	SrcPoolAdd = "pool_add"
	// SrcPoolPromote is a transaction becoming executable, moving from the
	// queued set to pending.
	SrcPoolPromote = "pool_promote"
	// SrcPoolOut is a transaction leaving the mempool, with the reason in note.
	// Paired with SrcPoolAdd by hash, it gives residency -- which is why no
	// timestamp is kept here: the sink already timestamps, and holding entry
	// times inside the pool would add unbounded state to a hot path.
	SrcPoolOut = "pool_out"
	// SrcReceiptMiss is a receipt lookup that found nothing yet.
	SrcReceiptMiss = "receipt_miss"
	// SrcReceiptHit is a receipt lookup that found it.
	SrcReceiptHit = "receipt_hit"
)

// Record is the sink. Nil means nobody is watching, which is the normal state
// of a production node. The embedder sets it once, at init, before any chain is
// running; it is not safe to swap while transactions are in flight.
//
// n carries a number whose meaning depends on src -- gas for an arrival,
// nothing elsewhere. note carries free-form context and MUST NOT contain a
// comma, since sinks are expected to be line based.
var Record func(src string, hash common.Hash, n uint64, note string)

// Watching reports whether a sink is registered, so callers can skip building
// an argument that would be thrown away.
func Watching() bool { return Record != nil }

// Emit forwards to the sink if there is one.
func Emit(src string, hash common.Hash, n uint64, note string) {
	if Record != nil {
		Record(src, hash, n, note)
	}
}
