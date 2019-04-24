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

//Tomoscan func
type Tomoscan struct {
	url      string
	apiKey   string
	TypeName string
}

//ResultEvent func
type ResultEvent struct {
	Result []tomochain.EventRaw `json:"result"`
}

func NewTomoScan(typeName string, url string, apiKey string) (*Tomoscan, error) {
	tomoscan := Tomoscan{
		url, apiKey, typeName,
	}
	return &tomoscan, nil
}

//TomoCall func
func (tomoscan *Tomoscan) TomoCall(to string, data string) (string, error) {
	url := tomoscan.url + "/api?module=proxy&action=eth_call&to=" +
		to + "&data=" + data + "&tag=latest&apikey=" + tomoscan.apiKey
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

//GetRate func
func (tomoscan *Tomoscan) GetRate(to string, data string) (string, error) {
	return "", errors.New("not support this func")
}

//GetLatestBlock func
func (tomoscan *Tomoscan) GetLatestBlock() (string, error) {
	url := tomoscan.url + "/api?module=proxy&action=eth_blockNumber"
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

//GetTypeName func
func (tomoscan *Tomoscan) GetTypeName() string {
	return tomoscan.TypeName
}
