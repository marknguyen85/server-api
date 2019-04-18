package fetcher

import (
	"github.com/marknguyen85/server-api/tomochain"
	mFetcher "github.com/marknguyen85/server-api/fetcher/market-fetcher"
)

type MarketFetcherInterface interface {
	GetRateUsdEther() (string, error)
	GetGeneralInfo(string) (*tomochain.TokenGeneralInfo, error)
	// GetTypeMarket() string
}

//var transactionPersistent = models.NewTransactionPersister()

func NewMarketFetcherInterface() MarketFetcherInterface {
	// var fetcher FetcherNormalInterface
	// switch typeName {
	// case "cmc":
	// 	fetcher = nFetcher.NewCMCFetcher()
	// 	break
	// case "coingecko":
	// 	fetcher = nFetcher.NewCGFetcher()
	// 	break
	// }
	// return fetcher
	return mFetcher.NewCGFetcher()
}
