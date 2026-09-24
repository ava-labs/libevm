// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package testprecompile

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

// TestSuiteMetaData contains all meta data concerning the TestSuite contract.
var TestSuiteMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIPrecompile\",\"name\":\"_precompile\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"nonPayableErrorMsg\",\"type\":\"string\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"func\",\"type\":\"string\"}],\"name\":\"Called\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"}],\"name\":\"Echo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"x\",\"type\":\"string\"}],\"name\":\"Echo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"input\",\"type\":\"bytes\"}],\"name\":\"EchoingFallback\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"int256\",\"name\":\"val\",\"type\":\"int256\"}],\"internalType\":\"structIPrecompile.Wrapper\",\"name\":\"x\",\"type\":\"tuple\"}],\"name\":\"Extract\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"},{\"internalType\":\"bytes2\",\"name\":\"y\",\"type\":\"bytes2\"},{\"internalType\":\"address\",\"name\":\"z\",\"type\":\"address\"}],\"name\":\"HashPacked\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"NeitherViewNorPure\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"Pure\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"err\",\"type\":\"bytes\"}],\"name\":\"RevertWith\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"Self\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"Transfer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"View\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60a060405260405161236f38038061236f83398181016040528101906100259190610227565b8173ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff1681525050805f90816100679190610491565b505050610560565b5f604051905090565b5f5ffd5b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6100a982610080565b9050919050565b5f6100ba8261009f565b9050919050565b6100ca816100b0565b81146100d4575f5ffd5b50565b5f815190506100e5816100c1565b92915050565b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b610139826100f3565b810181811067ffffffffffffffff8211171561015857610157610103565b5b80604052505050565b5f61016a61006f565b90506101768282610130565b919050565b5f67ffffffffffffffff82111561019557610194610103565b5b61019e826100f3565b9050602081019050919050565b8281835e5f83830152505050565b5f6101cb6101c68461017b565b610161565b9050828152602081018484840111156101e7576101e66100ef565b5b6101f28482856101ab565b509392505050565b5f82601f83011261020e5761020d6100eb565b5b815161021e8482602086016101b9565b91505092915050565b5f5f6040838503121561023d5761023c610078565b5b5f61024a858286016100d7565b925050602083015167ffffffffffffffff81111561026b5761026a61007c565b5b610277858286016101fa565b9150509250929050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f60028204905060018216806102cf57607f821691505b6020821081036102e2576102e161028b565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026103447fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610309565b61034e8683610309565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61039261038d61038884610366565b61036f565b610366565b9050919050565b5f819050919050565b6103ab83610378565b6103bf6103b782610399565b848454610315565b825550505050565b5f5f905090565b6103d66103c7565b6103e18184846103a2565b505050565b5b81811015610404576103f95f826103ce565b6001810190506103e7565b5050565b601f8211156104495761041a816102e8565b610423846102fa565b81016020851015610432578190505b61044661043e856102fa565b8301826103e6565b50505b505050565b5f82821c905092915050565b5f6104695f198460080261044e565b1980831691505092915050565b5f610481838361045a565b9150826002028217905092915050565b61049a82610281565b67ffffffffffffffff8111156104b3576104b2610103565b5b6104bd82546102b8565b6104c8828285610408565b5f60209050601f8311600181146104f9575f84156104e7578287015190505b6104f18582610476565b865550610558565b601f198416610507866102e8565b5f5b8281101561052e57848901518255600182019150602085019450602081019050610509565b8683101561054b5784890151610547601f89168261045a565b8355505b6001600288020188555050505b505050505050565b608051611db16105be5f395f81816101d2015281816102b101528181610454015281816105b8015281816106f2015281816107e0015281816108170152818161092a01528181610a3501528181610b870152610d340152611db15ff3fe608060405234801561000f575f5ffd5b50600436106100a7575f3560e01c8063ad8108a41161006f578063ad8108a414610101578063b4a818081461011d578063c62c692f14610139578063d7cc1f3714610143578063db84d7c01461015f578063f5ad2fb81461017b576100a7565b80631686f265146100ab57806334d6d9be146100b5578063406dade3146100d1578063a7263000146100db578063a93cbd97146100e5575b5f5ffd5b6100b3610185565b005b6100cf60048036038101906100ca9190610e51565b6101cf565b005b6100d96102ae565b005b6100e3610406565b005b6100ff60048036038101906100fa9190610fb8565b610450565b005b61011b6004803603810190610116919061106f565b6105b6565b005b61013760048036038101906101329190610fb8565b610698565b005b6101416107de565b005b61015d60048036038101906101589190611149565b6108fd565b005b61017960048036038101906101749190611237565b610a0c565b005b610183610b3a565b005b610198631686f26560e01b60015f610b83565b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516101c5906112d8565b60405180910390a1565b807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166334d6d9be836040518263ffffffff1660e01b81526004016102299190611305565b602060405180830381865afa158015610244573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102689190611332565b146102765761027561135d565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516102a3906113d4565b60405180910390a150565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16631b7fb4b0602a6040518263ffffffff1660e01b815260040160206040518083038185885af115801561031b573d5f5f3e3d5ffd5b50505050506040513d601f19601f820116820180604052508101906103409190611332565b9050602a81146103535761035261135d565b5b5f636fb1b0e960e01b9050610377815f60405180602001604052805f815250610cc9565b5f5f60405160200161038991906114eb565b60405160208183030381529060405290506103a682600183610cc9565b6103b9631686f26560e01b600183610cc9565b6103cc63f5ad2fb860e01b600183610cc9565b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516103f99061154b565b60405180910390a1505050565b61041963a726300060e01b600180610b83565b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a604051610446906115b3565b60405180910390a1565b5f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663a93cbd9760e01b846040516024016104a29190611631565b604051602081830303815290604052907bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff838183161783525050505060405161050c919061168b565b5f604051808303815f865af19150503d805f8114610545576040519150601f19603f3d011682016040523d82523d5f602084013e61054a565b606091505b5091509150811561055e5761055d61135d565b5b828051906020012081805190602001201461057c5761057b61135d565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516105a9906116eb565b60405180910390a1505050565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663ad8108a4826040518263ffffffff1660e01b815260040161060f9190611732565b602060405180830381865afa15801561062a573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061064e919061175f565b815f0151146106605761065f61135d565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a60405161068d906117d4565b60405180910390a150565b5f5f826040516024016106ab9190611631565b6040516020818303038152906040529060e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff838183161783525050505090505f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1683604051610735919061168b565b5f60405180830381855afa9150503d805f811461076d576040519150601f19603f3d011682016040523d82523d5f602084013e610772565b606091505b5091509150816107855761078461135d565b5b82805190602001208180519060200120146107a3576107a261135d565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516107d09061183c565b60405180910390a150505050565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663c62c692f6040518163ffffffff1660e01b8152600401602060405180830381865afa15801561087e573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906108a2919061186e565b73ffffffffffffffffffffffffffffffffffffffff16146108c6576108c561135d565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516108f3906118e3565b60405180910390a1565b82828260405160200161091293929190611986565b604051602081830303815290604052805190602001207f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663d7cc1f378585856040518463ffffffff1660e01b8152600401610985939291906119e0565b602060405180830381865afa1580156109a0573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906109c49190611a48565b146109d2576109d161135d565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a6040516109ff90611abd565b60405180910390a1505050565b80604051602001610a1d9190611b15565b604051602081830303815290604052805190602001207f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663db84d7c0836040518263ffffffff1660e01b8152600401610a8c9190611b63565b5f60405180830381865afa158015610aa6573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f82011682018060405250810190610ace9190611bf1565b604051602001610ade9190611b15565b6040516020818303038152906040528051906020012014610b0257610b0161135d565b5b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a604051610b2f90611c82565b60405180910390a150565b610b4c63f5ad2fb860e01b5f5f610b83565b7f3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a604051610b7990611cea565b60405180910390a1565b5f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1685604051602401604051602081830303815290604052907bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050604051610c2d919061168b565b5f604051808303815f865af19150503d805f8114610c66576040519150601f19603f3d011682016040523d82523d5f602084013e610c6b565b606091505b509150915081610c7e57610c7d61135d565b5b5f5f82806020019051810190610c949190611d3d565b9150915085151582151514610cac57610cab61135d565b5b84151581151514610cc057610cbf61135d565b5b50505050505050565b5f83604051602401604051602081830303815290604052907bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff838183161783525050505090505f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff168584604051610d78919061168b565b5f6040518083038185875af1925050503d805f8114610db2576040519150601f19603f3d011682016040523d82523d5f602084013e610db7565b606091505b50915091505f845103610dd75781610dd257610dd161135d565b5b610e05565b8115610de657610de561135d565b5b8380519060200120818051906020012014610e0457610e0361135d565b5b5b505050505050565b5f604051905090565b5f5ffd5b5f5ffd5b5f819050919050565b610e3081610e1e565b8114610e3a575f5ffd5b50565b5f81359050610e4b81610e27565b92915050565b5f60208284031215610e6657610e65610e16565b5b5f610e7384828501610e3d565b91505092915050565b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b610eca82610e84565b810181811067ffffffffffffffff82111715610ee957610ee8610e94565b5b80604052505050565b5f610efb610e0d565b9050610f078282610ec1565b919050565b5f67ffffffffffffffff821115610f2657610f25610e94565b5b610f2f82610e84565b9050602081019050919050565b828183375f83830152505050565b5f610f5c610f5784610f0c565b610ef2565b905082815260208101848484011115610f7857610f77610e80565b5b610f83848285610f3c565b509392505050565b5f82601f830112610f9f57610f9e610e7c565b5b8135610faf848260208601610f4a565b91505092915050565b5f60208284031215610fcd57610fcc610e16565b5b5f82013567ffffffffffffffff811115610fea57610fe9610e1a565b5b610ff684828501610f8b565b91505092915050565b5f5ffd5b5f819050919050565b61101581611003565b811461101f575f5ffd5b50565b5f813590506110308161100c565b92915050565b5f6020828403121561104b5761104a610fff565b5b6110556020610ef2565b90505f61106484828501611022565b5f8301525092915050565b5f6020828403121561108457611083610e16565b5b5f61109184828501611036565b91505092915050565b5f7fffff00000000000000000000000000000000000000000000000000000000000082169050919050565b6110ce8161109a565b81146110d8575f5ffd5b50565b5f813590506110e9816110c5565b92915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f611118826110ef565b9050919050565b6111288161110e565b8114611132575f5ffd5b50565b5f813590506111438161111f565b92915050565b5f5f5f606084860312156111605761115f610e16565b5b5f61116d86828701610e3d565b935050602061117e868287016110db565b925050604061118f86828701611135565b9150509250925092565b5f67ffffffffffffffff8211156111b3576111b2610e94565b5b6111bc82610e84565b9050602081019050919050565b5f6111db6111d684611199565b610ef2565b9050828152602081018484840111156111f7576111f6610e80565b5b611202848285610f3c565b509392505050565b5f82601f83011261121e5761121d610e7c565b5b813561122e8482602086016111c9565b91505092915050565b5f6020828403121561124c5761124b610e16565b5b5f82013567ffffffffffffffff81111561126957611268610e1a565b5b6112758482850161120a565b91505092915050565b5f82825260208201905092915050565b7f56696577282900000000000000000000000000000000000000000000000000005f82015250565b5f6112c260068361127e565b91506112cd8261128e565b602082019050919050565b5f6020820190508181035f8301526112ef816112b6565b9050919050565b6112ff81610e1e565b82525050565b5f6020820190506113185f8301846112f6565b92915050565b5f8151905061132c81610e27565b92915050565b5f6020828403121561134757611346610e16565b5b5f6113548482850161131e565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52600160045260245ffd5b7f4563686f2875696e7432353629000000000000000000000000000000000000005f82015250565b5f6113be600d8361127e565b91506113c98261138a565b602082019050919050565b5f6020820190508181035f8301526113eb816113b2565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061143657607f821691505b602082108103611449576114486113f2565b5b50919050565b5f81905092915050565b5f819050815f5260205f209050919050565b5f81546114778161141f565b611481818661144f565b9450600182165f811461149b57600181146114b0576114e2565b60ff19831686528115158202860193506114e2565b6114b985611459565b5f5b838110156114da578154818901526001820191506020810190506114bb565b838801955050505b50505092915050565b5f6114f6828461146b565b915081905092915050565b7f5472616e736665722829000000000000000000000000000000000000000000005f82015250565b5f611535600a8361127e565b915061154082611501565b602082019050919050565b5f6020820190508181035f83015261156281611529565b9050919050565b7f4e656974686572566965774e6f725075726528290000000000000000000000005f82015250565b5f61159d60148361127e565b91506115a882611569565b602082019050919050565b5f6020820190508181035f8301526115ca81611591565b9050919050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f611603826115d1565b61160d81856115db565b935061161d8185602086016115eb565b61162681610e84565b840191505092915050565b5f6020820190508181035f83015261164981846115f9565b905092915050565b5f81905092915050565b5f611665826115d1565b61166f8185611651565b935061167f8185602086016115eb565b80840191505092915050565b5f611696828461165b565b915081905092915050565b7f52657665727457697468282e2e2e2900000000000000000000000000000000005f82015250565b5f6116d5600f8361127e565b91506116e0826116a1565b602082019050919050565b5f6020820190508181035f830152611702816116c9565b9050919050565b61171281611003565b82525050565b602082015f82015161172c5f850182611709565b50505050565b5f6020820190506117455f830184611718565b92915050565b5f815190506117598161100c565b92915050565b5f6020828403121561177457611773610e16565b5b5f6117818482850161174b565b91505092915050565b7f45787472616374282e2e2e2900000000000000000000000000000000000000005f82015250565b5f6117be600c8361127e565b91506117c98261178a565b602082019050919050565b5f6020820190508181035f8301526117eb816117b2565b9050919050565b7f4563686f696e6746616c6c6261636b282e2e2e290000000000000000000000005f82015250565b5f61182660148361127e565b9150611831826117f2565b602082019050919050565b5f6020820190508181035f8301526118538161181a565b9050919050565b5f815190506118688161111f565b92915050565b5f6020828403121561188357611882610e16565b5b5f6118908482850161185a565b91505092915050565b7f53656c66282900000000000000000000000000000000000000000000000000005f82015250565b5f6118cd60068361127e565b91506118d882611899565b602082019050919050565b5f6020820190508181035f8301526118fa816118c1565b9050919050565b5f819050919050565b61191b61191682610e1e565b611901565b82525050565b5f819050919050565b61193b6119368261109a565b611921565b82525050565b5f8160601b9050919050565b5f61195782611941565b9050919050565b5f6119688261194d565b9050919050565b61198061197b8261110e565b61195e565b82525050565b5f611991828661190a565b6020820191506119a1828561192a565b6002820191506119b1828461196f565b601482019150819050949350505050565b6119cb8161109a565b82525050565b6119da8161110e565b82525050565b5f6060820190506119f35f8301866112f6565b611a0060208301856119c2565b611a0d60408301846119d1565b949350505050565b5f819050919050565b611a2781611a15565b8114611a31575f5ffd5b50565b5f81519050611a4281611a1e565b92915050565b5f60208284031215611a5d57611a5c610e16565b5b5f611a6a84828501611a34565b91505092915050565b7f486173685061636b6564282e2e2e2900000000000000000000000000000000005f82015250565b5f611aa7600f8361127e565b9150611ab282611a73565b602082019050919050565b5f6020820190508181035f830152611ad481611a9b565b9050919050565b5f81519050919050565b5f611aef82611adb565b611af9818561144f565b9350611b098185602086016115eb565b80840191505092915050565b5f611b208284611ae5565b915081905092915050565b5f611b3582611adb565b611b3f818561127e565b9350611b4f8185602086016115eb565b611b5881610e84565b840191505092915050565b5f6020820190508181035f830152611b7b8184611b2b565b905092915050565b5f611b95611b9084611199565b610ef2565b905082815260208101848484011115611bb157611bb0610e80565b5b611bbc8482856115eb565b509392505050565b5f82601f830112611bd857611bd7610e7c565b5b8151611be8848260208601611b83565b91505092915050565b5f60208284031215611c0657611c05610e16565b5b5f82015167ffffffffffffffff811115611c2357611c22610e1a565b5b611c2f84828501611bc4565b91505092915050565b7f4563686f28737472696e672900000000000000000000000000000000000000005f82015250565b5f611c6c600c8361127e565b9150611c7782611c38565b602082019050919050565b5f6020820190508181035f830152611c9981611c60565b9050919050565b7f50757265282900000000000000000000000000000000000000000000000000005f82015250565b5f611cd460068361127e565b9150611cdf82611ca0565b602082019050919050565b5f6020820190508181035f830152611d0181611cc8565b9050919050565b5f8115159050919050565b611d1c81611d08565b8114611d26575f5ffd5b50565b5f81519050611d3781611d13565b92915050565b5f5f60408385031215611d5357611d52610e16565b5b5f611d6085828601611d29565b9250506020611d7185828601611d29565b915050925092905056fea2646970667358221220f0b141f0869405a047b6257eeea3f5957b650b4c10613fbb449d8778f8d6986764736f6c634300081c0033",
}

