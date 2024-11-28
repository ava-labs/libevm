// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package main

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ava-labs/libevm"
	"github.com/ava-labs/libevm/accounts/abi"
	"github.com/ava-labs/libevm/accounts/abi/bind"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IPrecompileWrapper is an auto generated low-level Go binding around an user-defined struct.
type IPrecompileWrapper struct {
	Val *big.Int
}

// PrecompileTestMetaData contains all meta data concerning the PrecompileTest contract.
var PrecompileTestMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIPrecompile\",\"name\":\"_precompile\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"func\",\"type\":\"string\"}],\"name\":\"Called\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"}],\"name\":\"Echo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"x\",\"type\":\"string\"}],\"name\":\"Echo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"input\",\"type\":\"bytes\"}],\"name\":\"EchoingFallback\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"int256\",\"name\":\"val\",\"type\":\"int256\"}],\"internalType\":\"structIPrecompile.Wrapper\",\"name\":\"x\",\"type\":\"tuple\"}],\"name\":\"Extract\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"},{\"internalType\":\"bytes2\",\"name\":\"y\",\"type\":\"bytes2\"},{\"internalType\":\"address\",\"name\":\"z\",\"type\":\"address\"}],\"name\":\"HashPacked\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"err\",\"type\":\"bytes\"}],\"name\":\"RevertWith\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"Self\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60a060405234801561000f575f5ffd5b506040516116d43803806116d4833981810160405281019061003191906100da565b8073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff168152505050610105565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6100988261006f565b9050919050565b5f6100a98261008e565b9050919050565b6100b98161009f565b81146100c3575f5ffd5b50565b5f815190506100d4816100b0565b92915050565b5f602082840312156100ef576100ee61006b565b5b5f6100fc848285016100c6565b91505092915050565b60805161158661014e5f395f81816101340152818161021401528181610378015281816104b2015281816105a0015281816105d7015281816106ea01526107f501526115865ff3fe608060405234801561000f575f5ffd5b506004361061007b575f3560e01c8063b4a8180811610059578063b4a81808146100d3578063c62c692f146100ef578063d7cc1f37146100f9578063db84d7c0146101155761007b565b806334d6d9be1461007f578063a93cbd971461009b578063ad8108a4146100b7575b5f5ffd5b6100996004803603810190610094919061093e565b610131565b005b6100b560048036038101906100b09190610aa5565b610210565b005b6100d160048036038101906100cc9190610b5c565b610376565b005b6100ed60048036038101906100e89190610aa5565b610458565b005b6100f761059e565b005b610113600480360381019061010e9190610c36565b6106bd565b005b61012f600480360381019061012a9190610d24565b6107cc565b005b807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166334d6d9be836040518263ffffffff1660e01b815260040161018b9190610d7a565b602060405180830381865afa1580156101a6573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906101ca9190610da7565b146101d8576101d7610dd2565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a60405161020590610e59565b60405180910390a150565b5f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663a93cbd9760e01b846040516024016102629190610ed7565b604051602081830303815290604052907bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff83818316178352505050506040516102cc9190610f31565b5f604051808303815f865af19150503d805f8114610305576040519150601f19603f3d011682016040523d82523d5f602084013e61030a565b606091505b5091509150811561031e5761031d610dd2565b5b828051906020012081805190602001201461033c5761033b610dd2565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a60405161036990610f91565b60405180910390a1505050565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663ad8108a4826040518263ffffffff1660e01b81526004016103cf9190610fd8565b602060405180830381865afa1580156103ea573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061040e9190611005565b815f0151146104205761041f610dd2565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a60405161044d9061107a565b60405180910390a150565b5f5f8260405160240161046b9190610ed7565b6040516020818303038152906040529060e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff838183161783525050505090505f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16836040516104f59190610f31565b5f60405180830381855afa9150503d805f811461052d576040519150601f19603f3d011682016040523d82523d5f602084013e610532565b606091505b50915091508161054557610544610dd2565b5b828051906020012081805190602001201461056357610562610dd2565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a604051610590906110e2565b60405180910390a150505050565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663c62c692f6040518163ffffffff1660e01b8152600401602060405180830381865afa15801561063e573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906106629190611114565b73ffffffffffffffffffffffffffffffffffffffff161461068657610685610dd2565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516106b390611189565b60405180910390a1565b8282826040516020016106d29392919061122c565b604051602081830303815290604052805190602001207f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663d7cc1f378585856040518463ffffffff1660e01b815260040161074593929190611286565b602060405180830381865afa158015610760573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061078491906112ee565b1461079257610791610dd2565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516107bf90611363565b60405180910390a1505050565b806040516020016107dd91906113c5565b604051602081830303815290604052805190602001207f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663db84d7c0836040518263ffffffff1660e01b815260040161084c9190611413565b5f60405180830381865afa158015610866573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f8201168201806040525081019061088e91906114a1565b60405160200161089e91906113c5565b60405160208183030381529060405280519060200120146108c2576108c1610dd2565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516108ef90611532565b60405180910390a150565b5f604051905090565b5f5ffd5b5f5ffd5b5f819050919050565b61091d8161090b565b8114610927575f5ffd5b50565b5f8135905061093881610914565b92915050565b5f6020828403121561095357610952610903565b5b5f6109608482850161092a565b91505092915050565b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6109b782610971565b810181811067ffffffffffffffff821117156109d6576109d5610981565b5b80604052505050565b5f6109e86108fa565b90506109f482826109ae565b919050565b5f67ffffffffffffffff821115610a1357610a12610981565b5b610a1c82610971565b9050602081019050919050565b828183375f83830152505050565b5f610a49610a44846109f9565b6109df565b905082815260208101848484011115610a6557610a6461096d565b5b610a70848285610a29565b509392505050565b5f82601f830112610a8c57610a8b610969565b5b8135610a9c848260208601610a37565b91505092915050565b5f60208284031215610aba57610ab9610903565b5b5f82013567ffffffffffffffff811115610ad757610ad6610907565b5b610ae384828501610a78565b91505092915050565b5f5ffd5b5f819050919050565b610b0281610af0565b8114610b0c575f5ffd5b50565b5f81359050610b1d81610af9565b92915050565b5f60208284031215610b3857610b37610aec565b5b610b4260206109df565b90505f610b5184828501610b0f565b5f8301525092915050565b5f60208284031215610b7157610b70610903565b5b5f610b7e84828501610b23565b91505092915050565b5f7fffff00000000000000000000000000000000000000000000000000000000000082169050919050565b610bbb81610b87565b8114610bc5575f5ffd5b50565b5f81359050610bd681610bb2565b92915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610c0582610bdc565b9050919050565b610c1581610bfb565b8114610c1f575f5ffd5b50565b5f81359050610c3081610c0c565b92915050565b5f5f5f60608486031215610c4d57610c4c610903565b5b5f610c5a8682870161092a565b9350506020610c6b86828701610bc8565b9250506040610c7c86828701610c22565b9150509250925092565b5f67ffffffffffffffff821115610ca057610c9f610981565b5b610ca982610971565b9050602081019050919050565b5f610cc8610cc384610c86565b6109df565b905082815260208101848484011115610ce457610ce361096d565b5b610cef848285610a29565b509392505050565b5f82601f830112610d0b57610d0a610969565b5b8135610d1b848260208601610cb6565b91505092915050565b5f60208284031215610d3957610d38610903565b5b5f82013567ffffffffffffffff811115610d5657610d55610907565b5b610d6284828501610cf7565b91505092915050565b610d748161090b565b82525050565b5f602082019050610d8d5f830184610d6b565b92915050565b5f81519050610da181610914565b92915050565b5f60208284031215610dbc57610dbb610903565b5b5f610dc984828501610d93565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52600160045260245ffd5b5f82825260208201905092915050565b7f4563686f2875696e7432353629000000000000000000000000000000000000005f82015250565b5f610e43600d83610dff565b9150610e4e82610e0f565b602082019050919050565b5f6020820190508181035f830152610e7081610e37565b9050919050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f610ea982610e77565b610eb38185610e81565b9350610ec3818560208601610e91565b610ecc81610971565b840191505092915050565b5f6020820190508181035f830152610eef8184610e9f565b905092915050565b5f81905092915050565b5f610f0b82610e77565b610f158185610ef7565b9350610f25818560208601610e91565b80840191505092915050565b5f610f3c8284610f01565b915081905092915050565b7f52657665727457697468282e2e2e2900000000000000000000000000000000005f82015250565b5f610f7b600f83610dff565b9150610f8682610f47565b602082019050919050565b5f6020820190508181035f830152610fa881610f6f565b9050919050565b610fb881610af0565b82525050565b602082015f820151610fd25f850182610faf565b50505050565b5f602082019050610feb5f830184610fbe565b92915050565b5f81519050610fff81610af9565b92915050565b5f6020828403121561101a57611019610903565b5b5f61102784828501610ff1565b91505092915050565b7f45787472616374282e2e2e2900000000000000000000000000000000000000005f82015250565b5f611064600c83610dff565b915061106f82611030565b602082019050919050565b5f6020820190508181035f83015261109181611058565b9050919050565b7f4563686f696e6746616c6c6261636b282e2e2e290000000000000000000000005f82015250565b5f6110cc601483610dff565b91506110d782611098565b602082019050919050565b5f6020820190508181035f8301526110f9816110c0565b9050919050565b5f8151905061110e81610c0c565b92915050565b5f6020828403121561112957611128610903565b5b5f61113684828501611100565b91505092915050565b7f53656c66282900000000000000000000000000000000000000000000000000005f82015250565b5f611173600683610dff565b915061117e8261113f565b602082019050919050565b5f6020820190508181035f8301526111a081611167565b9050919050565b5f819050919050565b6111c16111bc8261090b565b6111a7565b82525050565b5f819050919050565b6111e16111dc82610b87565b6111c7565b82525050565b5f8160601b9050919050565b5f6111fd826111e7565b9050919050565b5f61120e826111f3565b9050919050565b61122661122182610bfb565b611204565b82525050565b5f61123782866111b0565b60208201915061124782856111d0565b6002820191506112578284611215565b601482019150819050949350505050565b61127181610b87565b82525050565b61128081610bfb565b82525050565b5f6060820190506112995f830186610d6b565b6112a66020830185611268565b6112b36040830184611277565b949350505050565b5f819050919050565b6112cd816112bb565b81146112d7575f5ffd5b50565b5f815190506112e8816112c4565b92915050565b5f6020828403121561130357611302610903565b5b5f611310848285016112da565b91505092915050565b7f486173685061636b6564282e2e2e2900000000000000000000000000000000005f82015250565b5f61134d600f83610dff565b915061135882611319565b602082019050919050565b5f6020820190508181035f83015261137a81611341565b9050919050565b5f81519050919050565b5f81905092915050565b5f61139f82611381565b6113a9818561138b565b93506113b9818560208601610e91565b80840191505092915050565b5f6113d08284611395565b915081905092915050565b5f6113e582611381565b6113ef8185610dff565b93506113ff818560208601610e91565b61140881610971565b840191505092915050565b5f6020820190508181035f83015261142b81846113db565b905092915050565b5f61144561144084610c86565b6109df565b9050828152602081018484840111156114615761146061096d565b5b61146c848285610e91565b509392505050565b5f82601f83011261148857611487610969565b5b8151611498848260208601611433565b91505092915050565b5f602082840312156114b6576114b5610903565b5b5f82015167ffffffffffffffff8111156114d3576114d2610907565b5b6114df84828501611474565b91505092915050565b7f4563686f28737472696e672900000000000000000000000000000000000000005f82015250565b5f61151c600c83610dff565b9150611527826114e8565b602082019050919050565b5f6020820190508181035f83015261154981611510565b905091905056fea2646970667358221220cf4e683be87368d3d249175e8c416ebdb3928dae67bb586945f4f9798c59c2a664736f6c634300081c0033",
}

