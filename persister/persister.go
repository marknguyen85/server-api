package persister

import (
	"github.com/marknguyen85/server-api/tomochain"
)

type RateUSD struct {
	Symbol   string `json:"symbol"`
	PriceUsd string `json:"price_usd"`
}

type Persister interface {
	GetRate() []tomochain.Rate
	GetIsNewRate() bool
	SetIsNewRate(bool)
	GetTimeUpdateRate() int64

	SaveRate([]tomochain.Rate, int64)

	SaveGeneralInfoTokens(map[string]*tomochain.TokenGeneralInfo)
	GetTokenInfo() map[string]*tomochain.TokenGeneralInfo

	GetRateUSD() []RateUSD
	GetRateTOMO() string
	GetIsNewRateUSD() bool
	SaveRateUSD(string) error
	SetNewRateUSD(bool)

	SaveMarketData(rates map[string]*tomochain.Rates, mapTokenInfo map[string]*tomochain.TokenGeneralInfo, tokens map[string]tomochain.Token)
	GetRightMarketData() map[string]*tomochain.RightMarketInfo
	// GetRightMarketDataCG() map[string]*tomochain.RightMarketInfo
	GetLast7D(listTokens string) map[string][]float64
	GetIsNewTrackerData() bool
	SetIsNewTrackerData(isNewTrackerData bool)
	SetIsNewMarketInfo(isNewMarketInfo bool)
	GetIsNewMarketInfo() bool
	// GetIsNewMarketInfoCG() bool
	GetTimeVersion() string

	IsFailedToFetchTracker() bool
}

//var transactionPersistent = models.NewTransactionPersister()

func NewPersister(name string) (Persister, error) {
	Persister, err := NewRamPersister()
	return Persister, err
}
