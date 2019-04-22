package persister

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/marknguyen85/server-api/tomochain"
)

const (
	STEP_SAVE_RATE      = 10 //1 minute
	MAXIMUM_SAVE_RECORD = 60 //60 records

	INTERVAL_UPDATE_KYBER_ENABLE       = 20
	INTERVAL_UPDATE_MAX_GAS            = 70
	INTERVAL_UPDATE_GAS                = 40
	INTERVAL_UPDATE_RATE_USD           = 610
	INTERVAL_UPDATE_GENERAL_TOKEN_INFO = 3600
	INTERVAL_UPDATE_GET_BLOCKNUM       = 20
	INTERVAL_UPDATE_GET_RATE           = 30
	INTERVAL_UPDATE_DATA_TRACKER       = 310
)

type RamPersister struct {
	mu      sync.RWMutex
	timeRun string

	kyberEnabled      bool
	isNewKyberEnabled bool

	rates     []tomochain.Rate
	isNewRate bool
	updatedAt int64

	latestBlock      string
	isNewLatestBlock bool

	rateUSD      []RateUSD
	rateTOMO     string
	isNewRateUsd bool

	// rateUSDCG      []RateUSD
	// rateTOMOCG      string
	// isNewRateUsdCG bool

	events     []tomochain.EventHistory
	isNewEvent bool

	maxGasPrice      string
	isNewMaxGasPrice bool

	gasPrice      *tomochain.GasPrice
	isNewGasPrice bool

	// ethRate      string
	// isNewEthRate bool

	tokenInfo map[string]*tomochain.TokenGeneralInfo
	// tokenInfoCG map[string]*tomochain.TokenGeneralInfo

	//isNewTokenInfo bool

	marketInfo              map[string]*tomochain.MarketInfo
	last7D                  map[string][]float64
	isNewTrackerData        bool
	numRequestFailedTracker int

	rightMarketInfo map[string]*tomochain.RightMarketInfo
	// rightMarketInfoCG map[string]*tomochain.RightMarketInfo

	isNewMarketInfo bool
	// isNewMarketInfoCG bool
}

func NewRamPersister() (*RamPersister, error) {
	var mu sync.RWMutex
	location, _ := time.LoadLocation("Asia/Bangkok")
	tNow := time.Now().In(location)
	timeRun := fmt.Sprintf("%02d:%02d:%02d %02d-%02d-%d", tNow.Hour(), tNow.Minute(), tNow.Second(), tNow.Day(), tNow.Month(), tNow.Year())

	kyberEnabled := true
	isNewKyberEnabled := true

	rates := []tomochain.Rate{}
	isNewRate := false

	latestBlock := "0"
	isNewLatestBlock := true

	rateUSD := make([]RateUSD, 0)
	rateTOMO := "0"
	isNewRateUsd := true
	// rateUSDCG := make([]RateUSD, 0)
	// rateTOMOCG := "0"
	// isNewRateUsdCG := true

	events := make([]tomochain.EventHistory, 0)
	isNewEvent := true

	maxGasPrice := "50"
	isNewMaxGasPrice := true

	gasPrice := tomochain.GasPrice{}
	isNewGasPrice := true

	// ethRate := "0"
	// isNewEthRate := true

	tokenInfo := map[string]*tomochain.TokenGeneralInfo{}
	// tokenInfoCG := map[string]*tomochain.TokenGeneralInfo{}
	//isNewTokenInfo := true

	marketInfo := map[string]*tomochain.MarketInfo{}
	last7D := map[string][]float64{}
	isNewTrackerData := true

	rightMarketInfo := map[string]*tomochain.RightMarketInfo{}
	// rightMarketInfoCG := map[string]*tomochain.RightMarketInfo{}

	isNewMarketInfo := true
	// isNewMarketInfoCG := true

	persister := &RamPersister{
		mu:                mu,
		timeRun:           timeRun,
		kyberEnabled:      kyberEnabled,
		isNewKyberEnabled: isNewKyberEnabled,
		rates:             rates,
		isNewRate:         isNewRate,
		updatedAt:         0,
		latestBlock:       latestBlock,
		isNewLatestBlock:  isNewLatestBlock,
		rateUSD:           rateUSD,
		rateTOMO:          rateTOMO,
		isNewRateUsd:      isNewRateUsd,
		// rateUSDCG:         rateUSDCG,
		// rateTOMOCG:         rateTOMOCG,
		// isNewRateUsdCG:    isNewRateUsdCG,
		events:           events,
		isNewEvent:       isNewEvent,
		maxGasPrice:      maxGasPrice,
		isNewMaxGasPrice: isNewMaxGasPrice,
		gasPrice:         &gasPrice,
		isNewGasPrice:    isNewGasPrice,
		tokenInfo:        tokenInfo,
		// tokenInfoCG:       tokenInfoCG,
		marketInfo:              marketInfo,
		last7D:                  last7D,
		isNewTrackerData:        isNewTrackerData,
		numRequestFailedTracker: 0,
		rightMarketInfo:         rightMarketInfo,
		// rightMarketInfoCG: rightMarketInfoCG,
		isNewMarketInfo: isNewMarketInfo,
		// isNewMarketInfoCG: isNewMarketInfoCG,
	}
	return persister, nil
}

