package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/rahul0tripathi/smelter/forkdb"
	"github.com/rahul0tripathi/smelter/vm"
)

func GetHash(uint64) common.Hash {
	return common.HexToHash("")
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func fn(ref *types.Header, chain core.ChainContext) func(n uint64) common.Hash {
	// Cache will initially contain [refHash.parent],
	// Then fill up with [refHash.p, refHash.pp, refHash.ppp, ...]
	var cache []common.Hash

	return func(n uint64) common.Hash {
		if ref.Number.Uint64() <= n {
			// This situation can happen if we're doing tracing and using
			// block overrides.
			return common.Hash{}
		}
		// If there's no hash cache yet, make one
		if len(cache) == 0 {
			cache = append(cache, ref.ParentHash)
		}
		if idx := ref.Number.Uint64() - n - 1; idx < uint64(len(cache)) {
			return cache[idx]
		}
		// No luck in the cache, but we can start iterating from the last element we already know
		lastKnownHash := cache[len(cache)-1]
		lastKnownNumber := ref.Number.Uint64() - uint64(len(cache))

		for {
			header := chain.GetHeader(lastKnownHash, lastKnownNumber)
			if header == nil {
				break
			}
			cache = append(cache, header.ParentHash)
			lastKnownHash = header.ParentHash
			lastKnownNumber = header.Number.Uint64() - 1
			if n == lastKnownNumber {
				return lastKnownHash
			}
		}
		return common.Hash{}
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr common.Address, amount *uint256.Int) bool {
	fmt.Println("can transfer", addr, amount.String())
	//return db.GetBalance(addr).Cmp(amount) >= 0
	return true
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, amount *uint256.Int) {
	fmt.Println("transfer called", sender, recipient, amount.String())
	//db.SubBalance(sender, amount, tracing.BalanceChangeTransfer)
	//db.AddBalance(recipient, amount, tracing.BalanceChangeTransfer)
}

type Config struct {
	ChainConfig *params.ChainConfig
	Difficulty  *big.Int
	Origin      common.Address
	Coinbase    common.Address
	BlockNumber *big.Int
	Time        uint64
	GasLimit    uint64
	GasPrice    *big.Int
	Value       *big.Int
	Debug       bool
	EVMConfig   vm.Config
	BaseFee     *big.Int
	BlobBaseFee *big.Int
	BlobHashes  []common.Hash
	BlobFeeCap  *big.Int
	Random      *common.Hash

	State     *state.StateDB
	GetHashFn func(n uint64) common.Hash
}

// sets defaults on the config
func setDefaults(cfg *Config) {
	if cfg.ChainConfig == nil {
		cfg.ChainConfig = &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      new(big.Int),
			DAOForkBlock:        new(big.Int),
			DAOForkSupport:      false,
			EIP150Block:         new(big.Int),
			EIP155Block:         new(big.Int),
			EIP158Block:         new(big.Int),
			ByzantiumBlock:      new(big.Int),
			ConstantinopleBlock: new(big.Int),
			PetersburgBlock:     new(big.Int),
			IstanbulBlock:       new(big.Int),
			MuirGlacierBlock:    new(big.Int),
			BerlinBlock:         new(big.Int),
			LondonBlock:         new(big.Int),
		}
	}

	if cfg.Difficulty == nil {
		cfg.Difficulty = new(big.Int)
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = math.MaxUint64
	}
	if cfg.GasPrice == nil {
		cfg.GasPrice = new(big.Int)
	}
	if cfg.Value == nil {
		cfg.Value = new(big.Int)
	}
	if cfg.BlockNumber == nil {
		cfg.BlockNumber = new(big.Int)
	}
	if cfg.GetHashFn == nil {
		cfg.GetHashFn = func(n uint64) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(new(big.Int).SetUint64(n).String())))
		}
	}
	if cfg.BaseFee == nil {
		cfg.BaseFee = big.NewInt(params.InitialBaseFee)
	}
	if cfg.BlobBaseFee == nil {
		cfg.BlobBaseFee = big.NewInt(params.BlobTxMinBlobGasprice)
	}
}

