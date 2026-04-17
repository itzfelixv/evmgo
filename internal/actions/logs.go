package actions

import (
	"context"
	"fmt"

	abiutil "github.com/itzfelixv/evmgo/internal/abi"
	"github.com/itzfelixv/evmgo/internal/eth"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

type LogQuery struct {
	Address string
	ABIPath string
	Event   string
	Topics  []string
	From    rpc.BlockRef
	To      rpc.BlockRef
}

type LogResult struct {
	Address         string   `json:"address"`
	BlockNumber     string   `json:"blockNumber"`
	TransactionHash string   `json:"transactionHash"`
	Data            string   `json:"data"`
	Topics          []string `json:"topics"`
}

func GetLogs(ctx context.Context, client *rpc.Client, query LogQuery) ([]LogResult, error) {
	normalizedAddress, err := eth.NormalizeAddress(query.Address)
	if err != nil {
		return nil, err
	}

	topics := make([]any, 0, len(query.Topics)+1)
	if query.Event != "" {
		if query.ABIPath == "" {
			return nil, fmt.Errorf("--abi is required when --event is set")
		}
		contractABI, err := abiutil.LoadFile(query.ABIPath)
		if err != nil {
			return nil, err
		}
		topic0, err := abiutil.EventTopic(contractABI, query.Event)
		if err != nil {
			return nil, err
		}
		topics = append(topics, topic0)
	}
	for _, topic := range query.Topics {
		topics = append(topics, topic)
	}

	filter := map[string]any{
		"address":   normalizedAddress,
		"fromBlock": query.From.RPCArg(),
		"toBlock":   query.To.RPCArg(),
	}
	if len(topics) > 0 {
		filter["topics"] = topics
	}

	var result []LogResult
	if err := client.Call(ctx, "eth_getLogs", []any{filter}, &result); err != nil {
		return nil, err
	}
	return result, nil
}
