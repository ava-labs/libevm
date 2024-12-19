// SPDX-License-Identifier: LGPL-3.0
pragma solidity 0.8.28;

/// @dev Interface of precompiled contract for implementation via `precompilegen`.
interface IPrecompile {
    function Echo(string memory) external view returns (string memory);

    function Echo(uint256) external view returns (uint256);

    function HashPacked(uint256, bytes2, address) external view returns (bytes32);

    struct Wrapper {
        int256 val;
    }

    function Extract(Wrapper memory) external view returns (int256);

    function Self() external view returns (address);

    function RevertWith(bytes memory) external;

    function View() external view returns (bool canReadState, bool canWriteState);

    function Pure() external pure returns (bool canReadState, bool canWriteState);

    function NeitherViewNorPure() external returns (bool canReadState, bool canWriteState);
}

/// @dev Testing contract to exercise the implementaiton of `IPrecompile`.
contract PrecompileTest {
    IPrecompile immutable precompile;

    constructor(IPrecompile _precompile) {
        precompile = _precompile;
    }

    /// @dev Emitted by each function to prove that it was successfully called.
    event Called(string func);

    function Echo(string memory x) external {
        assert(keccak256(abi.encodePacked(precompile.Echo(x))) == keccak256(abi.encodePacked(x)));
        emit Called("Echo(string)");
    }

    function Echo(uint256 x) external {
        assert(precompile.Echo(x) == x);
        emit Called("Echo(uint256)");
    }

    function Extract(IPrecompile.Wrapper memory x) external {
        assert(x.val == precompile.Extract(x));
        emit Called("Extract(...)");
    }

    function HashPacked(uint256 x, bytes2 y, address z) external {
        assert(precompile.HashPacked(x, y, z) == keccak256(abi.encodePacked(x, y, z)));
        emit Called("HashPacked(...)");
    }

    function Self() external {
        assert(precompile.Self() == address(precompile));
        emit Called("Self()");
    }

    function RevertWith(bytes memory err) external {
        (bool ok, bytes memory ret) =
            address(precompile).call(abi.encodeWithSelector(IPrecompile.RevertWith.selector, err));
        assert(!ok);
        assert(keccak256(ret) == keccak256(err));

        emit Called("RevertWith(...)");
    }

    function EchoingFallback(bytes memory input) external {
        bytes memory data = abi.encodeWithSelector( /*non-existent selector*/ 0, input);
        (bool ok, bytes memory ret) = address(precompile).staticcall(data);
        assert(ok);
        // Note equality with `data`, not `input` as the fallback echoes its
        // entire calldata, which includes the non-matching selector.
        assert(keccak256(ret) == keccak256(data));

        emit Called("EchoingFallback(...)");
    }

    function assertReadOnly(bytes4 selector, bool expectCanReadState, bool expectCanWriteState) internal {
        // We use a low-level call() here to ensure that the Solidity compiler
        // doesn't issue a staticcall() to view/pure functions as this would
        // hide our forcing of a read-only state.
        (bool ok, bytes memory ret) = address(precompile).call(abi.encodeWithSelector(selector));
        assert(ok);
        (bool canReadState, bool canWriteState) = abi.decode(ret, (bool, bool));
        assert(canReadState == expectCanReadState);
        assert(canWriteState == expectCanWriteState);
    }

    function View() external {
        assertReadOnly(IPrecompile.View.selector, true, false);
        emit Called("View()");
    }

    function Pure() external {
        assertReadOnly(IPrecompile.Pure.selector, false, false);
        emit Called("Pure()");
    }

    function NeitherViewNorPure() external {
        assertReadOnly(IPrecompile.NeitherViewNorPure.selector, true, true);
        emit Called("NeitherViewNorPure()");
    }
}
