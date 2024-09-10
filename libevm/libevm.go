package libevm

// PrecompiledContract is an exact copy of vm.PrecompiledContract, mirrored here
// for instances where importing that package would result in a circular
// dependency.
type PrecompiledContract interface {
	RequiredGas(input []byte) uint64
	Run(input []byte) ([]byte, error)
}
