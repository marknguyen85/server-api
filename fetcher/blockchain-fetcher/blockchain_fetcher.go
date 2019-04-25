package bfetcher

import (
	"context"
	"log"
	"time"

	// "strconv"

	"github.com/tomochain/tomochain/common/hexutil"
	"github.com/tomochain/tomochain/rpc"
)

type BlockchainFetcher struct {
	client   *rpc.Client
	url      string
	TypeName string

	timeout time.Duration
}

func NewBlockchainFetcher(typeName string, endpoint string, apiKey string) (*BlockchainFetcher, error) {
	client, err := rpc.DialHTTP(endpoint)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	timeout := 5 * time.Second
	blockchain := BlockchainFetcher{
		client:   client,
		url:      endpoint,
		TypeName: typeName,
		timeout:  timeout,
	}
	return &blockchain, nil
}

//TomoCall func
func (blcFetcher *BlockchainFetcher) TomoCall(to string, data string) (string, error) {
	params := make(map[string]string)
	params["data"] = "0x" + data
	params["to"] = to

	ctx, cancel := context.WithTimeout(context.Background(), blcFetcher.timeout)
	defer cancel()
	var result string
	err := blcFetcher.client.CallContext(ctx, &result, "eth_call", params, "latest")
	if err != nil {
		log.Print(err)
		return "", err
	}

	return result, nil
}

//GetRate func
func (blcFetcher *BlockchainFetcher) GetRate(to string, data string) (string, error) {
	params := make(map[string]string)
	params["data"] = "0x" + data
	params["to"] = to

	ctx, cancel := context.WithTimeout(context.Background(), blcFetcher.timeout)
	defer cancel()
	var result string
	err := blcFetcher.client.CallContext(ctx, &result, "eth_call", params, "latest")
	if err != nil {
		return "", err
	}

	return result, nil

}

//GetLatestBlock func
func (blcFetcher *BlockchainFetcher) GetLatestBlock() (string, error) {
	var blockNum *hexutil.Big
	ctx, cancel := context.WithTimeout(context.Background(), blcFetcher.timeout)
	defer cancel()
	err := blcFetcher.client.CallContext(ctx, &blockNum, "eth_blockNumber", "latest")
	if err != nil {
		return "", err
	}
	return blockNum.ToInt().String(), nil
}

//TopicParam struct
type TopicParam struct {
	FromBlock string   `json:"fromBlock"`
	ToBlock   string   `json:"toBlock"`
	Address   string   `json:"address"`
	Topics    []string `json:"topics"`
}

//GetTypeName get type name
func (blcFetcher *BlockchainFetcher) GetTypeName() string {
	return blcFetcher.TypeName
}
