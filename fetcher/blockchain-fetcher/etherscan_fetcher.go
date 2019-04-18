package bfetcher

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/ChainTex/server-go/tomochain"
	fCommon "github.com/ChainTex/server-go/fetcher/fetcher-common"
	"github.com/tomochain/tomochain/common/hexutil"
)

// api_key for tracker.kyber
const (
	TIME_TO_DELETE = 18000
)

type Etherscan struct {
	url      string
	apiKey   string
	TypeName string
}

type ResultEvent struct {
	Result []tomochain.EventRaw `json:"result"`
}

func NewEtherScan(typeName string, url string, apiKey string) (*Etherscan, error) {
	etherscan := Etherscan{
		url, apiKey, typeName,
	}
	return &etherscan, nil
}

func (self *Etherscan) EthCall(to string, data string) (string, error) {
	url := self.url + "/api?module=proxy&action=eth_call&to=" +
		to + "&data=" + data + "&tag=latest&apikey=" + self.apiKey
	b, err := fCommon.HTTPCall(url)
	if err != nil {
		log.Print(err)
		return "", err
	}
	result := tomochain.ResultRpc{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Print(err)
		return "", err
	}

	return result.Result, nil

}

func (self *Etherscan) GetRate(to string, data string) (string, error) {
	return "", errors.New("not support this func")
}

func (self *Etherscan) GetLatestBlock() (string, error) {
	url := self.url + "/api?module=proxy&action=eth_blockNumber"
	b, err := fCommon.HTTPCall(url)
	if err != nil {
		log.Print(err)
		return "", err
	}
	blockNum := tomochain.ResultRpc{}
	err = json.Unmarshal(b, &blockNum)
	if err != nil {
		return "", err
	}
	num, err := hexutil.DecodeBig(blockNum.Result)
	if err != nil {
		return "", err
	}
	return num.String(), nil
}

// func (self *Etherscan) GetEvents(fromBlock string, toBlock string, network string, tradeTopic string) (*[]tomochain.EventRaw, error) {
// 	url := self.url + "/api?module=logs&action=getLogs&fromBlock=" +
// 		fromBlock + "&toBlock=" + toBlock + "&address=" + network + "&topic0=" +
// 		tradeTopic + "&apikey=" + self.apiKey
// 	response, err := http.Get(url)
// 	if err != nil {
// 		log.Print(err)
// 		return nil, err
// 	}
// 	if response.StatusCode != 200 {
// 		return nil, errors.New("Status code is 200")
// 	}

// 	defer (response.Body).Close()
// 	b, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		log.Print(err)
// 		return nil, err
// 	}
// 	result := ResultEvent{}
// 	err = json.Unmarshal(b, &result)
// 	if err != nil {
// 		log.Print(err)
// 		return nil, err
// 	}

// 	return &result.Result, nil
// }

// func (self *Etherscan) GetRateUsd(tickers []string) ([]io.ReadCloser, error) {
// 	outPut := make([]io.ReadCloser, 0)
// 	for _, ticker := range tickers {
// 		response, err := http.Get("https://api.coinmarketcap.com/v1/ticker/" + ticker)
// 		if err != nil {
// 			log.Print(err)
// 			return nil, err
// 		}
// 		outPut = append(outPut, response.Body)
// 	}

// 	return outPut, nil
// }

func (self *Etherscan) GetTypeName() string {
	return self.TypeName
}
