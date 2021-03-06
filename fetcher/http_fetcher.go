package fetcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/marknguyen85/server-api/common"
	fCommon "github.com/marknguyen85/server-api/fetcher/fetcher-common"
	"github.com/marknguyen85/server-api/tomochain"
)

const (
	TIME_TO_DELETE  = 18000
	API_KEY_TRACKER = "jHGlaMKcGn5cCBxQCGwusS4VcnH0C6tN"
)

type HTTPFetcher struct {
	tradingAPIEndpoint string
	gasStationEndPoint string
	apiEndpoint        string
}

func NewHTTPFetcher(tradingAPIEndpoint, gasStationEndpoint, apiEndpoint string) *HTTPFetcher {
	return &HTTPFetcher{
		tradingAPIEndpoint: tradingAPIEndpoint,
		gasStationEndPoint: gasStationEndpoint,
		apiEndpoint:        apiEndpoint,
	}
}

func (httpFetcher *HTTPFetcher) GetListToken() ([]tomochain.Token, error) {
	b, err := fCommon.HTTPCall(httpFetcher.tradingAPIEndpoint)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	var result tomochain.TokenConfig
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	if result.Success == false {
		err = errors.New("Cannot get list token")
		return nil, err
	}
	data := result.Data
	if len(data) == 0 {
		err = errors.New("list token from api is empty")
		return nil, err
	}

	return data, nil
}

type GasStation struct {
	Fast     float64 `json:"fast"`
	Standard float64 `json:"average"`
	Low      float64 `json:"safeLow"`
}

func (httpFetcher *HTTPFetcher) GetGasPrice() (*tomochain.GasPrice, error) {
	b, err := fCommon.HTTPCall(httpFetcher.gasStationEndPoint)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	var gasPrice GasStation
	err = json.Unmarshal(b, &gasPrice)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	fast := big.NewFloat(gasPrice.Fast / 10)
	standard := big.NewFloat((gasPrice.Fast + gasPrice.Standard) / 20)
	low := big.NewFloat(gasPrice.Low / 10)
	defaultGas := standard

	return &tomochain.GasPrice{
		fast.String(), standard.String(), low.String(), defaultGas.String(),
	}, nil
}

// get data from tracker.kyber

func (httpFetcher *HTTPFetcher) GetRate7dData() (map[string]*tomochain.Rates, error) {
	trackerAPI := httpFetcher.apiEndpoint + "/rates7d"
	b, err := fCommon.HTTPCall(trackerAPI)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	trackerData := map[string]*tomochain.Rates{}
	err = json.Unmarshal(b, &trackerData)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return trackerData, nil
}

func (httpFetcher *HTTPFetcher) GetUserInfo(url string) (*common.UserInfo, error) {
	userInfo := &common.UserInfo{}
	b, err := fCommon.HTTPCall(url)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	err = json.Unmarshal(b, userInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return userInfo, nil
}

//TokenPrices
type TokenPrice struct {
	Data []struct {
		Symbol string  `json:"symbol"`
		Price  float64 `json:"price"`
	} `json:"data"`
	Error      bool   `json:"error"`
	TimeUpdate uint64 `json:"timeUpdated"`
}

// GetRateUsdTomo get usd from api
func (httpFetcher *HTTPFetcher) GetRateUsdTomo() (string, error) {
	var ethPrice string
	url := fmt.Sprintf("%s/token_price?currency=USD", httpFetcher.apiEndpoint)
	b, err := fCommon.HTTPCall(url)
	if err != nil {
		log.Print(err)
		return ethPrice, err
	}
	var tokenPrice TokenPrice
	err = json.Unmarshal(b, &tokenPrice)
	if err != nil {
		log.Println(err)
		return ethPrice, err
	}
	if tokenPrice.Error {
		return ethPrice, errors.New("cannot get token price from api")
	}
	for _, v := range tokenPrice.Data {
		if v.Symbol == common.TOMOSymbol {
			ethPrice = fmt.Sprintf("%.6f", v.Price)
			break
		}
	}
	return ethPrice, nil
}
