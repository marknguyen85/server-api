package mFetcher

import (
	"encoding/json"
	"fmt"
	"log"

	fCommon "github.com/marknguyen85/server-api/fetcher/fetcher-common"
	"github.com/marknguyen85/server-api/tomochain"
)

type CGFetcher struct {
	API        string
	typeMarket string
}

func NewCGFetcher() *CGFetcher {
	return &CGFetcher{
		API:        "https://api.coingecko.com/api/v3",
		typeMarket: "coingecko",
	}
}

func (self *CGFetcher) GetRateUsdEther() (string, error) {
	// typeMarket := self.typeMarket
	url := self.API + "/coins/tomochain"
	b, err := fCommon.HTTPCall(url)
	if err != nil {
		log.Print(err)
		return "", err
	}
	rateItem := tomochain.RateUSDCG{}
	err = json.Unmarshal(b, &rateItem)
	if err != nil {
		log.Print(err)
		return "", err
	}
	rateString := fmt.Sprintf("%.6f", rateItem.MarketData.CurrentPrice.USD)
	return rateString, nil
}

func (self *CGFetcher) GetGeneralInfo(coinID string) (*tomochain.TokenGeneralInfo, error) {
	url := fmt.Sprintf("%s/coins/%s?tickers=false&community_data=false&developer_data=false&sparkline=false", self.API, coinID)
	b, err := fCommon.HTTPCall(url)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	tokenItem := tomochain.TokenInfoCoinGecko{}
	err = json.Unmarshal(b, &tokenItem)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	tokenGenalInfo := tokenItem.ToTokenInfoCMC()
	return &tokenGenalInfo, nil
}

// func (self *CGFetcher) GetTypeMarket() string {
// 	return self.typeMarket
// }
