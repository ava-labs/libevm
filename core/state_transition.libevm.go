// Copyright 2024 the libevm authors.
//
// The libevm additions to go-ethereum are free software: you can redistribute
// them and/or modify them under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// The libevm additions are distributed in the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see
// <http://www.gnu.org/licenses/>.

package core

import (
	"math/big"

	"github.com/holiman/uint256"

	"github.com/ava-labs/libevm/log"
	"github.com/ava-labs/libevm/params"
)

func (st *StateTransition) rulesHooks() params.RulesHooks {
	bCtx := st.evm.Context
	rules := st.evm.ChainConfig().Rules(bCtx.BlockNumber, bCtx.Random != nil, bCtx.Time)
	return rules.Hooks()
}

// canExecuteTransaction is a convenience wrapper for calling the
// [params.RulesHooks.CanExecuteTransaction] hook.
func (st *StateTransition) canExecuteTransaction() error {
	hooks := st.rulesHooks()
	if err := hooks.CanExecuteTransaction(st.msg.From, st.msg.To, st.state); err != nil {
		log.Debug(
			"Transaction execution blocked by libevm hook",
			"from", st.msg.From,
			"to", st.msg.To,
			"hooks", log.TypeOf(hooks),
			"reason", err,
		)
		return err
	}
	return nil
}

// consumeMinimumGas updates the gas remaining to reflect the value returned by
// [params.RulesHooks.MinimumGasConsumption]. It MUST be called after all code
// that modifies gas consumption; i.e. `st.gasRemaining` MUST remain constant
// after consumeMinimumGas returns.
func (st *StateTransition) consumeMinimumGas() {
	limit := st.msg.GasLimit
	minConsume := st.rulesHooks().MinimumGasConsumption(st.msg.GasLimit)
	if minConsume > limit {
		minConsume = limit // as documented in [params.RulesHooks]
	}

	maxRemaining := limit - minConsume
	if st.gasRemaining < maxRemaining {
		return
	}

	diff := st.gasRemaining - maxRemaining
	st.gasRemaining -= diff
	if err := st.gp.SubGas(diff); err != nil {
		// This would mean that the transaction wouldn't have been able to spend
		// up to its limit.
		log.Crit(
			"Broken gas-charging invariant",
			"tx limit", limit,
			"min consume", minConsume,
			"extra consume", diff,
			"SubGas() error", err,
		)
	}

	spend := new(big.Int).Mul(st.msg.GasPrice, new(big.Int).SetUint64(diff))
	st.state.SubBalance(st.msg.From, uint256.MustFromBig(spend))

	log.Debug(
		"Consumed extra gas to enforce minimum",
		"tx_limit", limit,
		"min_consumption", minConsume,
		"extra_consumption", diff,
	)
}