func (rPersister *RamPersister) SaveGeneralInfoTokens(generalInfo map[string]*tomochain.TokenGeneralInfo) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.tokenInfo = generalInfo
	// rPersister.tokenInfoCG = generalInfoCG
}

func (rPersister *RamPersister) GetTokenInfo() map[string]*tomochain.TokenGeneralInfo {
	rPersister.mu.RLock()
	defer rPersister.mu.RUnlock()
	return rPersister.tokenInfo
}

/////------------------------------
func (rPersister *RamPersister) GetRate() []tomochain.Rate {
	rPersister.mu.RLock()
	defer rPersister.mu.RUnlock()
	return rPersister.rates
}

func (rPersister *RamPersister) GetTimeUpdateRate() int64 {
	rPersister.mu.RLock()
	defer rPersister.mu.RUnlock()
	return rPersister.updatedAt
}

func (rPersister *RamPersister) SetIsNewRate(isNewRate bool) {
	rPersister.mu.RLock()
	defer rPersister.mu.RUnlock()
	// return rPersister.rates
	rPersister.isNewRate = isNewRate
}

func (rPersister *RamPersister) GetIsNewRate() bool {
	rPersister.mu.RLock()
	defer rPersister.mu.RUnlock()
	return rPersister.isNewRate
}

func (rPersister *RamPersister) SaveRate(rates []tomochain.Rate, timestamp int64) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.rates = rates
	if timestamp != 0 {
		rPersister.updatedAt = timestamp
	}
}

//--------------------------------------------------------
func (rPersister *RamPersister) SaveKyberEnabled(enabled bool) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.kyberEnabled = enabled
	rPersister.isNewKyberEnabled = true
}

func (rPersister *RamPersister) SetNewKyberEnabled(isNew bool) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.isNewKyberEnabled = isNew
}

func (rPersister *RamPersister) GetKyberEnabled() bool {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	return rPersister.kyberEnabled
}

func (rPersister *RamPersister) GetNewKyberEnabled() bool {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	return rPersister.isNewKyberEnabled
}

//--------------------------------------------------------

//--------------------------------------------------------

func (rPersister *RamPersister) SetNewMaxGasPrice(isNew bool) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.isNewMaxGasPrice = isNew
	return
}

func (rPersister *RamPersister) SaveMaxGasPrice(maxGasPrice string) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.maxGasPrice = maxGasPrice
	rPersister.isNewMaxGasPrice = true
	return
}
func (rPersister *RamPersister) GetMaxGasPrice() string {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	return rPersister.maxGasPrice
}
func (rPersister *RamPersister) GetNewMaxGasPrice() bool {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	return rPersister.isNewMaxGasPrice
}

//--------------------------------------------------------

//--------------------------------------------------------------

func (rPersister *RamPersister) SaveGasPrice(gasPrice *tomochain.GasPrice) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.gasPrice = gasPrice
	rPersister.isNewGasPrice = true
}
func (rPersister *RamPersister) SetNewGasPrice(isNew bool) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.isNewGasPrice = isNew
}
func (rPersister *RamPersister) GetGasPrice() *tomochain.GasPrice {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	return rPersister.gasPrice
}
func (rPersister *RamPersister) GetNewGasPrice() bool {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	return rPersister.isNewGasPrice
}

//-----------------------------------------------------------

func (rPersister *RamPersister) GetRateUSD() []RateUSD {
	rPersister.mu.RLock()
	defer rPersister.mu.RUnlock()
	return rPersister.rateUSD
}

func (rPersister *RamPersister) GetRateTOMO() string {
	rPersister.mu.RLock()
	defer rPersister.mu.RUnlock()
	return rPersister.rateTOMO
}

func (rPersister *RamPersister) GetIsNewRateUSD() bool {
	rPersister.mu.RLock()
	defer rPersister.mu.RUnlock()
	return rPersister.isNewRateUsd
}

func (rPersister *RamPersister) SaveRateUSD(rateUSDEth string) error {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()

	rates := make([]RateUSD, 0)
	// ratesCG := make([]RateUSD, 0)

	itemRateEth := RateUSD{Symbol: "TOMO", PriceUsd: rateUSDEth}
	// itemRateEthCG := RateUSD{Symbol: "TOMO", PriceUsd: rateUSDEthCG}
	rates = append(rates, itemRateEth)
	// ratesCG = append(ratesCG, itemRateEthCG)
	for _, item := range rPersister.rates {
		if item.Source != "TOMO" {
			priceUsd, err := CalculateRateUSD(item.Rate, rateUSDEth)
			if err != nil {
				log.Print(err)
				rPersister.isNewRateUsd = false
				return nil
			}

			sourceSymbol := item.Source
			if sourceSymbol == "TOMOOS" {
				sourceSymbol = "BQX"
			}
			itemRate := RateUSD{Symbol: sourceSymbol, PriceUsd: priceUsd}
			rates = append(rates, itemRate)
		}
	}

	rPersister.rateUSD = rates
	rPersister.rateTOMO = rateUSDEth
	rPersister.isNewRateUsd = true

	return nil
}

