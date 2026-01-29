package filters

import "github.com/ava-labs/libevm/core/types"

func getBloomFromHeader(header *types.Header, backend Backend) types.Bloom {
	type bloomHeader interface {
		HeaderBloom(*types.Header) types.Bloom
	}

	if bh, ok := backend.(bloomHeader); ok {
		return bh.HeaderBloom(header)
	}
	return header.Bloom
}
