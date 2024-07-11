// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package polygondatacommittee

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
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

// PolygondatacommitteeMetaData contains all meta data concerning the Polygondatacommittee contract.
var PolygondatacommitteeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"CommitteeAddressDoesNotExist\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EmptyURLNotAllowed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TooManyRequiredSignatures\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UnexpectedAddrsAndSignaturesSize\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UnexpectedAddrsBytesLength\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UnexpectedCommitteeHash\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WrongAddrOrder\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"committeeHash\",\"type\":\"bytes32\"}],\"name\":\"CommitteeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"committeeHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAmountOfMembers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProcotolName\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"members\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"url\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"requiredAmountOfSignatures\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_requiredAmountOfSignatures\",\"type\":\"uint256\"},{\"internalType\":\"string[]\",\"name\":\"urls\",\"type\":\"string[]\"},{\"internalType\":\"bytes\",\"name\":\"addrsBytes\",\"type\":\"bytes\"}],\"name\":\"setupCommittee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"signedHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signaturesAndAddrs\",\"type\":\"bytes\"}],\"name\":\"verifyMessage\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b50600436106100c4575f3560e01c8063715018a61161007d578063dce1e2b611610058578063dce1e2b614610172578063e4f171201461017a578063f2fde38b146101b9575f80fd5b8063715018a61461013a5780638129fc1c146101425780638da5cb5b1461014a575f80fd5b80635daf08ca116100ad5780635daf08ca146100f0578063609d45441461011a5780636beedd3914610131575f80fd5b8063078fba2a146100c85780633b51be4b146100dd575b5f80fd5b6100db6100d6366004610fab565b6101cc565b005b6100db6100eb36600461104d565b6104c8565b6101036100fe366004611095565b61070a565b60405161011192919061110d565b60405180910390f35b61012360665481565b604051908152602001610111565b61012360655481565b6100db6107d5565b6100db6107e8565b60335460405173ffffffffffffffffffffffffffffffffffffffff9091168152602001610111565b606754610123565b604080518082018252601981527f44617461417661696c6162696c697479436f6d6d697474656500000000000000602082015290516101119190611144565b6100db6101c736600461115d565b610979565b6101d4610a2d565b828581101561020f576040517f2e7dcd6e00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b61021a6014826111bd565b8214610252576040517f2ab6a12900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b61025d60675f610ec3565b5f805b8281101561046c575f6102746014836111bd565b90505f8682876102856014836111d4565b92610292939291906111e7565b61029b9161120e565b60601c90508888848181106102b2576102b2611256565b90506020028101906102c49190611283565b90505f036102fe576040517fb54b70e400000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b8073ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1610610363576040517fd53cfbe000000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b606760405180604001604052808b8b8781811061038257610382611256565b90506020028101906103949190611283565b8080601f0160208091040260200160405190810160405280939291908181526020018383808284375f92018290525093855250505073ffffffffffffffffffffffffffffffffffffffff8516602092830152835460018101855593815220815191926002020190819061040790826113b0565b5060209190910151600190910180547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff9092169190911790559250819050610464816114c8565b915050610260565b50838360405161047d9291906114ff565b6040519081900381206066819055606589905581527f831403fd381b3e6ac875d912ec2eee0e0203d0d29f7b3e0c96fc8f582d6db6579060200160405180910390a150505050505050565b6065545f6104d78260416111bd565b9050808310806104fb575060146104ee828561150e565b6104f8919061154e565b15155b15610532576040517f6b8eec4600000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b606654610541848381886111e7565b60405161054f9291906114ff565b60405180910390201461058e576040517f6b156b2800000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f80601461059c848761150e565b6105a69190611561565b90505f5b84811015610700575f6105be6041836111bd565b90505f6106198a8a848b6105d36041836111d4565b926105e0939291906111e7565b8080601f0160208091040260200160405190810160405280939291908181526020018383808284375f92019190915250610aae92505050565b90505f855b858110156106b2575f6106326014836111bd565b61063c908a6111d4565b90505f8c828d61064d6014836111d4565b9261065a939291906111e7565b6106639161120e565b60601c905073ffffffffffffffffffffffffffffffffffffffff8516810361069d576106908360016111d4565b98506001935050506106b2565b505080806106aa906114c8565b91505061061e565b50806106ea576040517fe12afaf500000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b50505080806106f8906114c8565b9150506105aa565b5050505050505050565b60678181548110610719575f80fd5b905f5260205f2090600202015f91509050805f01805461073890611311565b80601f016020809104026020016040519081016040528092919081815260200182805461076490611311565b80156107af5780601f10610786576101008083540402835291602001916107af565b820191905f5260205f20905b81548152906001019060200180831161079257829003601f168201915b5050506001909301549192505073ffffffffffffffffffffffffffffffffffffffff1682565b6107dd610a2d565b6107e65f610ad2565b565b5f54610100900460ff161580801561080657505f54600160ff909116105b8061081f5750303b15801561081f57505f5460ff166001145b6108b0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201527f647920696e697469616c697a656400000000000000000000000000000000000060648201526084015b60405180910390fd5b5f80547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00166001179055801561090c575f80547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff166101001790555b610914610b48565b8015610976575f80547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff169055604051600181527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a15b50565b610981610a2d565b73ffffffffffffffffffffffffffffffffffffffff8116610a24576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f646472657373000000000000000000000000000000000000000000000000000060648201526084016108a7565b61097681610ad2565b60335473ffffffffffffffffffffffffffffffffffffffff1633146107e6576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e657260448201526064016108a7565b5f805f610abb8585610be7565b91509150610ac881610c29565b5090505b92915050565b6033805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0905f90a35050565b5f54610100900460ff16610bde576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201527f6e697469616c697a696e6700000000000000000000000000000000000000000060648201526084016108a7565b6107e633610ad2565b5f808251604103610c1b576020830151604084015160608501515f1a610c0f87828585610ddb565b94509450505050610c22565b505f905060025b9250929050565b5f816004811115610c3c57610c3c611574565b03610c445750565b6001816004811115610c5857610c58611574565b03610cbf576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601860248201527f45434453413a20696e76616c6964207369676e6174757265000000000000000060448201526064016108a7565b6002816004811115610cd357610cd3611574565b03610d3a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f45434453413a20696e76616c6964207369676e6174757265206c656e6774680060448201526064016108a7565b6003816004811115610d4e57610d4e611574565b03610976576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f45434453413a20696e76616c6964207369676e6174757265202773272076616c60448201527f756500000000000000000000000000000000000000000000000000000000000060648201526084016108a7565b5f807f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0831115610e1057505f90506003610eba565b604080515f8082526020820180845289905260ff881692820192909252606081018690526080810185905260019060a0016020604051602081039080840390855afa158015610e61573d5f803e3d5ffd5b50506040517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0015191505073ffffffffffffffffffffffffffffffffffffffff8116610eb4575f60019250925050610eba565b91505f90505b94509492505050565b5080545f8255600202905f5260205f209081019061097691905b80821115610f23575f610ef08282610f27565b506001810180547fffffffffffffffffffffffff0000000000000000000000000000000000000000169055600201610edd565b5090565b508054610f3390611311565b5f825580601f10610f42575050565b601f0160209004905f5260205f209081019061097691905b80821115610f23575f8155600101610f5a565b5f8083601f840112610f7d575f80fd5b50813567ffffffffffffffff811115610f94575f80fd5b602083019150836020828501011115610c22575f80fd5b5f805f805f60608688031215610fbf575f80fd5b85359450602086013567ffffffffffffffff80821115610fdd575f80fd5b818801915088601f830112610ff0575f80fd5b813581811115610ffe575f80fd5b8960208260051b8501011115611012575f80fd5b60208301965080955050604088013591508082111561102f575f80fd5b5061103c88828901610f6d565b969995985093965092949392505050565b5f805f6040848603121561105f575f80fd5b83359250602084013567ffffffffffffffff81111561107c575f80fd5b61108886828701610f6d565b9497909650939450505050565b5f602082840312156110a5575f80fd5b5035919050565b5f81518084525f5b818110156110d0576020818501810151868301820152016110b4565b505f6020828601015260207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f83011685010191505092915050565b604081525f61111f60408301856110ac565b905073ffffffffffffffffffffffffffffffffffffffff831660208301529392505050565b602081525f61115660208301846110ac565b9392505050565b5f6020828403121561116d575f80fd5b813573ffffffffffffffffffffffffffffffffffffffff81168114611156575f80fd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b8082028115828204841417610acc57610acc611190565b80820180821115610acc57610acc611190565b5f80858511156111f5575f80fd5b83861115611201575f80fd5b5050820193919092039150565b7fffffffffffffffffffffffffffffffffffffffff000000000000000000000000813581811691601485101561124e5780818660140360031b1b83161692505b505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f8083357fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe18436030181126112b6575f80fd5b83018035915067ffffffffffffffff8211156112d0575f80fd5b602001915036819003821315610c22575f80fd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b600181811c9082168061132557607f821691505b60208210810361135c577f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b50919050565b601f8211156113ab575f81815260208120601f850160051c810160208610156113885750805b601f850160051c820191505b818110156113a757828155600101611394565b5050505b505050565b815167ffffffffffffffff8111156113ca576113ca6112e4565b6113de816113d88454611311565b84611362565b602080601f831160018114611430575f84156113fa5750858301515b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff600386901b1c1916600185901b1785556113a7565b5f858152602081207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08616915b8281101561147c5788860151825594840194600190910190840161145d565b50858210156114b857878501517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff600388901b60f8161c191681555b5050505050600190811b01905550565b5f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82036114f8576114f8611190565b5060010190565b818382375f9101908152919050565b81810381811115610acc57610acc611190565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601260045260245ffd5b5f8261155c5761155c611521565b500690565b5f8261156f5761156f611521565b500490565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffdfea2646970667358221220c5a31942a7d3cb96f5b9df10b4a4ec6779e0daf93e3bcdcf92b7684fbc5585cf64736f6c63430008140033",
}

