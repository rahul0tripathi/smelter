package entity

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rahul0tripathi/smelter/utils"
)

type TransactionSearchResponse struct {
	Txs      []*SerializedTransaction `json:"txs"`
	Receipts []*SerializedReceipt     `json:"receipts"`
}

type BlockIssuance struct {
	BlockReward uint64 `json:"blockReward"`
	UncleReward uint64 `json:"uncleReward"`
	Issuance    uint64 `json:"issuance"`
}

type BlockData struct {
	Hash                  common.Hash    `json:"hash"`
	ParentHash            common.Hash    `json:"parentHash"`
	Number                *big.Int       `json:"number"`
	Timestamp             uint64         `json:"timestamp"`
	Nonce                 string         `json:"nonce"`
	Difficulty            *big.Int       `json:"difficulty"`
	TotalDifficulty       string         `json:"totalDifficulty"`
	GasLimit              uint64         `json:"gasLimit"`
	GasUsed               uint64         `json:"gasUsed"`
	Miner                 common.Address `json:"miner"`
	ExtraData             string         `json:"extraData"`
	ReceiptsRoot          common.Hash    `json:"receiptsRoot,omitempty"`
	StateRoot             common.Hash    `json:"stateRoot,omitempty"`
	Transactions          []common.Hash  `json:"transactions,omitempty"`
	BaseFeePerGas         *big.Int       `json:"baseFeePerGas,omitempty"`
	BlobGasUsed           *big.Int       `json:"blobGasUsed,omitempty"`
	ExcessBlobGas         *big.Int       `json:"excessBlobGas,omitempty"`
	ParentBeaconBlockRoot common.Hash    `json:"parentBeaconBlockRoot,omitempty"`
	TransactionCount      uint64         `json:"transactionCount"`
	Size                  string         `json:"size"`
	Sha3Uncles            common.Hash    `json:"sha3Uncles"`
}

type BlockDetailResponse struct {
	Block     BlockData     `json:"block"`
	TotalFees string        `json:"totalFees"`
	Issuance  BlockIssuance `json:"issuance"`
}

func SerializeBlockDetail(block *types.Block) *BlockDetailResponse {
	if block == nil {
		return nil
	}
	txs := make([]common.Hash, 0)
	for _, tx := range block.Transactions() {
		txs = append(txs, tx.Hash())
	}

	return &BlockDetailResponse{
		Block: BlockData{
			Hash:             block.Hash(),
			ParentHash:       block.ParentHash(),
			Number:           block.Number(),
			Timestamp:        uint64(time.Now().Unix()),
			Nonce:            utils.Big2Hex(new(big.Int).SetUint64(block.Nonce())),
			Difficulty:       block.Difficulty(),
			GasLimit:         block.GasLimit(),
			GasUsed:          block.GasUsed(),
			Miner:            common.Address{},
			ReceiptsRoot:     block.ReceiptHash(),
			StateRoot:        block.Root(),
			Transactions:     txs,
			ExtraData:        "0x00",
			TransactionCount: uint64(len(txs)),
			Size:             "0x0",
			TotalDifficulty:  "0x0",
		},
		Issuance:  BlockIssuance{},
		TotalFees: "0x0",
	}
}

type FullBlock struct {
	BlockData
	Transactions []*SerializedTransaction `json:"transactions"`
}

type BlockTransactionsResponse struct {
	FullBlock FullBlock            `json:"fullblock"`
	Receipts  []*SerializedReceipt `json:"receipts"`
}
