# libevm

[![API Reference](https://pkg.go.dev/badge/github.com/ava-labs/libevm)](https://pkg.go.dev/github.com/ava-labs/libevm?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/ava-labs/libevm)](https://goreportcard.com/report/github.com/ava-labs/libevm)
[![Go Build & Test](https://github.com/ava-labs/libevm/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/ava-labs/libevm/actions/workflows/go.yml)

The Ethereum Virtual Machine (EVM) as a library, `libevm` is a fork of [`geth`](https://github.com/ethereum/go-ethereum) with injectable configuration directives.
Although designed to support the Avalanche [C-Chain](https://github.com/ava-labs/coreth) and [EVM-L1s](https://github.com/ava-labs/subnet-evm) (formerly *subnets*), configuration is general-purpose and backwards-compatible with `geth`.
We are immensely grateful for the hard work of the `geth` authors, and hope that our contribution can be of value to others too. Thank you!

Libevm is notably following these rules:

- minimal changes to geth-original code
- any extension to the code must be compatible with geth
- any extension to the code must be done in separate files, usually ending with `.libevm.go` or `.libevm_test.go`

This allows for:

- easy Geth fork-sync upgrade
- Common code base to use for its consumers [coreth](https://github.com/ava-labs/coreth) and [subnet-evm](https://github.com/ava-labs/subnet-evm), with customization defined on the consumer side.

The reason for this fork was essentially to reduce the maintenance burden of [coreth](https://github.com/ava-labs/coreth) and [subnet-evm](https://github.com/ava-labs/subnet-evm):

- upgrading the geth base was very complex, long and risky
- there was a constant worry about breaking changes in geth: is this geth source code here, will this make it incompatible with Geth?

## Geth upgrade process

1. libevm should be in sync with the targeted geth release, with a release tag ready to be used for its consumers.
1. Consumers (coreth and subnet-evm) should bump their libevm dependency version to the one previously created.
1. Consumers (coreth and subnet-evm) should manually compare and adapt their code to match the targeted geth release code.
This is how it was done before libevm, the difference being that now this is a reduced effort.

## Implementation of libevm

libevm relies on "hooks", which allows for libevm consumers to define a custom behavior for various places in the geth code, with minimal changes to the geth original code.

Essentially a hook is an interface defined in libevm, and implemented in its consumers.

Each hook is *registered* by the consumer by calling **once** at global scope a libevm-defined `RegisterExtras` function, with its own implementation of the hook. If the `RegisterExtras` function is not called, the hook defaults to a no-op hook, matching Geth's original behavior; this is the case in libevm code on its own.

The following highlights some of the hooks defined in libevm, together with examples of their implementations in libevm consumers. These are based on:

- libevm as of commit [979064c](https://github.com/ava-labs/libevm/blob/979064cfdbc1aa8dfae118dfb1344db34d37a164)
- coreth as of commit [0d68be6](https://github.com/ava-labs/coreth/blob/0d68be6b92be7c34095487b3a512b87b8b923caa)
- subnet-evm as of commit [4f496a0](https://github.com/ava-labs/subnet-evm/blob/4f496a00f226309aa701d33ac28b33658bb2b697)

 Note it may become outdated as the code evolves, but it can still serve as an example.

### `params.ChainConfigHooks`

The `ChainConfigHooks` interface can be implemented and registered to customize the behavior of the chain configuration in libevm. It is defined in [`params/hooks.libevm.go`](../params/hooks.libevm.go).

The hook is retrieved from the method `ChainConfig.Hooks()`.

It is used in:

- the `ChainConfig.Description` method, to append the libevm consumer's description to the Geth chain configuration description, using the hook method `Description` method.
- the `ChainConfig.checkCompatible` method, to check if the libevm consumer's chain configuration is compatible on top of the Geth compatibility check, using the hook method `CheckConfigCompatible` method.
- the `ChainConfig.CheckConfigForkOrder` method, to check if the libevm consumer's chain configuration has its fork order correct on top of the Geth fork order check, using the hook method `CheckConfigForkOrder` method.

Implementations of the hooks:

- [coreth](https://github.com/ava-labs/coreth/blob/0d68be6b92be7c34095487b3a512b87b8b923caa/params/extras/config.go#L113)
- [subnet-evm](https://github.com/ava-labs/subnet-evm/blob/4f496a00f226309aa701d33ac28b33658bb2b697/params/extras/config.go#L118)

### `params.RulesHooks`

The `RulesHooks` interface can be implemented and registered to customize the behavior of the chain configuration rules in libevm. It is defined in [`params/hooks.libevm.go`](../params/hooks.libevm.go).

The hook is retrieved from the method `Rules.Hooks()`.

It is used in:

- the `core/vm` package `EVM.canCreateContract` method, to check if, according to the libevm consumer rules, a contract can be created, using the hook method `CanCreateContract`.
- the `core/vm` package `EVM.precompile` method, to override precompile if needed, according to the libevm consumer rules, and using the hook method `PrecompileOverride`.
- the `core/vm` package `ActivePrecompiles` function, to add the libevm consumer active precompiles to the active precompiles of Geth, using the hook method `ActivePrecompiles`.
- the `core` package `StateTransition.TransitionDb` method, to check if the libevm consumer rules allow to execute a transaction, using the hook method `CanExecuteTransaction`.

Implementations of the hooks:

- [coreth](https://github.com/ava-labs/coreth/blob/0d68be6b92be7c34095487b3a512b87b8b923caa/params/hooks_libevm.go#L25)
- [subnet-evm](https://github.com/ava-labs/subnet-evm/blob/4f496a00f226309aa701d33ac28b33658bb2b697/params/hooks_libevm.go#L24)

### `core/vm.Hooks`

The `vm.Hooks` interface can be implemented and registered to customize the behavior of the EVM in libevm. It is defined in [`core/vm/hooks.libevm.go`](../core/vm/hooks.libevm.go).

The hook is registered with `RegisterHooks` and is directly assigned to a global variable `libevmHooks`, which is then used to access the hook.

It is used in:

- the `core/vm` package `NewEVM` function, to override EVM creation arguments (such as the block context, state database, etc.), using the hook method `OverrideNewEVMArgs`.
- the `core/vm` package `EVM.Reset` method, to override EVM reset arguments (transaction context and state database), using the hook method `OverrideEVMResetArgs`.

Implementations of the hooks:

- [coreth](https://github.com/ava-labs/coreth/blob/0d68be6b92be7c34095487b3a512b87b8b923caa/core/evm.go#L50)
- [subnet-evm](https://github.com/ava-labs/subnet-evm/blob/4f496a00f226309aa701d33ac28b33658bb2b697/core/evm.go#L50)

### `core/types.HeaderHooks`

The `HeaderHooks` interface can be implemented and registered to customize the behavior of the block header in libevm. It is defined in [`core/types/block.libevm.go`](../core/types/block.libevm.go).

The hook is retrieved from the method `Header.hooks()`.

It is used exclusively in the `core/types` package in:

- the `CopyHeader` function, to deep copy extra header fields defined by the libevm consumer, using the hook method `PostCopy`.
- the Header JSON encoding and decoding, by hooking its `EncodeJSON` and `DecodeJSON` methods respectively in the `Header.MarshalJSON` and `Header.UnmarshalJSON` methods.
- the Header RLP encoding and decoding, by hooking its `EncodeRLP` and `DecodeRLP` methods respectively in the `Header.EncodeRLP` and `Header.DecodeRLP` methods.

Implementations of the hooks:

- [coreth](https://github.com/ava-labs/coreth/blob/0d68be6b92be7c34095487b3a512b87b8b923caa/plugin/evm/customtypes/header_ext.go#L37)
- [subnet-evm](https://github.com/ava-labs/subnet-evm/blob/4f496a00f226309aa701d33ac28b33658bb2b697/plugin/evm/customtypes/header_ext.go#L37)

### `core/types.BlockBody`

The `BlockBody` interface can be implemented and registered to customize the behavior of the block **and** body in libevm. It is defined in [`core/types/block.libevm.go`](../core/types/block.libevm.go).

The hook is retrieved either from:

- the method `Block.hooks()`
- the method `Body.hooks()`

This is because the hook, so far in the libevm consumers, has the same "payload" (fields in the implementation of `BlockBody`) for both the Body and the Block. It does have different methods for the `Block` and `Body` though, as explained below.

It is used exclusively in the `core/types` package in:

- the Block RLP encoding by injecting the entire hooks in the RLP encoding through `Block.EncodeRLP`. The hook method `BlockRLPFieldsForEncoding` is then picked up in the hook-compatible added code `extblock.EncodeRLP`, where `extblock` is the geth-internal typed used for RLP encoding. Its `EncodeRLP` method is added code to customize the RLP encoding through hooks.
- the Block RLP decoding by injecting the entire hooks in the RLP decoding through `Block.DecodeRLP`. The hook method `BlockRLPFieldPointersForDecoding` is then picked up in the hook-compatible added code `extblock.DecodeRLP`, where `extblock` is the geth-internal typed used for RLP decoding. Its `DecodeRLP` method is added code to customize the RLP decoding through hooks.
- the Body RLP encoding by using the hook method `BodyRLPFieldsForEncoding` directly in the `Body.EncodeRLP` method, which is new separate code.
- the Body RLP decoding by using the hook method `BodyRLPFieldPointersForDecoding` directly in the `Body.DecodeRLP` method, which is new separate code.

Implementations of the hooks:

- [coreth](https://github.com/ava-labs/coreth/blob/0d68be6b92be7c34095487b3a512b87b8b923caa/plugin/evm/customtypes/block_ext.go#L22)
- subnet-evm: **NONE!** Because subnet-evm uses the same block and body (body at least from an RLP point of view) as Geth, there is no need for hooks or extra payloads.