// TestSuiteABI is the input ABI used to generate the binding from.
// Deprecated: Use TestSuiteMetaData.ABI instead.
var TestSuiteABI = TestSuiteMetaData.ABI

// TestSuiteBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use TestSuiteMetaData.Bin instead.
var TestSuiteBin = TestSuiteMetaData.Bin

// DeployTestSuite deploys a new Ethereum contract, binding an instance of TestSuite to it.
func DeployTestSuite(auth *bind.TransactOpts, backend bind.ContractBackend, _precompile common.Address, nonPayableErrorMsg string) (common.Address, *types.Transaction, *TestSuite, error) {
	parsed, err := TestSuiteMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(TestSuiteBin), backend, _precompile, nonPayableErrorMsg)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &TestSuite{TestSuiteCaller: TestSuiteCaller{contract: contract}, TestSuiteTransactor: TestSuiteTransactor{contract: contract}, TestSuiteFilterer: TestSuiteFilterer{contract: contract}}, nil
}

// TestSuite is an auto generated Go binding around an Ethereum contract.
type TestSuite struct {
	TestSuiteCaller     // Read-only binding to the contract
	TestSuiteTransactor // Write-only binding to the contract
	TestSuiteFilterer   // Log filterer for contract events
}

// TestSuiteCaller is an auto generated read-only Go binding around an Ethereum contract.
type TestSuiteCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestSuiteTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TestSuiteTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestSuiteFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TestSuiteFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestSuiteSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TestSuiteSession struct {
	Contract     *TestSuite        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TestSuiteCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TestSuiteCallerSession struct {
	Contract *TestSuiteCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// TestSuiteTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TestSuiteTransactorSession struct {
	Contract     *TestSuiteTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// TestSuiteRaw is an auto generated low-level Go binding around an Ethereum contract.
type TestSuiteRaw struct {
	Contract *TestSuite // Generic contract binding to access the raw methods on
}

// TestSuiteCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TestSuiteCallerRaw struct {
	Contract *TestSuiteCaller // Generic read-only contract binding to access the raw methods on
}

// TestSuiteTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TestSuiteTransactorRaw struct {
	Contract *TestSuiteTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTestSuite creates a new instance of TestSuite, bound to a specific deployed contract.
func NewTestSuite(address common.Address, backend bind.ContractBackend) (*TestSuite, error) {
	contract, err := bindTestSuite(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TestSuite{TestSuiteCaller: TestSuiteCaller{contract: contract}, TestSuiteTransactor: TestSuiteTransactor{contract: contract}, TestSuiteFilterer: TestSuiteFilterer{contract: contract}}, nil
}

// NewTestSuiteCaller creates a new read-only instance of TestSuite, bound to a specific deployed contract.
func NewTestSuiteCaller(address common.Address, caller bind.ContractCaller) (*TestSuiteCaller, error) {
	contract, err := bindTestSuite(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestSuiteCaller{contract: contract}, nil
}

// NewTestSuiteTransactor creates a new write-only instance of TestSuite, bound to a specific deployed contract.
func NewTestSuiteTransactor(address common.Address, transactor bind.ContractTransactor) (*TestSuiteTransactor, error) {
	contract, err := bindTestSuite(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestSuiteTransactor{contract: contract}, nil
}

// NewTestSuiteFilterer creates a new log filterer instance of TestSuite, bound to a specific deployed contract.
func NewTestSuiteFilterer(address common.Address, filterer bind.ContractFilterer) (*TestSuiteFilterer, error) {
	contract, err := bindTestSuite(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestSuiteFilterer{contract: contract}, nil
}

// bindTestSuite binds a generic wrapper to an already deployed contract.
func bindTestSuite(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TestSuiteMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestSuite *TestSuiteRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestSuite.Contract.TestSuiteCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestSuite *TestSuiteRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestSuite.Contract.TestSuiteTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestSuite *TestSuiteRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestSuite.Contract.TestSuiteTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestSuite *TestSuiteCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestSuite.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestSuite *TestSuiteTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestSuite.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestSuite *TestSuiteTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestSuite.Contract.contract.Transact(opts, method, params...)
}

// Echo is a paid mutator transaction binding the contract method 0x34d6d9be.
//
// Solidity: function Echo(uint256 x) returns()
func (_TestSuite *TestSuiteTransactor) Echo(opts *bind.TransactOpts, x *big.Int) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "Echo", x)
}

// Echo is a paid mutator transaction binding the contract method 0x34d6d9be.
//
// Solidity: function Echo(uint256 x) returns()
func (_TestSuite *TestSuiteSession) Echo(x *big.Int) (*types.Transaction, error) {
	return _TestSuite.Contract.Echo(&_TestSuite.TransactOpts, x)
}

// Echo is a paid mutator transaction binding the contract method 0x34d6d9be.
//
// Solidity: function Echo(uint256 x) returns()
func (_TestSuite *TestSuiteTransactorSession) Echo(x *big.Int) (*types.Transaction, error) {
	return _TestSuite.Contract.Echo(&_TestSuite.TransactOpts, x)
}

// Echo0 is a paid mutator transaction binding the contract method 0xdb84d7c0.
//
// Solidity: function Echo(string x) returns()
func (_TestSuite *TestSuiteTransactor) Echo0(opts *bind.TransactOpts, x string) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "Echo0", x)
}

// Echo0 is a paid mutator transaction binding the contract method 0xdb84d7c0.
//
// Solidity: function Echo(string x) returns()
func (_TestSuite *TestSuiteSession) Echo0(x string) (*types.Transaction, error) {
	return _TestSuite.Contract.Echo0(&_TestSuite.TransactOpts, x)
}

// Echo0 is a paid mutator transaction binding the contract method 0xdb84d7c0.
//
// Solidity: function Echo(string x) returns()
func (_TestSuite *TestSuiteTransactorSession) Echo0(x string) (*types.Transaction, error) {
	return _TestSuite.Contract.Echo0(&_TestSuite.TransactOpts, x)
}

// EchoingFallback is a paid mutator transaction binding the contract method 0xb4a81808.
//
// Solidity: function EchoingFallback(bytes input) returns()
func (_TestSuite *TestSuiteTransactor) EchoingFallback(opts *bind.TransactOpts, input []byte) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "EchoingFallback", input)
}

