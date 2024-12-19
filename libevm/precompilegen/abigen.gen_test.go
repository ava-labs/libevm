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
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIPrecompile\",\"name\":\"_precompile\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"func\",\"type\":\"string\"}],\"name\":\"Called\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"}],\"name\":\"Echo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"x\",\"type\":\"string\"}],\"name\":\"Echo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"input\",\"type\":\"bytes\"}],\"name\":\"EchoingFallback\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"int256\",\"name\":\"val\",\"type\":\"int256\"}],\"internalType\":\"structIPrecompile.Wrapper\",\"name\":\"x\",\"type\":\"tuple\"}],\"name\":\"Extract\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"},{\"internalType\":\"bytes2\",\"name\":\"y\",\"type\":\"bytes2\"},{\"internalType\":\"address\",\"name\":\"z\",\"type\":\"address\"}],\"name\":\"HashPacked\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"NeitherViewNorPure\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"Pure\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"err\",\"type\":\"bytes\"}],\"name\":\"RevertWith\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"Self\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"View\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60a060405234801561000f575f5ffd5b50604051611ae8380380611ae8833981810160405281019061003191906100da565b8073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff168152505050610105565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6100988261006f565b9050919050565b5f6100a98261008e565b9050919050565b6100b98161009f565b81146100c3575f5ffd5b50565b5f815190506100d4816100b0565b92915050565b5f602082840312156100ef576100ee61006b565b5b5f6100fc848285016100c6565b91505092915050565b6080516119936101555f395f81816101bd015281816102e70152818161044b0152818161058501528181610673015281816106aa015281816107bd015281816108c80152610a1a01526119935ff3fe608060405234801561000f575f5ffd5b506004361061009c575f3560e01c8063b4a8180811610064578063b4a8180814610108578063c62c692f14610124578063d7cc1f371461012e578063db84d7c01461014a578063f5ad2fb8146101665761009c565b80631686f265146100a057806334d6d9be146100aa578063a7263000146100c6578063a93cbd97146100d0578063ad8108a4146100ec575b5f5ffd5b6100a8610170565b005b6100c460048036038101906100bf9190610ba0565b6101ba565b005b6100ce610299565b005b6100ea60048036038101906100e59190610d07565b6102e3565b005b61010660048036038101906101019190610dbe565b610449565b005b610122600480360381019061011d9190610d07565b61052b565b005b61012c610671565b005b61014860048036038101906101439190610e98565b610790565b005b610164600480360381019061015f9190610f86565b61089f565b005b61016e6109cd565b005b610183631686f26560e01b60015f610a16565b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516101b090611027565b60405180910390a1565b807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166334d6d9be836040518263ffffffff1660e01b81526004016102149190611054565b602060405180830381865afa15801561022f573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102539190611081565b14610261576102606110ac565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a60405161028e90611123565b60405180910390a150565b6102ac63a726300060e01b600180610a16565b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516102d99061118b565b60405180910390a1565b5f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663a93cbd9760e01b846040516024016103359190611209565b604051602081830303815290604052907bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff838183161783525050505060405161039f9190611263565b5f604051808303815f865af19150503d805f81146103d8576040519150601f19603f3d011682016040523d82523d5f602084013e6103dd565b606091505b509150915081156103f1576103f06110ac565b5b828051906020012081805190602001201461040f5761040e6110ac565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a60405161043c906112c3565b60405180910390a1505050565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663ad8108a4826040518263ffffffff1660e01b81526004016104a2919061130a565b602060405180830381865afa1580156104bd573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906104e19190611337565b815f0151146104f3576104f26110ac565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a604051610520906113ac565b60405180910390a150565b5f5f8260405160240161053e9190611209565b6040516020818303038152906040529060e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff838183161783525050505090505f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16836040516105c89190611263565b5f60405180830381855afa9150503d805f8114610600576040519150601f19603f3d011682016040523d82523d5f602084013e610605565b606091505b509150915081610618576106176110ac565b5b8280519060200120818051906020012014610636576106356110ac565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a60405161066390611414565b60405180910390a150505050565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663c62c692f6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610711573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906107359190611446565b73ffffffffffffffffffffffffffffffffffffffff1614610759576107586110ac565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a604051610786906114bb565b60405180910390a1565b8282826040516020016107a59392919061155e565b604051602081830303815290604052805190602001207f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663d7cc1f378585856040518463ffffffff1660e01b8152600401610818939291906115b8565b602060405180830381865afa158015610833573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906108579190611620565b14610865576108646110ac565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a60405161089290611695565b60405180910390a1505050565b806040516020016108b091906116f7565b604051602081830303815290604052805190602001207f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663db84d7c0836040518263ffffffff1660e01b815260040161091f9190611745565b5f60405180830381865afa158015610939573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f8201168201806040525081019061096191906117d3565b60405160200161097191906116f7565b6040516020818303038152906040528051906020012014610995576109946110ac565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516109c290611864565b60405180910390a150565b6109df63f5ad2fb860e01b5f5f610a16565b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a604051610a0c906118cc565b60405180910390a1565b5f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1685604051602401604051602081830303815290604052907bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050604051610ac09190611263565b5f604051808303815f865af19150503d805f8114610af9576040519150601f19603f3d011682016040523d82523d5f602084013e610afe565b606091505b509150915081610b1157610b106110ac565b5b5f5f82806020019051810190610b27919061191f565b9150915085151582151514610b3f57610b3e6110ac565b5b84151581151514610b5357610b526110ac565b5b50505050505050565b5f604051905090565b5f5ffd5b5f5ffd5b5f819050919050565b610b7f81610b6d565b8114610b89575f5ffd5b50565b5f81359050610b9a81610b76565b92915050565b5f60208284031215610bb557610bb4610b65565b5b5f610bc284828501610b8c565b91505092915050565b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b610c1982610bd3565b810181811067ffffffffffffffff82111715610c3857610c37610be3565b5b80604052505050565b5f610c4a610b5c565b9050610c568282610c10565b919050565b5f67ffffffffffffffff821115610c7557610c74610be3565b5b610c7e82610bd3565b9050602081019050919050565b828183375f83830152505050565b5f610cab610ca684610c5b565b610c41565b905082815260208101848484011115610cc757610cc6610bcf565b5b610cd2848285610c8b565b509392505050565b5f82601f830112610cee57610ced610bcb565b5b8135610cfe848260208601610c99565b91505092915050565b5f60208284031215610d1c57610d1b610b65565b5b5f82013567ffffffffffffffff811115610d3957610d38610b69565b5b610d4584828501610cda565b91505092915050565b5f5ffd5b5f819050919050565b610d6481610d52565b8114610d6e575f5ffd5b50565b5f81359050610d7f81610d5b565b92915050565b5f60208284031215610d9a57610d99610d4e565b5b610da46020610c41565b90505f610db384828501610d71565b5f8301525092915050565b5f60208284031215610dd357610dd2610b65565b5b5f610de084828501610d85565b91505092915050565b5f7fffff00000000000000000000000000000000000000000000000000000000000082169050919050565b610e1d81610de9565b8114610e27575f5ffd5b50565b5f81359050610e3881610e14565b92915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610e6782610e3e565b9050919050565b610e7781610e5d565b8114610e81575f5ffd5b50565b5f81359050610e9281610e6e565b92915050565b5f5f5f60608486031215610eaf57610eae610b65565b5b5f610ebc86828701610b8c565b9350506020610ecd86828701610e2a565b9250506040610ede86828701610e84565b9150509250925092565b5f67ffffffffffffffff821115610f0257610f01610be3565b5b610f0b82610bd3565b9050602081019050919050565b5f610f2a610f2584610ee8565b610c41565b905082815260208101848484011115610f4657610f45610bcf565b5b610f51848285610c8b565b509392505050565b5f82601f830112610f6d57610f6c610bcb565b5b8135610f7d848260208601610f18565b91505092915050565b5f60208284031215610f9b57610f9a610b65565b5b5f82013567ffffffffffffffff811115610fb857610fb7610b69565b5b610fc484828501610f59565b91505092915050565b5f82825260208201905092915050565b7f56696577282900000000000000000000000000000000000000000000000000005f82015250565b5f611011600683610fcd565b915061101c82610fdd565b602082019050919050565b5f6020820190508181035f83015261103e81611005565b9050919050565b61104e81610b6d565b82525050565b5f6020820190506110675f830184611045565b92915050565b5f8151905061107b81610b76565b92915050565b5f6020828403121561109657611095610b65565b5b5f6110a38482850161106d565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52600160045260245ffd5b7f4563686f2875696e7432353629000000000000000000000000000000000000005f82015250565b5f61110d600d83610fcd565b9150611118826110d9565b602082019050919050565b5f6020820190508181035f83015261113a81611101565b9050919050565b7f4e656974686572566965774e6f725075726528290000000000000000000000005f82015250565b5f611175601483610fcd565b915061118082611141565b602082019050919050565b5f6020820190508181035f8301526111a281611169565b9050919050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f6111db826111a9565b6111e581856111b3565b93506111f58185602086016111c3565b6111fe81610bd3565b840191505092915050565b5f6020820190508181035f83015261122181846111d1565b905092915050565b5f81905092915050565b5f61123d826111a9565b6112478185611229565b93506112578185602086016111c3565b80840191505092915050565b5f61126e8284611233565b915081905092915050565b7f52657665727457697468282e2e2e2900000000000000000000000000000000005f82015250565b5f6112ad600f83610fcd565b91506112b882611279565b602082019050919050565b5f6020820190508181035f8301526112da816112a1565b9050919050565b6112ea81610d52565b82525050565b602082015f8201516113045f8501826112e1565b50505050565b5f60208201905061131d5f8301846112f0565b92915050565b5f8151905061133181610d5b565b92915050565b5f6020828403121561134c5761134b610b65565b5b5f61135984828501611323565b91505092915050565b7f45787472616374282e2e2e2900000000000000000000000000000000000000005f82015250565b5f611396600c83610fcd565b91506113a182611362565b602082019050919050565b5f6020820190508181035f8301526113c38161138a565b9050919050565b7f4563686f696e6746616c6c6261636b282e2e2e290000000000000000000000005f82015250565b5f6113fe601483610fcd565b9150611409826113ca565b602082019050919050565b5f6020820190508181035f83015261142b816113f2565b9050919050565b5f8151905061144081610e6e565b92915050565b5f6020828403121561145b5761145a610b65565b5b5f61146884828501611432565b91505092915050565b7f53656c66282900000000000000000000000000000000000000000000000000005f82015250565b5f6114a5600683610fcd565b91506114b082611471565b602082019050919050565b5f6020820190508181035f8301526114d281611499565b9050919050565b5f819050919050565b6114f36114ee82610b6d565b6114d9565b82525050565b5f819050919050565b61151361150e82610de9565b6114f9565b82525050565b5f8160601b9050919050565b5f61152f82611519565b9050919050565b5f61154082611525565b9050919050565b61155861155382610e5d565b611536565b82525050565b5f61156982866114e2565b6020820191506115798285611502565b6002820191506115898284611547565b601482019150819050949350505050565b6115a381610de9565b82525050565b6115b281610e5d565b82525050565b5f6060820190506115cb5f830186611045565b6115d8602083018561159a565b6115e560408301846115a9565b949350505050565b5f819050919050565b6115ff816115ed565b8114611609575f5ffd5b50565b5f8151905061161a816115f6565b92915050565b5f6020828403121561163557611634610b65565b5b5f6116428482850161160c565b91505092915050565b7f486173685061636b6564282e2e2e2900000000000000000000000000000000005f82015250565b5f61167f600f83610fcd565b915061168a8261164b565b602082019050919050565b5f6020820190508181035f8301526116ac81611673565b9050919050565b5f81519050919050565b5f81905092915050565b5f6116d1826116b3565b6116db81856116bd565b93506116eb8185602086016111c3565b80840191505092915050565b5f61170282846116c7565b915081905092915050565b5f611717826116b3565b6117218185610fcd565b93506117318185602086016111c3565b61173a81610bd3565b840191505092915050565b5f6020820190508181035f83015261175d818461170d565b905092915050565b5f61177761177284610ee8565b610c41565b90508281526020810184848401111561179357611792610bcf565b5b61179e8482856111c3565b509392505050565b5f82601f8301126117ba576117b9610bcb565b5b81516117ca848260208601611765565b91505092915050565b5f602082840312156117e8576117e7610b65565b5b5f82015167ffffffffffffffff81111561180557611804610b69565b5b611811848285016117a6565b91505092915050565b7f4563686f28737472696e672900000000000000000000000000000000000000005f82015250565b5f61184e600c83610fcd565b91506118598261181a565b602082019050919050565b5f6020820190508181035f83015261187b81611842565b9050919050565b7f50757265282900000000000000000000000000000000000000000000000000005f82015250565b5f6118b6600683610fcd565b91506118c182611882565b602082019050919050565b5f6020820190508181035f8301526118e3816118aa565b9050919050565b5f8115159050919050565b6118fe816118ea565b8114611908575f5ffd5b50565b5f81519050611919816118f5565b92915050565b5f5f6040838503121561193557611934610b65565b5b5f6119428582860161190b565b92505060206119538582860161190b565b915050925092905056fea26469706673582212206fbd1578963795306db34fa421e0e514e3e99979c9760a9b6f393505a99b5b3864736f6c634300081c0033",
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

// NeitherViewNorPure is a paid mutator transaction binding the contract method 0xa7263000.
//
// Solidity: function NeitherViewNorPure() returns()
func (_PrecompileTest *PrecompileTestTransactor) NeitherViewNorPure(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PrecompileTest.contract.Transact(opts, "NeitherViewNorPure")
}

// NeitherViewNorPure is a paid mutator transaction binding the contract method 0xa7263000.
//
// Solidity: function NeitherViewNorPure() returns()
func (_PrecompileTest *PrecompileTestSession) NeitherViewNorPure() (*types.Transaction, error) {
	return _PrecompileTest.Contract.NeitherViewNorPure(&_PrecompileTest.TransactOpts)
}

// NeitherViewNorPure is a paid mutator transaction binding the contract method 0xa7263000.
//
// Solidity: function NeitherViewNorPure() returns()
func (_PrecompileTest *PrecompileTestTransactorSession) NeitherViewNorPure() (*types.Transaction, error) {
	return _PrecompileTest.Contract.NeitherViewNorPure(&_PrecompileTest.TransactOpts)
}

// Pure is a paid mutator transaction binding the contract method 0xf5ad2fb8.
//
// Solidity: function Pure() returns()
func (_PrecompileTest *PrecompileTestTransactor) Pure(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PrecompileTest.contract.Transact(opts, "Pure")
}

// Pure is a paid mutator transaction binding the contract method 0xf5ad2fb8.
//
// Solidity: function Pure() returns()
func (_PrecompileTest *PrecompileTestSession) Pure() (*types.Transaction, error) {
	return _PrecompileTest.Contract.Pure(&_PrecompileTest.TransactOpts)
}

// Pure is a paid mutator transaction binding the contract method 0xf5ad2fb8.
//
// Solidity: function Pure() returns()
func (_PrecompileTest *PrecompileTestTransactorSession) Pure() (*types.Transaction, error) {
	return _PrecompileTest.Contract.Pure(&_PrecompileTest.TransactOpts)
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

// View is a paid mutator transaction binding the contract method 0x1686f265.
//
// Solidity: function View() returns()
func (_PrecompileTest *PrecompileTestTransactor) View(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PrecompileTest.contract.Transact(opts, "View")
}

// View is a paid mutator transaction binding the contract method 0x1686f265.
//
// Solidity: function View() returns()
func (_PrecompileTest *PrecompileTestSession) View() (*types.Transaction, error) {
	return _PrecompileTest.Contract.View(&_PrecompileTest.TransactOpts)
}

// View is a paid mutator transaction binding the contract method 0x1686f265.
//
// Solidity: function View() returns()
func (_PrecompileTest *PrecompileTestTransactorSession) View() (*types.Transaction, error) {
	return _PrecompileTest.Contract.View(&_PrecompileTest.TransactOpts)
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
