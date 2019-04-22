package fetcher

import (
	mFetcher "github.com/marknguyen85/server-api/fetcher/market-fetcher"
	"github.com/marknguyen85/server-api/tomochain"
)

type MarketFetcherInterface interface {
	GetRateUsdTomo() (string, error)
	GetGeneralInfo(string) (*tomochain.TokenGeneralInfo, error)
}

func NewMarketFetcherInterface() MarketFetcherInterface {
	return mFetcher.NewCGFetcher()
}