// EchoingFallback is a paid mutator transaction binding the contract method 0xb4a81808.
//
// Solidity: function EchoingFallback(bytes input) returns()
func (_TestSuite *TestSuiteSession) EchoingFallback(input []byte) (*types.Transaction, error) {
	return _TestSuite.Contract.EchoingFallback(&_TestSuite.TransactOpts, input)
}

// EchoingFallback is a paid mutator transaction binding the contract method 0xb4a81808.
//
// Solidity: function EchoingFallback(bytes input) returns()
func (_TestSuite *TestSuiteTransactorSession) EchoingFallback(input []byte) (*types.Transaction, error) {
	return _TestSuite.Contract.EchoingFallback(&_TestSuite.TransactOpts, input)
}

// Extract is a paid mutator transaction binding the contract method 0xad8108a4.
//
// Solidity: function Extract((int256) x) returns()
func (_TestSuite *TestSuiteTransactor) Extract(opts *bind.TransactOpts, x IPrecompileWrapper) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "Extract", x)
}

// Extract is a paid mutator transaction binding the contract method 0xad8108a4.
//
// Solidity: function Extract((int256) x) returns()
func (_TestSuite *TestSuiteSession) Extract(x IPrecompileWrapper) (*types.Transaction, error) {
	return _TestSuite.Contract.Extract(&_TestSuite.TransactOpts, x)
}