// PolygondatacommitteeABI is the input ABI used to generate the binding from.
// Deprecated: Use PolygondatacommitteeMetaData.ABI instead.
var PolygondatacommitteeABI = PolygondatacommitteeMetaData.ABI

// PolygondatacommitteeBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use PolygondatacommitteeMetaData.Bin instead.
var PolygondatacommitteeBin = PolygondatacommitteeMetaData.Bin

// DeployPolygondatacommittee deploys a new Ethereum contract, binding an instance of Polygondatacommittee to it.
func DeployPolygondatacommittee(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Polygondatacommittee, error) {
	parsed, err := PolygondatacommitteeMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(PolygondatacommitteeBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Polygondatacommittee{PolygondatacommitteeCaller: PolygondatacommitteeCaller{contract: contract}, PolygondatacommitteeTransactor: PolygondatacommitteeTransactor{contract: contract}, PolygondatacommitteeFilterer: PolygondatacommitteeFilterer{contract: contract}}, nil
}

// Polygondatacommittee is an auto generated Go binding around an Ethereum contract.
type Polygondatacommittee struct {
	PolygondatacommitteeCaller     // Read-only binding to the contract
	PolygondatacommitteeTransactor // Write-only binding to the contract
	PolygondatacommitteeFilterer   // Log filterer for contract events
}

// PolygondatacommitteeCaller is an auto generated read-only Go binding around an Ethereum contract.
type PolygondatacommitteeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PolygondatacommitteeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PolygondatacommitteeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PolygondatacommitteeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PolygondatacommitteeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PolygondatacommitteeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PolygondatacommitteeSession struct {
	Contract     *Polygondatacommittee // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// PolygondatacommitteeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PolygondatacommitteeCallerSession struct {
	Contract *PolygondatacommitteeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// PolygondatacommitteeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PolygondatacommitteeTransactorSession struct {
	Contract     *PolygondatacommitteeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// PolygondatacommitteeRaw is an auto generated low-level Go binding around an Ethereum contract.
type PolygondatacommitteeRaw struct {
	Contract *Polygondatacommittee // Generic contract binding to access the raw methods on
}

// PolygondatacommitteeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PolygondatacommitteeCallerRaw struct {
	Contract *PolygondatacommitteeCaller // Generic read-only contract binding to access the raw methods on
}

// PolygondatacommitteeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PolygondatacommitteeTransactorRaw struct {
	Contract *PolygondatacommitteeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPolygondatacommittee creates a new instance of Polygondatacommittee, bound to a specific deployed contract.
func NewPolygondatacommittee(address common.Address, backend bind.ContractBackend) (*Polygondatacommittee, error) {
	contract, err := bindPolygondatacommittee(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Polygondatacommittee{PolygondatacommitteeCaller: PolygondatacommitteeCaller{contract: contract}, PolygondatacommitteeTransactor: PolygondatacommitteeTransactor{contract: contract}, PolygondatacommitteeFilterer: PolygondatacommitteeFilterer{contract: contract}}, nil
}

// NewPolygondatacommitteeCaller creates a new read-only instance of Polygondatacommittee, bound to a specific deployed contract.
func NewPolygondatacommitteeCaller(address common.Address, caller bind.ContractCaller) (*PolygondatacommitteeCaller, error) {
	contract, err := bindPolygondatacommittee(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PolygondatacommitteeCaller{contract: contract}, nil
}

// NewPolygondatacommitteeTransactor creates a new write-only instance of Polygondatacommittee, bound to a specific deployed contract.
func NewPolygondatacommitteeTransactor(address common.Address, transactor bind.ContractTransactor) (*PolygondatacommitteeTransactor, error) {
	contract, err := bindPolygondatacommittee(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PolygondatacommitteeTransactor{contract: contract}, nil
}

// NewPolygondatacommitteeFilterer creates a new log filterer instance of Polygondatacommittee, bound to a specific deployed contract.
func NewPolygondatacommitteeFilterer(address common.Address, filterer bind.ContractFilterer) (*PolygondatacommitteeFilterer, error) {
	contract, err := bindPolygondatacommittee(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PolygondatacommitteeFilterer{contract: contract}, nil
}

// bindPolygondatacommittee binds a generic wrapper to an already deployed contract.
func bindPolygondatacommittee(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PolygondatacommitteeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Polygondatacommittee *PolygondatacommitteeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Polygondatacommittee.Contract.PolygondatacommitteeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Polygondatacommittee *PolygondatacommitteeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.PolygondatacommitteeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Polygondatacommittee *PolygondatacommitteeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.PolygondatacommitteeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Polygondatacommittee *PolygondatacommitteeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Polygondatacommittee.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Polygondatacommittee *PolygondatacommitteeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Polygondatacommittee *PolygondatacommitteeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.contract.Transact(opts, method, params...)
}

// CommitteeHash is a free data retrieval call binding the contract method 0x609d4544.
//
// Solidity: function committeeHash() view returns(bytes32)
func (_Polygondatacommittee *PolygondatacommitteeCaller) CommitteeHash(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Polygondatacommittee.contract.Call(opts, &out, "committeeHash")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CommitteeHash is a free data retrieval call binding the contract method 0x609d4544.
//
// Solidity: function committeeHash() view returns(bytes32)
func (_Polygondatacommittee *PolygondatacommitteeSession) CommitteeHash() ([32]byte, error) {
	return _Polygondatacommittee.Contract.CommitteeHash(&_Polygondatacommittee.CallOpts)
}

// CommitteeHash is a free data retrieval call binding the contract method 0x609d4544.
//
// Solidity: function committeeHash() view returns(bytes32)
func (_Polygondatacommittee *PolygondatacommitteeCallerSession) CommitteeHash() ([32]byte, error) {
	return _Polygondatacommittee.Contract.CommitteeHash(&_Polygondatacommittee.CallOpts)
}

// GetAmountOfMembers is a free data retrieval call binding the contract method 0xdce1e2b6.
//
// Solidity: function getAmountOfMembers() view returns(uint256)
func (_Polygondatacommittee *PolygondatacommitteeCaller) GetAmountOfMembers(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Polygondatacommittee.contract.Call(opts, &out, "getAmountOfMembers")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAmountOfMembers is a free data retrieval call binding the contract method 0xdce1e2b6.
//
// Solidity: function getAmountOfMembers() view returns(uint256)
func (_Polygondatacommittee *PolygondatacommitteeSession) GetAmountOfMembers() (*big.Int, error) {
	return _Polygondatacommittee.Contract.GetAmountOfMembers(&_Polygondatacommittee.CallOpts)
}

// GetAmountOfMembers is a free data retrieval call binding the contract method 0xdce1e2b6.
//
// Solidity: function getAmountOfMembers() view returns(uint256)
func (_Polygondatacommittee *PolygondatacommitteeCallerSession) GetAmountOfMembers() (*big.Int, error) {
	return _Polygondatacommittee.Contract.GetAmountOfMembers(&_Polygondatacommittee.CallOpts)
}

// GetProcotolName is a free data retrieval call binding the contract method 0xe4f17120.
//
// Solidity: function getProcotolName() pure returns(string)
func (_Polygondatacommittee *PolygondatacommitteeCaller) GetProcotolName(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Polygondatacommittee.contract.Call(opts, &out, "getProcotolName")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetProcotolName is a free data retrieval call binding the contract method 0xe4f17120.
//
// Solidity: function getProcotolName() pure returns(string)
func (_Polygondatacommittee *PolygondatacommitteeSession) GetProcotolName() (string, error) {
	return _Polygondatacommittee.Contract.GetProcotolName(&_Polygondatacommittee.CallOpts)
}

// GetProcotolName is a free data retrieval call binding the contract method 0xe4f17120.
//
// Solidity: function getProcotolName() pure returns(string)
func (_Polygondatacommittee *PolygondatacommitteeCallerSession) GetProcotolName() (string, error) {
	return _Polygondatacommittee.Contract.GetProcotolName(&_Polygondatacommittee.CallOpts)
}

// Members is a free data retrieval call binding the contract method 0x5daf08ca.
//
// Solidity: function members(uint256 ) view returns(string url, address addr)
func (_Polygondatacommittee *PolygondatacommitteeCaller) Members(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Url  string
	Addr common.Address
}, error) {
	var out []interface{}
	err := _Polygondatacommittee.contract.Call(opts, &out, "members", arg0)

	outstruct := new(struct {
		Url  string
		Addr common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Url = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Addr = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// Members is a free data retrieval call binding the contract method 0x5daf08ca.
//
// Solidity: function members(uint256 ) view returns(string url, address addr)
func (_Polygondatacommittee *PolygondatacommitteeSession) Members(arg0 *big.Int) (struct {
	Url  string
	Addr common.Address
}, error) {
	return _Polygondatacommittee.Contract.Members(&_Polygondatacommittee.CallOpts, arg0)
}

// Members is a free data retrieval call binding the contract method 0x5daf08ca.
//
// Solidity: function members(uint256 ) view returns(string url, address addr)
func (_Polygondatacommittee *PolygondatacommitteeCallerSession) Members(arg0 *big.Int) (struct {
	Url  string
	Addr common.Address
}, error) {
	return _Polygondatacommittee.Contract.Members(&_Polygondatacommittee.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Polygondatacommittee *PolygondatacommitteeCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Polygondatacommittee.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Polygondatacommittee *PolygondatacommitteeSession) Owner() (common.Address, error) {
	return _Polygondatacommittee.Contract.Owner(&_Polygondatacommittee.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Polygondatacommittee *PolygondatacommitteeCallerSession) Owner() (common.Address, error) {
	return _Polygondatacommittee.Contract.Owner(&_Polygondatacommittee.CallOpts)
}

// RequiredAmountOfSignatures is a free data retrieval call binding the contract method 0x6beedd39.
//
// Solidity: function requiredAmountOfSignatures() view returns(uint256)
func (_Polygondatacommittee *PolygondatacommitteeCaller) RequiredAmountOfSignatures(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Polygondatacommittee.contract.Call(opts, &out, "requiredAmountOfSignatures")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RequiredAmountOfSignatures is a free data retrieval call binding the contract method 0x6beedd39.
//
// Solidity: function requiredAmountOfSignatures() view returns(uint256)
func (_Polygondatacommittee *PolygondatacommitteeSession) RequiredAmountOfSignatures() (*big.Int, error) {
	return _Polygondatacommittee.Contract.RequiredAmountOfSignatures(&_Polygondatacommittee.CallOpts)
}

// RequiredAmountOfSignatures is a free data retrieval call binding the contract method 0x6beedd39.
//
// Solidity: function requiredAmountOfSignatures() view returns(uint256)
func (_Polygondatacommittee *PolygondatacommitteeCallerSession) RequiredAmountOfSignatures() (*big.Int, error) {
	return _Polygondatacommittee.Contract.RequiredAmountOfSignatures(&_Polygondatacommittee.CallOpts)
}

// VerifyMessage is a free data retrieval call binding the contract method 0x3b51be4b.
//
// Solidity: function verifyMessage(bytes32 signedHash, bytes signaturesAndAddrs) view returns()
func (_Polygondatacommittee *PolygondatacommitteeCaller) VerifyMessage(opts *bind.CallOpts, signedHash [32]byte, signaturesAndAddrs []byte) error {
	var out []interface{}
	err := _Polygondatacommittee.contract.Call(opts, &out, "verifyMessage", signedHash, signaturesAndAddrs)

	if err != nil {
		return err
	}

	return err

}

// VerifyMessage is a free data retrieval call binding the contract method 0x3b51be4b.
//
// Solidity: function verifyMessage(bytes32 signedHash, bytes signaturesAndAddrs) view returns()
func (_Polygondatacommittee *PolygondatacommitteeSession) VerifyMessage(signedHash [32]byte, signaturesAndAddrs []byte) error {
	return _Polygondatacommittee.Contract.VerifyMessage(&_Polygondatacommittee.CallOpts, signedHash, signaturesAndAddrs)
}

// VerifyMessage is a free data retrieval call binding the contract method 0x3b51be4b.
//
// Solidity: function verifyMessage(bytes32 signedHash, bytes signaturesAndAddrs) view returns()
func (_Polygondatacommittee *PolygondatacommitteeCallerSession) VerifyMessage(signedHash [32]byte, signaturesAndAddrs []byte) error {
	return _Polygondatacommittee.Contract.VerifyMessage(&_Polygondatacommittee.CallOpts, signedHash, signaturesAndAddrs)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Polygondatacommittee *PolygondatacommitteeTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Polygondatacommittee.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Polygondatacommittee *PolygondatacommitteeSession) Initialize() (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.Initialize(&_Polygondatacommittee.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Polygondatacommittee *PolygondatacommitteeTransactorSession) Initialize() (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.Initialize(&_Polygondatacommittee.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Polygondatacommittee *PolygondatacommitteeTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Polygondatacommittee.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Polygondatacommittee *PolygondatacommitteeSession) RenounceOwnership() (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.RenounceOwnership(&_Polygondatacommittee.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Polygondatacommittee *PolygondatacommitteeTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.RenounceOwnership(&_Polygondatacommittee.TransactOpts)
}

// SetupCommittee is a paid mutator transaction binding the contract method 0x078fba2a.
//
// Solidity: function setupCommittee(uint256 _requiredAmountOfSignatures, string[] urls, bytes addrsBytes) returns()
func (_Polygondatacommittee *PolygondatacommitteeTransactor) SetupCommittee(opts *bind.TransactOpts, _requiredAmountOfSignatures *big.Int, urls []string, addrsBytes []byte) (*types.Transaction, error) {
	return _Polygondatacommittee.contract.Transact(opts, "setupCommittee", _requiredAmountOfSignatures, urls, addrsBytes)
}

// SetupCommittee is a paid mutator transaction binding the contract method 0x078fba2a.
//
// Solidity: function setupCommittee(uint256 _requiredAmountOfSignatures, string[] urls, bytes addrsBytes) returns()
func (_Polygondatacommittee *PolygondatacommitteeSession) SetupCommittee(_requiredAmountOfSignatures *big.Int, urls []string, addrsBytes []byte) (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.SetupCommittee(&_Polygondatacommittee.TransactOpts, _requiredAmountOfSignatures, urls, addrsBytes)
}

// SetupCommittee is a paid mutator transaction binding the contract method 0x078fba2a.
//
// Solidity: function setupCommittee(uint256 _requiredAmountOfSignatures, string[] urls, bytes addrsBytes) returns()
func (_Polygondatacommittee *PolygondatacommitteeTransactorSession) SetupCommittee(_requiredAmountOfSignatures *big.Int, urls []string, addrsBytes []byte) (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.SetupCommittee(&_Polygondatacommittee.TransactOpts, _requiredAmountOfSignatures, urls, addrsBytes)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Polygondatacommittee *PolygondatacommitteeTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Polygondatacommittee.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Polygondatacommittee *PolygondatacommitteeSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.TransferOwnership(&_Polygondatacommittee.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Polygondatacommittee *PolygondatacommitteeTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Polygondatacommittee.Contract.TransferOwnership(&_Polygondatacommittee.TransactOpts, newOwner)
}

// PolygondatacommitteeCommitteeUpdatedIterator is returned from FilterCommitteeUpdated and is used to iterate over the raw logs and unpacked data for CommitteeUpdated events raised by the Polygondatacommittee contract.
type PolygondatacommitteeCommitteeUpdatedIterator struct {
	Event *PolygondatacommitteeCommitteeUpdated // Event containing the contract specifics and raw log

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
func (it *PolygondatacommitteeCommitteeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PolygondatacommitteeCommitteeUpdated)
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
		it.Event = new(PolygondatacommitteeCommitteeUpdated)
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
func (it *PolygondatacommitteeCommitteeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PolygondatacommitteeCommitteeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PolygondatacommitteeCommitteeUpdated represents a CommitteeUpdated event raised by the Polygondatacommittee contract.
type PolygondatacommitteeCommitteeUpdated struct {
	CommitteeHash [32]byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterCommitteeUpdated is a free log retrieval operation binding the contract event 0x831403fd381b3e6ac875d912ec2eee0e0203d0d29f7b3e0c96fc8f582d6db657.
//
// Solidity: event CommitteeUpdated(bytes32 committeeHash)
func (_Polygondatacommittee *PolygondatacommitteeFilterer) FilterCommitteeUpdated(opts *bind.FilterOpts) (*PolygondatacommitteeCommitteeUpdatedIterator, error) {

	logs, sub, err := _Polygondatacommittee.contract.FilterLogs(opts, "CommitteeUpdated")
	if err != nil {
		return nil, err
	}
	return &PolygondatacommitteeCommitteeUpdatedIterator{contract: _Polygondatacommittee.contract, event: "CommitteeUpdated", logs: logs, sub: sub}, nil
}

// WatchCommitteeUpdated is a free log subscription operation binding the contract event 0x831403fd381b3e6ac875d912ec2eee0e0203d0d29f7b3e0c96fc8f582d6db657.
//
// Solidity: event CommitteeUpdated(bytes32 committeeHash)
func (_Polygondatacommittee *PolygondatacommitteeFilterer) WatchCommitteeUpdated(opts *bind.WatchOpts, sink chan<- *PolygondatacommitteeCommitteeUpdated) (event.Subscription, error) {

	logs, sub, err := _Polygondatacommittee.contract.WatchLogs(opts, "CommitteeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PolygondatacommitteeCommitteeUpdated)
				if err := _Polygondatacommittee.contract.UnpackLog(event, "CommitteeUpdated", log); err != nil {
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

// ParseCommitteeUpdated is a log parse operation binding the contract event 0x831403fd381b3e6ac875d912ec2eee0e0203d0d29f7b3e0c96fc8f582d6db657.
//
// Solidity: event CommitteeUpdated(bytes32 committeeHash)
func (_Polygondatacommittee *PolygondatacommitteeFilterer) ParseCommitteeUpdated(log types.Log) (*PolygondatacommitteeCommitteeUpdated, error) {
	event := new(PolygondatacommitteeCommitteeUpdated)
	if err := _Polygondatacommittee.contract.UnpackLog(event, "CommitteeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PolygondatacommitteeInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Polygondatacommittee contract.
type PolygondatacommitteeInitializedIterator struct {
	Event *PolygondatacommitteeInitialized // Event containing the contract specifics and raw log

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
func (it *PolygondatacommitteeInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PolygondatacommitteeInitialized)
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
		it.Event = new(PolygondatacommitteeInitialized)
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
func (it *PolygondatacommitteeInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PolygondatacommitteeInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PolygondatacommitteeInitialized represents a Initialized event raised by the Polygondatacommittee contract.
type PolygondatacommitteeInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Polygondatacommittee *PolygondatacommitteeFilterer) FilterInitialized(opts *bind.FilterOpts) (*PolygondatacommitteeInitializedIterator, error) {

	logs, sub, err := _Polygondatacommittee.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &PolygondatacommitteeInitializedIterator{contract: _Polygondatacommittee.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Polygondatacommittee *PolygondatacommitteeFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *PolygondatacommitteeInitialized) (event.Subscription, error) {

	logs, sub, err := _Polygondatacommittee.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PolygondatacommitteeInitialized)
				if err := _Polygondatacommittee.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Polygondatacommittee *PolygondatacommitteeFilterer) ParseInitialized(log types.Log) (*PolygondatacommitteeInitialized, error) {
	event := new(PolygondatacommitteeInitialized)
	if err := _Polygondatacommittee.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PolygondatacommitteeOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Polygondatacommittee contract.
type PolygondatacommitteeOwnershipTransferredIterator struct {
	Event *PolygondatacommitteeOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *PolygondatacommitteeOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PolygondatacommitteeOwnershipTransferred)
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
		it.Event = new(PolygondatacommitteeOwnershipTransferred)
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
func (it *PolygondatacommitteeOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PolygondatacommitteeOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PolygondatacommitteeOwnershipTransferred represents a OwnershipTransferred event raised by the Polygondatacommittee contract.
type PolygondatacommitteeOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Polygondatacommittee *PolygondatacommitteeFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*PolygondatacommitteeOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Polygondatacommittee.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &PolygondatacommitteeOwnershipTransferredIterator{contract: _Polygondatacommittee.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Polygondatacommittee *PolygondatacommitteeFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *PolygondatacommitteeOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Polygondatacommittee.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PolygondatacommitteeOwnershipTransferred)
				if err := _Polygondatacommittee.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Polygondatacommittee *PolygondatacommitteeFilterer) ParseOwnershipTransferred(log types.Log) (*PolygondatacommitteeOwnershipTransferred, error) {
	event := new(PolygondatacommitteeOwnershipTransferred)
	if err := _Polygondatacommittee.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