func main() {
	cfg := Config{}
	setDefaults(&cfg)

	fmt.Println("hello world ")

	sender := common.HexToAddress("0x0000000000000000000000000000000000000000")
	cfg.Origin = sender
	db := forkdb.ForkDB{}
	tracer := tracing.Hooks{
		OnTxStart: func(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
			fmt.Println("Transaction Start:", tx)
		},
		OnTxEnd: func(receipt *types.Receipt, err error) {
			fmt.Println("Transaction End:", receipt, receipt)
		},
		OnEnter: func(depth int, op byte, from, to common.Address, input []byte, gas uint64, value *big.Int) {
			fmt.Println("Enter:", depth, op, from, to, input, gas, value)
		},
		OnExit: func(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
			fmt.Println("Exit:", depth, output, gasUsed, err, reverted)
		},
		OnOpcode: func(
			pc uint64,
			op byte,
			gas, cost uint64,
			scope tracing.OpContext,
			rData []byte,
			depth int,
			err error,
		) {
			fmt.Println("Opcode:", pc, op, gas, cost, scope, rData, depth, err)
		},
		OnFault: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
			fmt.Println("Fault:", pc, op, gas, cost, scope, depth, err)
		},
		OnGasChange: func(prevGas, newGas uint64, reason tracing.GasChangeReason) {
			fmt.Println("Gas Change:", prevGas, newGas, reason)
		},
		OnBlockchainInit: func(chainConfig *params.ChainConfig) {
			fmt.Println("Blockchain Init:", chainConfig)
		},
		OnClose: func() {
			fmt.Println("Close")
		},
		OnBlockStart: func(event tracing.BlockEvent) {
			fmt.Println("Block Start:", event.Block)
		},
		OnBlockEnd: func(err error) {
			fmt.Println("Block End:", err)
		},
		OnSkippedBlock: func(event tracing.BlockEvent) {
			fmt.Println("Skipped Block:", event.Block)
		},
		OnGenesisBlock: func(genesis *types.Block, alloc types.GenesisAlloc) {
			fmt.Println("Genesis Block:", genesis, alloc)
		},
		OnSystemCallStart: func() {
			fmt.Println("System Call Start:")
		},
		OnSystemCallEnd: func() {
			fmt.Println("System Call End:")
		},
		OnBalanceChange: func(addr common.Address, prev, new *big.Int, reason tracing.BalanceChangeReason) {
			fmt.Println("Balance Change:", addr, prev, new, string(reason))
		},
		OnNonceChange: func(address common.Address, prevNonce, newNonce uint64) {
			fmt.Println("Nonce Change:", address, prevNonce, newNonce)
		},
		OnCodeChange: func(
			addr common.Address,
			prevCodeHash common.Hash,
			prevCode []byte,
			codeHash common.Hash,
			code []byte,
		) {
			fmt.Println("Code Change:", addr, prevCode, prevCodeHash, codeHash, code)
		},
		OnStorageChange: func(addr common.Address, slot common.Hash, prev, new common.Hash) {
			fmt.Println("Storage Change:", addr, slot, prev, new)
		},
		OnLog: func(log *types.Log) {
			fmt.Println("Log:", log)
		},
	}

	cfg.EVMConfig.Tracer = &tracer

	txContext := vm.TxContext{
		Origin:     cfg.Origin,
		GasPrice:   cfg.GasPrice,
		BlobHashes: cfg.BlobHashes,
		BlobFeeCap: cfg.BlobFeeCap,
	}

	blockContext := vm.BlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHash,
		Coinbase:    cfg.Coinbase,
		BlockNumber: cfg.BlockNumber,
		Time:        cfg.Time,
		Difficulty:  cfg.Difficulty,
		GasLimit:    cfg.GasLimit,
		BaseFee:     cfg.BaseFee,
		BlobBaseFee: cfg.BlobBaseFee,
		Random:      nil,
	}
	senderRef := vm.AccountRef(cfg.Origin)
	target := common.HexToAddress("0x0000000000000000000000000000000000000001")
	vmenv := vm.NewEVM(blockContext, txContext, &db, cfg.ChainConfig, cfg.EVMConfig)
	data, _ := hexutil.Decode("0x")
	ret, leftOverGas, err := vmenv.Call(
		senderRef,
		target,
		data,
		30000000,
		uint256.MustFromBig(new(big.Int).SetInt64(100)),
	)

	fmt.Println(ret, leftOverGas, err)
}