// Extract is a paid mutator transaction binding the contract method 0xad8108a4.
//
// Solidity: function Extract((int256) x) returns()
func (_TestSuite *TestSuiteTransactorSession) Extract(x IPrecompileWrapper) (*types.Transaction, error) {
	return _TestSuite.Contract.Extract(&_TestSuite.TransactOpts, x)
}

// HashPacked is a paid mutator transaction binding the contract method 0xd7cc1f37.
//
// Solidity: function HashPacked(uint256 x, bytes2 y, address z) returns()
func (_TestSuite *TestSuiteTransactor) HashPacked(opts *bind.TransactOpts, x *big.Int, y [2]byte, z common.Address) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "HashPacked", x, y, z)
}

// HashPacked is a paid mutator transaction binding the contract method 0xd7cc1f37.
//
// Solidity: function HashPacked(uint256 x, bytes2 y, address z) returns()
func (_TestSuite *TestSuiteSession) HashPacked(x *big.Int, y [2]byte, z common.Address) (*types.Transaction, error) {
	return _TestSuite.Contract.HashPacked(&_TestSuite.TransactOpts, x, y, z)
}

// HashPacked is a paid mutator transaction binding the contract method 0xd7cc1f37.
//
// Solidity: function HashPacked(uint256 x, bytes2 y, address z) returns()
func (_TestSuite *TestSuiteTransactorSession) HashPacked(x *big.Int, y [2]byte, z common.Address) (*types.Transaction, error) {
	return _TestSuite.Contract.HashPacked(&_TestSuite.TransactOpts, x, y, z)
}

