// SPDX-License-Identifier: LGPL-3.0-only
// Copyright 2025 Ava Labs, Inc. All rights reserved.
pragma solidity 0.8.30;

import {P256} from "./P256.sol";

contract P256SmokeTest {
    error ValidSignatureFailed();
    error InvalidSignaturePassed();

    event ValidSignaturePassed(bytes32 hash);

    constructor(bytes32 hash, bytes32 sigR, bytes32 sigS, bytes32 pubX, bytes32 pubY) {
        require(verify(hash, sigR, sigS, pubX, pubY), ValidSignatureFailed());
        emit ValidSignaturePassed(hash);

        bytes32 badHash = keccak256(abi.encode(hash));
        require(!verify(badHash, sigR, sigS, pubX, pubY), InvalidSignaturePassed());
    }

    function verify(bytes32 hash, bytes32 sigR, bytes32 sigS, bytes32 pubX, bytes32 pubY) public view returns (bool) {
        return P256.verify(hash, sigR, sigS, pubX, pubY);
    }
}
