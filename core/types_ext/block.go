// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package types contains data types related to Ethereum consensus.
package types_ext

import (
	"io"
	"math/big"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/rlp"
)

type (
	BlockNonce = types.BlockNonce
	Bloom      = types.Bloom
)

func init() {
	// Register the header type for RLP serialization
	types.RegisteredHeaderSerializer = headerSerializable{}
}

//go:generate go run ../../rlp/rlpgen -type headerSerializable -out gen_header_rlp.go
type Header = types.Header

func (headerSerializable) EncodeRLP(h *types.Header, w io.Writer) error {
	var hs headerSerializable

	// Set the shared fields with upstream
	hs.ParentHash = h.ParentHash
	hs.UncleHash = h.UncleHash
	hs.Coinbase = h.Coinbase
	hs.Root = h.Root
	hs.TxHash = h.TxHash
	hs.ReceiptHash = h.ReceiptHash
	hs.Bloom = h.Bloom
	hs.Difficulty = h.Difficulty
	hs.Number = h.Number
	hs.GasLimit = h.GasLimit
	hs.GasUsed = h.GasUsed
	hs.Time = h.Time
	hs.Extra = h.Extra
	hs.MixDigest = h.MixDigest

	// Set the extra payload
	if h.ExtraPayload != nil {
		hs.AddedStuff = h.ExtraPayload.(*headerExtra).AddedStuff
	}

	return rlp.Encode(w, &hs)
}

func (headerSerializable) DecodeRLP(h *types.Header, s *rlp.Stream) error {
	var hs headerSerializable
	if err := s.Decode(&hs); err != nil {
		return err
	}

	// Set the shared fields with upstream
	h.ParentHash = hs.ParentHash
	h.UncleHash = hs.UncleHash
	h.Coinbase = hs.Coinbase
	h.Root = hs.Root
	h.TxHash = hs.TxHash
	h.ReceiptHash = hs.ReceiptHash
	h.Bloom = hs.Bloom
	h.Difficulty = hs.Difficulty
	h.Number = hs.Number
	h.GasLimit = hs.GasLimit
	h.GasUsed = hs.GasUsed
	h.Time = hs.Time
	h.Extra = hs.Extra
	h.MixDigest = hs.MixDigest
	h.Nonce = hs.Nonce
	h.BaseFee = hs.BaseFee
	h.BlobGasUsed = hs.BlobGasUsed
	h.ExcessBlobGas = hs.ExcessBlobGas
	h.ParentBeaconRoot = hs.ParentBeaconRoot

	// Set the extra payload
	h.ExtraPayload = &headerExtra{
		AddedStuff: hs.AddedStuff,
	}
	return nil
}

// Separate the extra payload from the header, so there is no duplicate of
// shared fields.
type headerExtra struct {
	AddedStuff *common.Hash
}

// The purpose of this struct is to specify the exact RLP serialization of the
// header, including shared fields and fields related to the extra payload.
type headerSerializable struct {
	ParentHash  common.Hash    `json:"parentHash"       gencodec:"required"`
	UncleHash   common.Hash    `json:"sha3Uncles"       gencodec:"required"`
	Coinbase    common.Address `json:"miner"`
	Root        common.Hash    `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash    `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash    `json:"receiptsRoot"     gencodec:"required"`
	Bloom       Bloom          `json:"logsBloom"        gencodec:"required"`
	Difficulty  *big.Int       `json:"difficulty"       gencodec:"required"`
	Number      *big.Int       `json:"number"           gencodec:"required"`
	GasLimit    uint64         `json:"gasLimit"         gencodec:"required"`
	GasUsed     uint64         `json:"gasUsed"          gencodec:"required"`
	Time        uint64         `json:"timestamp"        gencodec:"required"`
	Extra       []byte         `json:"extraData"        gencodec:"required"`
	MixDigest   common.Hash    `json:"mixHash"`
	Nonce       BlockNonce     `json:"nonce"`

	// Added stuff in the middle of the header
	AddedStuff *common.Hash `json:"addedStuff" rlp:"optional"`

	// BaseFee was added by EIP-1559 and is ignored in legacy headers.
	BaseFee *big.Int `json:"baseFeePerGas" rlp:"optional"`

	// BlobGasUsed was added by EIP-4844 and is ignored in legacy headers.
	BlobGasUsed *uint64 `json:"blobGasUsed" rlp:"optional"`

	// ExcessBlobGas was added by EIP-4844 and is ignored in legacy headers.
	ExcessBlobGas *uint64 `json:"excessBlobGas" rlp:"optional"`

	// ParentBeaconRoot was added by EIP-4788 and is ignored in legacy headers.
	ParentBeaconRoot *common.Hash `json:"parentBeaconBlockRoot" rlp:"optional"`
}

// XXX: Need to support size too in upstream code
// var headerSize = common.StorageSize(reflect.TypeOf(Header{}).Size())