// NeitherViewNorPure is a paid mutator transaction binding the contract method 0xa7263000.
//
// Solidity: function NeitherViewNorPure() returns()
func (_TestSuite *TestSuiteTransactor) NeitherViewNorPure(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "NeitherViewNorPure")
}

// NeitherViewNorPure is a paid mutator transaction binding the contract method 0xa7263000.
//
// Solidity: function NeitherViewNorPure() returns()
func (_TestSuite *TestSuiteSession) NeitherViewNorPure() (*types.Transaction, error) {
	return _TestSuite.Contract.NeitherViewNorPure(&_TestSuite.TransactOpts)
}

// NeitherViewNorPure is a paid mutator transaction binding the contract method 0xa7263000.
//
// Solidity: function NeitherViewNorPure() returns()
func (_TestSuite *TestSuiteTransactorSession) NeitherViewNorPure() (*types.Transaction, error) {
	return _TestSuite.Contract.NeitherViewNorPure(&_TestSuite.TransactOpts)
}

// Pure is a paid mutator transaction binding the contract method 0xf5ad2fb8.
//
// Solidity: function Pure() returns()
func (_TestSuite *TestSuiteTransactor) Pure(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "Pure")
}

// Pure is a paid mutator transaction binding the contract method 0xf5ad2fb8.
//
// Solidity: function Pure() returns()
func (_TestSuite *TestSuiteSession) Pure() (*types.Transaction, error) {
	return _TestSuite.Contract.Pure(&_TestSuite.TransactOpts)
}

