package mFetcher

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/marknguyen85/server-api/tomochain"
	fCommon "github.com/marknguyen85/server-api/fetcher/fetcher-common"
)

type CMCFetcher struct {
	APIV1      string
	APIV2      string
	typeMarket string
}

func NewCMCFetcher() *CMCFetcher {
	return &CMCFetcher{
		APIV1:      "https://api.coinmarketcap.com/v1",
		APIV2:      "https://api.coinmarketcap.com/v2",
		typeMarket: "cmc",
	}
}

func (self *CMCFetcher) GetRateUsdEther() (string, error) {
	// typeMarket := self.typeMarket
	url := self.APIV1 + "/ticker/tomochain"
	b, err := fCommon.HTTPCall(url)
	if err != nil {
		log.Print(err)
		return "", err
	}
	rateItem := make([]tomochain.RateUSD, 0)
	err = json.Unmarshal(b, &rateItem)
	if err != nil {
		log.Print(err)
		return "", err
	}
	return rateItem[0].PriceUsd, nil
}

func (self *CMCFetcher) GetGeneralInfo(usdId string) (*tomochain.TokenGeneralInfo, error) {
	url := self.APIV2 + "/ticker/" + usdId + "/?convert=TOMO"
	b, err := fCommon.HTTPCall(url)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	tokenItem := map[string]tomochain.TokenGeneralInfo{}
	err = json.Unmarshal(b, &tokenItem)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	if data, ok := tokenItem["data"]; ok {
		data.MarketCap = data.Quotes["TOMO"].MarketCap
		return &data, nil
	}
	err = errors.New("Cannot find data key in return quotes of ticker")
	log.Print(err)
	return nil, err
}

// func (self *CMCFetcher) GetTypeMarket() string {
// 	return self.typeMarket
// }
