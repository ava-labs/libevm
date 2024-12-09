package types_ext

import (
	"math/big"
	"testing"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/rlp"
	"github.com/stretchr/testify/require"
)

func TestRLP(t *testing.T) {
	// expected (eg, from network or disk)
	added := common.Hash{111}
	asIs := headerSerializable{
		Number:     big.NewInt(10), // Shared field
		AddedStuff: &added,         // Added field
	}

	rlpAsIs, err := rlp.EncodeToBytes(&asIs)
	require.NoError(t, err)

	// now parse it via upstream type
	var upstream Header
	err = rlp.DecodeBytes(rlpAsIs, &upstream)
	require.NoError(t, err)

	// check that the fields are the same
	require.Equal(t, asIs.Number, upstream.Number)
	require.Equal(t, asIs.AddedStuff, upstream.ExtraPayload.(*headerExtra).AddedStuff)

	// now encode the upstream type
	rlpUpstream, err := rlp.EncodeToBytes(&upstream)
	require.NoError(t, err)
	require.Equal(t, rlpAsIs, rlpUpstream)
}