func CalculateRateUSD(rateEther string, rateUSD string) (string, error) {
	bigRateUSD, ok := new(big.Float).SetString(rateUSD)
	if !ok {
		err := errors.New("Cannot convert rate usd of ether to big float")
		return "", err
	}
	bigRateEth, ok := new(big.Float).SetString(rateEther)
	if !ok {
		err := errors.New("Cannot convert rate token-eth to big float")
		return "", err
	}
	i, e := big.NewInt(10), big.NewInt(18)
	i.Exp(i, e, nil)
	weight := new(big.Float).SetInt(i)

	rateUSDBig := new(big.Float).Mul(bigRateUSD, bigRateEth)
	rateUSDNormal := new(big.Float).Quo(rateUSDBig, weight)
	return rateUSDNormal.String(), nil
}

func (rPersister *RamPersister) SetNewRateUSD(isNew bool) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.isNewRateUsd = isNew
}

func (rPersister *RamPersister) GetLatestBlock() string {
	rPersister.mu.RLock()
	defer rPersister.mu.RUnlock()
	return rPersister.latestBlock
}

func (rPersister *RamPersister) SaveLatestBlock(blockNumber string) error {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.latestBlock = blockNumber
	rPersister.isNewLatestBlock = true
	return nil
}

func (rPersister *RamPersister) GetIsNewLatestBlock() bool {
	rPersister.mu.RLock()
	defer rPersister.mu.RUnlock()
	return rPersister.isNewLatestBlock
}

func (rPersister *RamPersister) SetNewLatestBlock(isNew bool) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.isNewLatestBlock = isNew
}

// ----------------------------------------
// return data from kyber tracker

// use this api for 3 infomations change, marketcap, volume
func (rPersister *RamPersister) GetRightMarketData() map[string]*tomochain.RightMarketInfo {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	return rPersister.rightMarketInfo
}

func (rPersister *RamPersister) GetIsNewTrackerData() bool {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	return rPersister.isNewTrackerData
}

func (rPersister *RamPersister) SetIsNewTrackerData(isNewTrackerData bool) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.isNewTrackerData = isNewTrackerData
	rPersister.numRequestFailedTracker = 0
}

func (rPersister *RamPersister) GetLast7D(listTokens string) map[string][]float64 {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	tokens := strings.Split(listTokens, "-")
	result := make(map[string][]float64)
	for _, symbol := range tokens {
		if rPersister.last7D[symbol] != nil {
			result[symbol] = rPersister.last7D[symbol]
		}
	}
	return result
}

func (rPersister *RamPersister) SaveMarketData(marketRate map[string]*tomochain.Rates, mapTokenInfo map[string]*tomochain.TokenGeneralInfo, tokens map[string]tomochain.Token) {
	lastSevenDays := map[string][]float64{}
	newResult := map[string]*tomochain.RightMarketInfo{}
	if len(mapTokenInfo) == 0 {
		rPersister.mu.RLock()
		mapTokenInfo = rPersister.tokenInfo
		rPersister.mu.RUnlock()
	}
	for symbol := range tokens {
		dataSevenDays := []float64{}
		rightMarketInfo := &tomochain.RightMarketInfo{}
		rateInfo := marketRate[symbol]
		if rateInfo != nil {
			dataSevenDays = rateInfo.P
			rightMarketInfo.Rate = &rateInfo.R
		}
		if tokenInfo := mapTokenInfo[symbol]; tokenInfo != nil {
			rightMarketInfo.Quotes = tokenInfo.Quotes
			rightMarketInfo.Change24H = tokenInfo.Change24H
		}

		if rateInfo == nil && rightMarketInfo.Quotes == nil {
			continue
		}

		newResult[symbol] = rightMarketInfo
		lastSevenDays[symbol] = dataSevenDays
	}

	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.last7D = lastSevenDays
	rPersister.rightMarketInfo = newResult
}

func (rPersister *RamPersister) SetIsNewMarketInfo(isNewMarketInfo bool) {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.isNewMarketInfo = isNewMarketInfo
}

func (rPersister *RamPersister) GetIsNewMarketInfo() bool {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	return rPersister.isNewMarketInfo
}

func (rPersister *RamPersister) GetTimeVersion() string {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	return rPersister.timeRun
}

func (rPersister *RamPersister) IsFailedToFetchTracker() bool {
	rPersister.mu.Lock()
	defer rPersister.mu.Unlock()
	rPersister.numRequestFailedTracker++
	if rPersister.numRequestFailedTracker > 12 {
		return true
	}
	return false
}
