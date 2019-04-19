package bfetcher

import (
	"encoding/json"
	"errors"
	"log"

	fCommon "github.com/marknguyen85/server-api/fetcher/fetcher-common"
	"github.com/marknguyen85/server-api/tomochain"
	"github.com/tomochain/tomochain/common/hexutil"
)

// api_key for tracker.kyber
const (
	TIME_TO_DELETE = 18000
)

type Tomoscan struct {
	url      string
	apiKey   string
	TypeName string
}

type ResultEvent struct {
	Result []tomochain.EventRaw `json:"result"`
}

func NewTomoScan(typeName string, url string, apiKey string) (*Tomoscan, error) {
	tomoscan := Tomoscan{
		url, apiKey, typeName,
	}
	return &tomoscan, nil
}

func (self *Tomoscan) EthCall(to string, data string) (string, error) {
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

func (self *Tomoscan) GetRate(to string, data string) (string, error) {
	return "", errors.New("not support this func")
}

func (self *Tomoscan) GetLatestBlock() (string, error) {
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

// func (self *Tomoscan) GetEvents(fromBlock string, toBlock string, network string, tradeTopic string) (*[]tomochain.EventRaw, error) {
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

// func (self *Tomoscan) GetRateUsd(tickers []string) ([]io.ReadCloser, error) {
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

func (self *Tomoscan) GetTypeName() string {
	return self.TypeName
}
