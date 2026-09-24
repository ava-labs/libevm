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

    function Payable() external payable returns (uint256 value);

    function NonPayable() external;
}