// PrecompileTestABI is the input ABI used to generate the binding from.
// Deprecated: Use PrecompileTestMetaData.ABI instead.
var PrecompileTestABI = PrecompileTestMetaData.ABI

// PrecompileTestBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use PrecompileTestMetaData.Bin instead.
var PrecompileTestBin = PrecompileTestMetaData.Bin

// DeployPrecompileTest deploys a new Ethereum contract, binding an instance of PrecompileTest to it.
func DeployPrecompileTest(auth *bind.TransactOpts, backend bind.ContractBackend, _precompile common.Address) (common.Address, *types.Transaction, *PrecompileTest, error) {
	parsed, err := PrecompileTestMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(PrecompileTestBin), backend, _precompile)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PrecompileTest{PrecompileTestCaller: PrecompileTestCaller{contract: contract}, PrecompileTestTransactor: PrecompileTestTransactor{contract: contract}, PrecompileTestFilterer: PrecompileTestFilterer{contract: contract}}, nil
}

// PrecompileTest is an auto generated Go binding around an Ethereum contract.
type PrecompileTest struct {
	PrecompileTestCaller     // Read-only binding to the contract
	PrecompileTestTransactor // Write-only binding to the contract
	PrecompileTestFilterer   // Log filterer for contract events
}

// PrecompileTestCaller is an auto generated read-only Go binding around an Ethereum contract.
type PrecompileTestCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrecompileTestTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PrecompileTestTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrecompileTestFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PrecompileTestFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrecompileTestSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PrecompileTestSession struct {
	Contract     *PrecompileTest   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PrecompileTestCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PrecompileTestCallerSession struct {
	Contract *PrecompileTestCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// PrecompileTestTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PrecompileTestTransactorSession struct {
	Contract     *PrecompileTestTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// PrecompileTestRaw is an auto generated low-level Go binding around an Ethereum contract.
type PrecompileTestRaw struct {
	Contract *PrecompileTest // Generic contract binding to access the raw methods on
}

// PrecompileTestCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PrecompileTestCallerRaw struct {
	Contract *PrecompileTestCaller // Generic read-only contract binding to access the raw methods on
}

// PrecompileTestTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PrecompileTestTransactorRaw struct {
	Contract *PrecompileTestTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPrecompileTest creates a new instance of PrecompileTest, bound to a specific deployed contract.
func NewPrecompileTest(address common.Address, backend bind.ContractBackend) (*PrecompileTest, error) {
	contract, err := bindPrecompileTest(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PrecompileTest{PrecompileTestCaller: PrecompileTestCaller{contract: contract}, PrecompileTestTransactor: PrecompileTestTransactor{contract: contract}, PrecompileTestFilterer: PrecompileTestFilterer{contract: contract}}, nil
}

// NewPrecompileTestCaller creates a new read-only instance of PrecompileTest, bound to a specific deployed contract.
func NewPrecompileTestCaller(address common.Address, caller bind.ContractCaller) (*PrecompileTestCaller, error) {
	contract, err := bindPrecompileTest(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PrecompileTestCaller{contract: contract}, nil
}

// NewPrecompileTestTransactor creates a new write-only instance of PrecompileTest, bound to a specific deployed contract.
func NewPrecompileTestTransactor(address common.Address, transactor bind.ContractTransactor) (*PrecompileTestTransactor, error) {
	contract, err := bindPrecompileTest(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PrecompileTestTransactor{contract: contract}, nil
}

// NewPrecompileTestFilterer creates a new log filterer instance of PrecompileTest, bound to a specific deployed contract.
func NewPrecompileTestFilterer(address common.Address, filterer bind.ContractFilterer) (*PrecompileTestFilterer, error) {
	contract, err := bindPrecompileTest(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PrecompileTestFilterer{contract: contract}, nil
}

// bindPrecompileTest binds a generic wrapper to an already deployed contract.
func bindPrecompileTest(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PrecompileTestMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PrecompileTest *PrecompileTestRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PrecompileTest.Contract.PrecompileTestCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PrecompileTest *PrecompileTestRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PrecompileTest.Contract.PrecompileTestTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PrecompileTest *PrecompileTestRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PrecompileTest.Contract.PrecompileTestTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PrecompileTest *PrecompileTestCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PrecompileTest.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PrecompileTest *PrecompileTestTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PrecompileTest.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PrecompileTest *PrecompileTestTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PrecompileTest.Contract.contract.Transact(opts, method, params...)
}

// Echo is a paid mutator transaction binding the contract method 0x34d6d9be.
//
// Solidity: function Echo(uint256 x) returns()
func (_PrecompileTest *PrecompileTestTransactor) Echo(opts *bind.TransactOpts, x *big.Int) (*types.Transaction, error) {
	return _PrecompileTest.contract.Transact(opts, "Echo", x)
}

// Echo is a paid mutator transaction binding the contract method 0x34d6d9be.
//
// Solidity: function Echo(uint256 x) returns()
func (_PrecompileTest *PrecompileTestSession) Echo(x *big.Int) (*types.Transaction, error) {
	return _PrecompileTest.Contract.Echo(&_PrecompileTest.TransactOpts, x)
}

// Echo is a paid mutator transaction binding the contract method 0x34d6d9be.
//
// Solidity: function Echo(uint256 x) returns()
func (_PrecompileTest *PrecompileTestTransactorSession) Echo(x *big.Int) (*types.Transaction, error) {
	return _PrecompileTest.Contract.Echo(&_PrecompileTest.TransactOpts, x)
}

// Echo0 is a paid mutator transaction binding the contract method 0xdb84d7c0.
//
// Solidity: function Echo(string x) returns()
func (_PrecompileTest *PrecompileTestTransactor) Echo0(opts *bind.TransactOpts, x string) (*types.Transaction, error) {
	return _PrecompileTest.contract.Transact(opts, "Echo0", x)
}

// Echo0 is a paid mutator transaction binding the contract method 0xdb84d7c0.
//
// Solidity: function Echo(string x) returns()
func (_PrecompileTest *PrecompileTestSession) Echo0(x string) (*types.Transaction, error) {
	return _PrecompileTest.Contract.Echo0(&_PrecompileTest.TransactOpts, x)
}

// Echo0 is a paid mutator transaction binding the contract method 0xdb84d7c0.
//
// Solidity: function Echo(string x) returns()
func (_PrecompileTest *PrecompileTestTransactorSession) Echo0(x string) (*types.Transaction, error) {
	return _PrecompileTest.Contract.Echo0(&_PrecompileTest.TransactOpts, x)
}

// EchoingFallback is a paid mutator transaction binding the contract method 0xb4a81808.
//
// Solidity: function EchoingFallback(bytes input) returns()
func (_PrecompileTest *PrecompileTestTransactor) EchoingFallback(opts *bind.TransactOpts, input []byte) (*types.Transaction, error) {
	return _PrecompileTest.contract.Transact(opts, "EchoingFallback", input)
}

// EchoingFallback is a paid mutator transaction binding the contract method 0xb4a81808.
//
// Solidity: function EchoingFallback(bytes input) returns()
func (_PrecompileTest *PrecompileTestSession) EchoingFallback(input []byte) (*types.Transaction, error) {
	return _PrecompileTest.Contract.EchoingFallback(&_PrecompileTest.TransactOpts, input)
}

// EchoingFallback is a paid mutator transaction binding the contract method 0xb4a81808.
//
// Solidity: function EchoingFallback(bytes input) returns()
func (_PrecompileTest *PrecompileTestTransactorSession) EchoingFallback(input []byte) (*types.Transaction, error) {
	return _PrecompileTest.Contract.EchoingFallback(&_PrecompileTest.TransactOpts, input)
}

// Extract is a paid mutator transaction binding the contract method 0xad8108a4.
//
// Solidity: function Extract((int256) x) returns()
func (_PrecompileTest *PrecompileTestTransactor) Extract(opts *bind.TransactOpts, x IPrecompileWrapper) (*types.Transaction, error) {
	return _PrecompileTest.contract.Transact(opts, "Extract", x)
}

// Extract is a paid mutator transaction binding the contract method 0xad8108a4.
//
// Solidity: function Extract((int256) x) returns()
func (_PrecompileTest *PrecompileTestSession) Extract(x IPrecompileWrapper) (*types.Transaction, error) {
	return _PrecompileTest.Contract.Extract(&_PrecompileTest.TransactOpts, x)
}

// Extract is a paid mutator transaction binding the contract method 0xad8108a4.
//
// Solidity: function Extract((int256) x) returns()
func (_PrecompileTest *PrecompileTestTransactorSession) Extract(x IPrecompileWrapper) (*types.Transaction, error) {
	return _PrecompileTest.Contract.Extract(&_PrecompileTest.TransactOpts, x)
}

// HashPacked is a paid mutator transaction binding the contract method 0xd7cc1f37.
//
// Solidity: function HashPacked(uint256 x, bytes2 y, address z) returns()
func (_PrecompileTest *PrecompileTestTransactor) HashPacked(opts *bind.TransactOpts, x *big.Int, y [2]byte, z common.Address) (*types.Transaction, error) {
	return _PrecompileTest.contract.Transact(opts, "HashPacked", x, y, z)
}

// HashPacked is a paid mutator transaction binding the contract method 0xd7cc1f37.
//
// Solidity: function HashPacked(uint256 x, bytes2 y, address z) returns()
func (_PrecompileTest *PrecompileTestSession) HashPacked(x *big.Int, y [2]byte, z common.Address) (*types.Transaction, error) {
	return _PrecompileTest.Contract.HashPacked(&_PrecompileTest.TransactOpts, x, y, z)
}

// HashPacked is a paid mutator transaction binding the contract method 0xd7cc1f37.
//
// Solidity: function HashPacked(uint256 x, bytes2 y, address z) returns()
func (_PrecompileTest *PrecompileTestTransactorSession) HashPacked(x *big.Int, y [2]byte, z common.Address) (*types.Transaction, error) {
	return _PrecompileTest.Contract.HashPacked(&_PrecompileTest.TransactOpts, x, y, z)
}

// RevertWith is a paid mutator transaction binding the contract method 0xa93cbd97.
//
// Solidity: function RevertWith(bytes err) returns()
func (_PrecompileTest *PrecompileTestTransactor) RevertWith(opts *bind.TransactOpts, err []byte) (*types.Transaction, error) {
	return _PrecompileTest.contract.Transact(opts, "RevertWith", err)
}

// RevertWith is a paid mutator transaction binding the contract method 0xa93cbd97.
//
// Solidity: function RevertWith(bytes err) returns()
func (_PrecompileTest *PrecompileTestSession) RevertWith(err []byte) (*types.Transaction, error) {
	return _PrecompileTest.Contract.RevertWith(&_PrecompileTest.TransactOpts, err)
}

// RevertWith is a paid mutator transaction binding the contract method 0xa93cbd97.
//
// Solidity: function RevertWith(bytes err) returns()
func (_PrecompileTest *PrecompileTestTransactorSession) RevertWith(err []byte) (*types.Transaction, error) {
	return _PrecompileTest.Contract.RevertWith(&_PrecompileTest.TransactOpts, err)
}

// Self is a paid mutator transaction binding the contract method 0xc62c692f.
//
// Solidity: function Self() returns()
func (_PrecompileTest *PrecompileTestTransactor) Self(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PrecompileTest.contract.Transact(opts, "Self")
}

// Self is a paid mutator transaction binding the contract method 0xc62c692f.
//
// Solidity: function Self() returns()
func (_PrecompileTest *PrecompileTestSession) Self() (*types.Transaction, error) {
	return _PrecompileTest.Contract.Self(&_PrecompileTest.TransactOpts)
}

// Self is a paid mutator transaction binding the contract method 0xc62c692f.
//
// Solidity: function Self() returns()
func (_PrecompileTest *PrecompileTestTransactorSession) Self() (*types.Transaction, error) {
	return _PrecompileTest.Contract.Self(&_PrecompileTest.TransactOpts)
}

// PrecompileTestCalledIterator is returned from FilterCalled and is used to iterate over the raw logs and unpacked data for Called events raised by the PrecompileTest contract.
type PrecompileTestCalledIterator struct {
	Event *PrecompileTestCalled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PrecompileTestCalledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrecompileTestCalled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PrecompileTestCalled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PrecompileTestCalledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrecompileTestCalledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrecompileTestCalled represents a Called event raised by the PrecompileTest contract.
type PrecompileTestCalled struct {
	Arg0 string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterCalled is a free log retrieval operation binding the contract event 0x3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a.
//
// Solidity: event Called(string func)
func (_PrecompileTest *PrecompileTestFilterer) FilterCalled(opts *bind.FilterOpts) (*PrecompileTestCalledIterator, error) {

	logs, sub, err := _PrecompileTest.contract.FilterLogs(opts, "Called")
	if err != nil {
		return nil, err
	}
	return &PrecompileTestCalledIterator{contract: _PrecompileTest.contract, event: "Called", logs: logs, sub: sub}, nil
}

// WatchCalled is a free log subscription operation binding the contract event 0x3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a.
//
// Solidity: event Called(string func)
func (_PrecompileTest *PrecompileTestFilterer) WatchCalled(opts *bind.WatchOpts, sink chan<- *PrecompileTestCalled) (event.Subscription, error) {

	logs, sub, err := _PrecompileTest.contract.WatchLogs(opts, "Called")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrecompileTestCalled)
				if err := _PrecompileTest.contract.UnpackLog(event, "Called", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCalled is a log parse operation binding the contract event 0x3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a.
//
// Solidity: event Called(string func)
func (_PrecompileTest *PrecompileTestFilterer) ParseCalled(log types.Log) (*PrecompileTestCalled, error) {
	event := new(PrecompileTestCalled)
	if err := _PrecompileTest.contract.UnpackLog(event, "Called", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
