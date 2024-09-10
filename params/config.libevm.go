package params

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/libevm/pseudo"
)

// Extras are arbitrary payloads to be added as extra fields in [ChainConfig]
// and [Rules] structs. See [RegisterExtras].
type Extras[C any, R any] struct {
	// NewRules, if non-nil is called at the end of [ChainConfig.Rules] with the
	// newly created [Rules] and other context from the method call. Its
	// returned value will be the extra payload of the [Rules]. If NewRules is
	// nil then so too will the [Rules] extra payload be a nil `*R`.
	//
	// NewRules MAY modify the [Rules] but MUST NOT modify the [ChainConfig].
	NewRules func(_ *ChainConfig, _ *Rules, _ *C, blockNum *big.Int, isMerge bool, timestamp uint64) *R
}

// RegisterExtras registers the types `C` and `R` such that they are carried as
// extra payloads in [ChainConfig] and [Rules] structs, respectively. It is
// expected to be called in an `init()` function and MUST NOT be called more
// than once. Both `C` and `R` MUST be structs.
//
// After registration, JSON unmarshalling of a [ChainConfig] will create a new
// `*C` and unmarshal the JSON key "extra" into it. Conversely, JSON marshalling
// will populate the "extra" key with the contents of the `*C`. Both the
// [json.Marshaler] and [json.Unmarshaler] interfaces are honoured if
// implemented by `C` and/or `R.`
//
// Calls to [ChainConfig.Rules] will call the `NewRules` function of the
// registered [Extras] to create a new `*R`.
//
// The payloads can be accessed via the [ExtraPayloadGetter.FromChainConfig] and
// [ExtraPayloadGetter.FromRules] methods of the getter returned by
// RegisterExtras.
func RegisterExtras[C any, R any](e Extras[C, R]) ExtraPayloadGetter[C, R] {
	if registeredExtras != nil {
		panic("re-registration of Extras")
	}
	mustBeStruct[C]()
	mustBeStruct[R]()
	registeredExtras = &extraConstructors{
		chainConfig: pseudo.NewConstructor[C](),
		rules:       pseudo.NewConstructor[R](),
		newForRules: e.newForRules,
	}
	return e.getter()
}

// registeredExtras holds non-generic constructors for the [Extras] types
// registered via [RegisterExtras].
var registeredExtras *extraConstructors

type extraConstructors struct {
	chainConfig, rules pseudo.Constructor
	newForRules        func(_ *ChainConfig, _ *Rules, blockNum *big.Int, isMerge bool, timestamp uint64) *pseudo.Type
}

func (e *Extras[C, R]) newForRules(c *ChainConfig, r *Rules, blockNum *big.Int, isMerge bool, timestamp uint64) *pseudo.Type {
	if e.NewRules == nil {
		return registeredExtras.rules.NilPointer()
	}
	rExtra := e.NewRules(c, r, e.getter().FromChainConfig(c), blockNum, isMerge, timestamp)
	return pseudo.From(rExtra).Type
}

func (*Extras[C, R]) getter() (g ExtraPayloadGetter[C, R]) { return }

// mustBeStruct panics if `T` isn't a struct.
func mustBeStruct[T any]() {
	if k := reflect.TypeFor[T]().Kind(); k != reflect.Struct {
		panic(notStructMessage[T]())
	}
}

// notStructMessage returns the message with which [mustBeStruct] might panic.
// It exists to avoid change-detector tests should the message contents change.
func notStructMessage[T any]() string {
	var x T
	return fmt.Sprintf("%T is not a struct", x)
}

// An ExtraPayloadGettter provides strongly typed access to the extra payloads
// carried by [ChainConfig] and [Rules] structs. The only valid way to construct
// a getter is by a call to [RegisterExtras].
type ExtraPayloadGetter[C any, R any] struct {
	_ struct{} // make godoc show unexported fields so nobody tries to make their own getter ;)
}

// FromChainConfig returns the ChainConfig's extra payload.
func (ExtraPayloadGetter[C, R]) FromChainConfig(c *ChainConfig) *C {
	return pseudo.MustNewValue[*C](c.extraPayload()).Get()
}

// FromRules returns the Rules' extra payload.
func (ExtraPayloadGetter[C, R]) FromRules(r *Rules) *R {
	return pseudo.MustNewValue[*R](r.extraPayload()).Get()
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
func (c *ChainConfig) UnmarshalJSON(data []byte) error {
	type raw ChainConfig // doesn't inherit methods so avoids recursing back here (infinitely)
	cc := &struct {
		*raw
		Extra *pseudo.Type `json:"extra"`
	}{
		raw:   (*raw)(c),                                 // embedded to achieve regular JSON unmarshalling
		Extra: registeredExtras.chainConfig.NilPointer(), // `c.extra` is otherwise unexported
	}

	if err := json.Unmarshal(data, cc); err != nil {
		return err
	}
	c.extra = cc.Extra
	return nil
}

// MarshalJSON implements the [json.Marshaler] interface.
func (c *ChainConfig) MarshalJSON() ([]byte, error) {
	// See UnmarshalJSON() for rationale.
	type raw ChainConfig
	cc := &struct {
		*raw
		Extra *pseudo.Type `json:"extra"`
	}{raw: (*raw)(c), Extra: c.extra}
	return json.Marshal(cc)
}

var _ interface {
	json.Marshaler
	json.Unmarshaler
} = (*ChainConfig)(nil)

// addRulesExtra is called at the end of [ChainConfig.Rules]; it exists to
// abstract the libevm-specific behaviour outside of original geth code.
func (c *ChainConfig) addRulesExtra(r *Rules, blockNum *big.Int, isMerge bool, timestamp uint64) {
	r.extra = nil
	if registeredExtras != nil {
		r.extra = registeredExtras.newForRules(c, r, blockNum, isMerge, timestamp)
	}
}

// extraPayload returns the ChainConfig's extra payload iff [RegisterExtras] has
// already been called. If the payload hasn't been populated (typically via
// unmarshalling of JSON), a nil value is constructed and returned.
func (c *ChainConfig) extraPayload() *pseudo.Type {
	if registeredExtras == nil {
		// This will only happen if someone constructs an [ExtraPayloadGetter]
		// directly, without a call to [RegisterExtras].
		//
		// See https://google.github.io/styleguide/go/best-practices#when-to-panic
		panic(fmt.Sprintf("%T.ExtraPayload() called before RegisterExtras()", c))
	}
	if c.extra == nil {
		c.extra = registeredExtras.chainConfig.NilPointer()
	}
	return c.extra
}

// extraPayload is equivalent to [ChainConfig.extraPayload].
func (r *Rules) extraPayload() *pseudo.Type {
	if registeredExtras == nil {
		// See ChainConfig.extraPayload() equivalent.
		panic(fmt.Sprintf("%T.ExtraPayload() called before RegisterExtras()", r))
	}
	if r.extra == nil {
		r.extra = registeredExtras.rules.NilPointer()
	}
	return r.extra
}