// Pure is a paid mutator transaction binding the contract method 0xf5ad2fb8.
//
// Solidity: function Pure() returns()
func (_TestSuite *TestSuiteTransactorSession) Pure() (*types.Transaction, error) {
	return _TestSuite.Contract.Pure(&_TestSuite.TransactOpts)
}

// RevertWith is a paid mutator transaction binding the contract method 0xa93cbd97.
//
// Solidity: function RevertWith(bytes err) returns()
func (_TestSuite *TestSuiteTransactor) RevertWith(opts *bind.TransactOpts, err []byte) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "RevertWith", err)
}

// RevertWith is a paid mutator transaction binding the contract method 0xa93cbd97.
//
// Solidity: function RevertWith(bytes err) returns()
func (_TestSuite *TestSuiteSession) RevertWith(err []byte) (*types.Transaction, error) {
	return _TestSuite.Contract.RevertWith(&_TestSuite.TransactOpts, err)
}

// RevertWith is a paid mutator transaction binding the contract method 0xa93cbd97.
//
// Solidity: function RevertWith(bytes err) returns()
func (_TestSuite *TestSuiteTransactorSession) RevertWith(err []byte) (*types.Transaction, error) {
	return _TestSuite.Contract.RevertWith(&_TestSuite.TransactOpts, err)
}

