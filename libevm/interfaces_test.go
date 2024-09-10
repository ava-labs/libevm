package libevm_test

import (
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/libevm"
)

// These two interfaces MUST be identical. If this breaks then the libevm copy
// MUST be updated.
var (
	// Each assignment demonstrates that the methods of the LHS interface are a
	// (non-strict) subset of the RHS interface's; both being possible
	// proves that they are identical.
	_ vm.PrecompiledContract     = (libevm.PrecompiledContract)(nil)
	_ libevm.PrecompiledContract = (vm.PrecompiledContract)(nil)
)
