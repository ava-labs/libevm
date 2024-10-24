package state

import (
	"testing"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/stretchr/testify/require"
)

func TestStateObjectEmpty(t *testing.T) {
	tests := []struct {
		name           string
		registerAndSet func(*types.StateAccount)
		wantEmpty      bool
	}{
		{
			name:           "no registered types.StateAccount extra payload",
			registerAndSet: func(*types.StateAccount) {},
			wantEmpty:      true,
		},
		{
			name: "erroneously non-nil types.StateAccountExtra when no registered payload",
			registerAndSet: func(acc *types.StateAccount) {
				acc.Extra = &types.StateAccountExtra{}
			},
			wantEmpty: true,
		},
		{
			name: "explicit false bool",
			registerAndSet: func(acc *types.StateAccount) {
				types.RegisterExtras[bool]().SetOnStateAccount(acc, false)
			},
			wantEmpty: true,
		},
		{
			name: "implicit false bool",
			registerAndSet: func(*types.StateAccount) {
				types.RegisterExtras[bool]()
			},
			wantEmpty: true,
		},
		{
			name: "true bool",
			registerAndSet: func(acc *types.StateAccount) {
				types.RegisterExtras[bool]().SetOnStateAccount(acc, true)
			},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			types.TestOnlyClearRegisteredExtras()
			t.Cleanup(types.TestOnlyClearRegisteredExtras)

			obj := newObject(nil, common.Address{}, nil)
			tt.registerAndSet(&obj.data)
			require.Equalf(t, tt.wantEmpty, obj.empty(), "%T.empty()", obj)
		})
	}
}