// Self is a paid mutator transaction binding the contract method 0xc62c692f.
//
// Solidity: function Self() returns()
func (_TestSuite *TestSuiteTransactor) Self(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "Self")
}

// Self is a paid mutator transaction binding the contract method 0xc62c692f.
//
// Solidity: function Self() returns()
func (_TestSuite *TestSuiteSession) Self() (*types.Transaction, error) {
	return _TestSuite.Contract.Self(&_TestSuite.TransactOpts)
}

// Self is a paid mutator transaction binding the contract method 0xc62c692f.
//
// Solidity: function Self() returns()
func (_TestSuite *TestSuiteTransactorSession) Self() (*types.Transaction, error) {
	return _TestSuite.Contract.Self(&_TestSuite.TransactOpts)
}

// Transfer is a paid mutator transaction binding the contract method 0x406dade3.
//
// Solidity: function Transfer() returns()
func (_TestSuite *TestSuiteTransactor) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "Transfer")
}

// Transfer is a paid mutator transaction binding the contract method 0x406dade3.
//
// Solidity: function Transfer() returns()
func (_TestSuite *TestSuiteSession) Transfer() (*types.Transaction, error) {
	return _TestSuite.Contract.Transfer(&_TestSuite.TransactOpts)
}

// Transfer is a paid mutator transaction binding the contract method 0x406dade3.
//
// Solidity: function Transfer() returns()
func (_TestSuite *TestSuiteTransactorSession) Transfer() (*types.Transaction, error) {
	return _TestSuite.Contract.Transfer(&_TestSuite.TransactOpts)
}

