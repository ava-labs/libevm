// SPDX-License-Identifier: LGPL-3.0-only
// Copyright 2025 Ava Labs, Inc. All rights reserved.
pragma solidity ^0.8.0;

library P256 {
    uint256 private constant out = 0;

    function verify(bytes32 hash, bytes32 sigR, bytes32 sigS, bytes32 pubX, bytes32 pubY) internal view returns (bool) {
        bool verified;
        assembly ("memory-safe") {
            mstore(out, 0)

            let buf := mload(0x40)
            mstore(buf, hash)
            mstore(add(buf, 0x20), sigR)
            mstore(add(buf, 0x40), sigS)
            mstore(add(buf, 0x60), pubX)
            mstore(add(buf, 0x80), pubY)

            let ok := staticcall(gas(), 0x100, buf, 0xa0, out, 0x20)
            verified := and(
                ok,
                mload(out) // known to be zero or one
            )
        }
        return verified;
    }
}
