// SPDX-License-Identifier: LGPL-3.0-only
// Copyright 2025 Ava Labs, Inc. All rights reserved.
pragma solidity 0.8.30;

import {P256} from "./P256.sol";

contract P256Proxy {
    function verify(bytes32 hash, bytes32 sigR, bytes32 sigS, bytes32 pubX, bytes32 pubY) external view returns (bool) {
        return P256.verify(hash, sigR, sigS, pubX, pubY);
    }
}