// View is a paid mutator transaction binding the contract method 0x1686f265.
//
// Solidity: function View() returns()
func (_TestSuite *TestSuiteTransactor) View(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestSuite.contract.Transact(opts, "View")
}

// View is a paid mutator transaction binding the contract method 0x1686f265.
//
// Solidity: function View() returns()
func (_TestSuite *TestSuiteSession) View() (*types.Transaction, error) {
	return _TestSuite.Contract.View(&_TestSuite.TransactOpts)
}

// View is a paid mutator transaction binding the contract method 0x1686f265.
//
// Solidity: function View() returns()
func (_TestSuite *TestSuiteTransactorSession) View() (*types.Transaction, error) {
	return _TestSuite.Contract.View(&_TestSuite.TransactOpts)
}

// TestSuiteCalledIterator is returned from FilterCalled and is used to iterate over the raw logs and unpacked data for Called events raised by the TestSuite contract.
type TestSuiteCalledIterator struct {
	Event *TestSuiteCalled // Event containing the contract specifics and raw log

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
func (it *TestSuiteCalledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestSuiteCalled)
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
		it.Event = new(TestSuiteCalled)
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
func (it *TestSuiteCalledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestSuiteCalledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestSuiteCalled represents a Called event raised by the TestSuite contract.
type TestSuiteCalled struct {
	Arg0 string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterCalled is a free log retrieval operation binding the contract event 0x3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a.
//
// Solidity: event Called(string func)
func (_TestSuite *TestSuiteFilterer) FilterCalled(opts *bind.FilterOpts) (*TestSuiteCalledIterator, error) {

	logs, sub, err := _TestSuite.contract.FilterLogs(opts, "Called")
	if err != nil {
		return nil, err
	}
	return &TestSuiteCalledIterator{contract: _TestSuite.contract, event: "Called", logs: logs, sub: sub}, nil
}

// WatchCalled is a free log subscription operation binding the contract event 0x3fd5724095fcb65d3b7e6d79ac130db7e9dbf891d59bb0da292d471bc40d564a.
//
// Solidity: event Called(string func)
func (_TestSuite *TestSuiteFilterer) WatchCalled(opts *bind.WatchOpts, sink chan<- *TestSuiteCalled) (event.Subscription, error) {

	logs, sub, err := _TestSuite.contract.WatchLogs(opts, "Called")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestSuiteCalled)
				if err := _TestSuite.contract.UnpackLog(event, "Called", log); err != nil {
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
func (_TestSuite *TestSuiteFilterer) ParseCalled(log types.Log) (*TestSuiteCalled, error) {
	event := new(TestSuiteCalled)
	if err := _TestSuite.contract.UnpackLog(event, "Called", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
