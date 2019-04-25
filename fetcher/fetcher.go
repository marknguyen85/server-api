package fetcher

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/big"
	"sync"

	// "strconv"
	"time"

	"github.com/marknguyen85/server-api/common"
	"github.com/marknguyen85/server-api/tomochain"
	// nFetcher "github.com/marknguyen85/server-api/fetcher/normal-fetcher"
)

const (
	TOMO_TO_WEI = 1000000000000000000
	MIN_TOMO    = 0.1
	KEY         = "chaintexsecret"

	timeW8Req = 500
)

type Connection struct {
	Endpoint string `json:"endPoint"`
	Type     string `json:"type"`
	Apikey   string `json:"api_key"`
}

type InfoData struct {
	mu         *sync.RWMutex
	ApiUsd     string               `json:"api_usd"`
	CoinMarket []string             `json:"coin_market"`
	TokenAPI   []tomochain.TokenAPI `json:"tokens"`

	Tokens        map[string]tomochain.Token
	BackupTokens  map[string]tomochain.Token
	TokenPriority map[string]tomochain.Token
	Connections   []Connection `json:"connections"`

	Network    string `json:"network"`
	NetworkAbi string
	TradeTopic string `json:"trade_topic"`

	TomoAddress string
	TomoSymbol  string

	AverageBlockTime int64 `json:"averageBlockTime"`

	GasStationEndpoint string `json:"gasstation_endpoint"`
	APIEndpoint        string `json:"api_endpoint"`
	ConfigEndpoint     string `json:"config_endpoint"`
	UserStatsEndpoint  string `json:"user_stats_endpoint"`
}

//GetListToken return list tokens supported
func (infoData *InfoData) GetListToken() map[string]tomochain.Token {
	infoData.mu.RLock()
	defer infoData.mu.RUnlock()
	return infoData.Tokens
}

//UpdateByBackupToken func
func (infoData *InfoData) UpdateByBackupToken() {
	infoData.mu.RLock()
	backupToken := infoData.BackupTokens
	infoData.mu.RUnlock()
	listPriority := make(map[string]tomochain.Token)
	for _, t := range backupToken {
		if t.Priority {
			listPriority[t.Symbol] = t
		}
	}

	infoData.mu.Lock()
	defer infoData.mu.Unlock()
	infoData.Tokens = backupToken
	infoData.TokenPriority = listPriority
}

//UpdateListToken func
func (infoData *InfoData) UpdateListToken(tokens, tokenPriority map[string]tomochain.Token) {
	infoData.mu.Lock()
	defer infoData.mu.Unlock()
	infoData.Tokens = tokens
	infoData.BackupTokens = tokens
	infoData.TokenPriority = tokenPriority
}

//GetTokenAPI func
func (infoData *InfoData) GetTokenAPI() []tomochain.TokenAPI {
	infoData.mu.RLock()
	defer infoData.mu.RUnlock()
	return infoData.TokenAPI
}

//GetListTokenPriority func
func (infoData *InfoData) GetListTokenPriority() map[string]tomochain.Token {
	infoData.mu.RLock()
	defer infoData.mu.RUnlock()
	return infoData.TokenPriority
}

//Fetcher struct
type Fetcher struct {
	info         *InfoData
	tomochain    *TomoChain
	fetIns       []FetcherInterface
	marketFetIns MarketFetcherInterface
	httpFetcher  *HTTPFetcher
}

//GetNumTokens func
func (fetcher *Fetcher) GetNumTokens() int {
	listTokens := fetcher.GetListToken()
	return len(listTokens)
}

//NewFetcher func
func NewFetcher(chainTexENV string) (*Fetcher, error) {
	var file []byte
	var err error
	switch chainTexENV {
	case "staging":
		file, err = ioutil.ReadFile("env/staging.json")
		if err != nil {
			log.Print(err)
			return nil, err
		}
		break
	case "production":
		file, err = ioutil.ReadFile("env/production.json")
		if err != nil {
			log.Print(err)
			return nil, err
		}
		break
	default:
		file, err = ioutil.ReadFile("env/testnet.json")
		if err != nil {
			log.Print(err)
			return nil, err
		}
		break
	}

	mu := &sync.RWMutex{}

	infoData := InfoData{
		mu:          mu,
		TomoAddress: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		TomoSymbol:  "TOMO",
		NetworkAbi:  `[{"constant":false,"inputs":[{"name":"alerter","type":"address"}],"name":"removeAlerter","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"token","type":"address"},{"name":"srcQty","type":"uint256"}],"name":"getExpectedFeeRate","outputs":[{"name":"expectedRate","type":"uint256"},{"name":"slippageRate","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"token","type":"address"},{"name":"srcAmount","type":"uint256"},{"name":"minConversionRate","type":"uint256"}],"name":"swapTokenToTomo","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"enabled","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"pendingAdmin","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getOperators","outputs":[{"name":"","type":"address[]"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"src","type":"address"},{"name":"srcAmount","type":"uint256"},{"name":"destAddress","type":"address"}],"name":"payTxFeeFast","outputs":[{"name":"","type":"uint256"}],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"name":"token","type":"address"},{"name":"amount","type":"uint256"},{"name":"sendTo","type":"address"}],"name":"withdrawToken","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"maxGasPrice","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"newAlerter","type":"address"}],"name":"addAlerter","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_networkContract","type":"address"}],"name":"setNetworkContract","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"payFeeCallers","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"user","type":"address"}],"name":"getUserCapInWei","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"src","type":"address"},{"name":"srcAmount","type":"uint256"},{"name":"dest","type":"address"},{"name":"minConversionRate","type":"uint256"}],"name":"swapTokenToToken","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"newAdmin","type":"address"}],"name":"transferAdmin","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"claimAdmin","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"newAdmin","type":"address"}],"name":"transferAdminQuickly","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getAlerters","outputs":[{"name":"","type":"address[]"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"src","type":"address"},{"name":"dest","type":"address"},{"name":"srcQty","type":"uint256"}],"name":"getExpectedRate","outputs":[{"name":"expectedRate","type":"uint256"},{"name":"slippageRate","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"user","type":"address"},{"name":"token","type":"address"}],"name":"getUserCapInTokenWei","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"newOperator","type":"address"}],"name":"addOperator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"src","type":"address"},{"name":"srcAmount","type":"uint256"},{"name":"dest","type":"address"},{"name":"destAddress","type":"address"},{"name":"maxDestAmount","type":"uint256"},{"name":"minConversionRate","type":"uint256"},{"name":"walletId","type":"address"}],"name":"swap","outputs":[{"name":"","type":"uint256"}],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"name":"operator","type":"address"}],"name":"removeOperator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"field","type":"bytes32"}],"name":"info","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"caller","type":"address"},{"name":"add","type":"bool"}],"name":"addPayFeeCaller","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"token","type":"address"},{"name":"minConversionRate","type":"uint256"}],"name":"swapTomoToToken","outputs":[{"name":"","type":"uint256"}],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"name":"src","type":"address"},{"name":"srcAmount","type":"uint256"},{"name":"dest","type":"address"},{"name":"destAddress","type":"address"},{"name":"maxDestAmount","type":"uint256"},{"name":"minConversionRate","type":"uint256"},{"name":"walletId","type":"address"}],"name":"trade","outputs":[{"name":"","type":"uint256"}],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"name":"amount","type":"uint256"},{"name":"sendTo","type":"address"}],"name":"withdrawEther","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"token","type":"address"},{"name":"user","type":"address"}],"name":"getBalance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"networkContract","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"admin","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"src","type":"address"},{"name":"srcAmount","type":"uint256"},{"name":"destAddress","type":"address"},{"name":"maxDestAmount","type":"uint256"},{"name":"minConversionRate","type":"uint256"}],"name":"payTxFee","outputs":[{"name":"","type":"uint256"}],"payable":true,"stateMutability":"payable","type":"function"},{"inputs":[{"name":"_admin","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"name":"caller","type":"address"},{"indexed":false,"name":"add","type":"bool"}],"name":"AddPayFeeCaller","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"trader","type":"address"},{"indexed":false,"name":"src","type":"address"},{"indexed":false,"name":"dest","type":"address"},{"indexed":false,"name":"actualSrcAmount","type":"uint256"},{"indexed":false,"name":"actualDestAmount","type":"uint256"}],"name":"ExecuteTrade","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"newNetworkContract","type":"address"},{"indexed":false,"name":"oldNetworkContract","type":"address"}],"name":"NetworkSet","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"token","type":"address"},{"indexed":false,"name":"amount","type":"uint256"},{"indexed":false,"name":"sendTo","type":"address"}],"name":"TokenWithdraw","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"amount","type":"uint256"},{"indexed":false,"name":"sendTo","type":"address"}],"name":"EtherWithdraw","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"pendingAdmin","type":"address"}],"name":"TransferAdminPending","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"newAdmin","type":"address"},{"indexed":false,"name":"previousAdmin","type":"address"}],"name":"AdminClaimed","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"newAlerter","type":"address"},{"indexed":false,"name":"isAdd","type":"bool"}],"name":"AlerterAdded","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"newOperator","type":"address"},{"indexed":false,"name":"isAdd","type":"bool"}],"name":"OperatorAdded","type":"event"}]`,
	}
	err = json.Unmarshal(file, &infoData)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	listToken := make(map[string]tomochain.Token)
	listBackup := make(map[string]tomochain.Token)
	infoData.Tokens = listToken
	for _, t := range infoData.TokenAPI {
		listBackup[t.Symbol] = tomochain.TokenAPIToToken(t)
	}
	infoData.BackupTokens = listBackup

	fetIns := make([]FetcherInterface, 0)
	for _, connection := range infoData.Connections {
		newFetcher, err := NewFetcherIns(connection.Type, connection.Endpoint, connection.Apikey)
		if err != nil {
			log.Print(err)
		} else {
			fetIns = append(fetIns, newFetcher)
		}
	}

	marketFetcherIns := NewMarketFetcherInterface()

	httpFetcher := NewHTTPFetcher(infoData.ConfigEndpoint, infoData.GasStationEndpoint, infoData.APIEndpoint)

	tomochain, err := NewTomoChain(infoData.Network, infoData.NetworkAbi, infoData.TradeTopic,
		infoData.AverageBlockTime)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	fetcher := &Fetcher{
		info:         &infoData,
		tomochain:    tomochain,
		fetIns:       fetIns,
		marketFetIns: marketFetcherIns,
		httpFetcher:  httpFetcher,
	}

	return fetcher, nil
}

//TryUpdateListToken func
func (fetcher *Fetcher) TryUpdateListToken() error {
	var err error
	for i := 0; i < 3; i++ {
		err = fetcher.UpdateListToken()
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}
		return nil
	}
	fetcher.info.UpdateByBackupToken()
	return nil
}

//UpdateListToken func
func (fetcher *Fetcher) UpdateListToken() error {
	var (
		err    error
		result []tomochain.Token
	)
	result, err = fetcher.httpFetcher.GetListToken()
	if err != nil {
		log.Println(err)
		return err
	}
	listToken := make(map[string]tomochain.Token)
	listPriority := make(map[string]tomochain.Token)
	for _, token := range result {
		if token.DelistTime == 0 || uint64(time.Now().UTC().Unix()) <= TIME_TO_DELETE+token.DelistTime {
			tokenID := token.Symbol
			if token.TokenID != "" {
				tokenID = token.TokenID
			}
			newToken := token
			newToken.TokenID = tokenID
			listToken[tokenID] = newToken
			if token.Priority {
				listPriority[tokenID] = newToken
			}
		}
	}
	fetcher.info.UpdateListToken(listToken, listPriority)
	return nil
}

// GetListTokenAPI api to get config token
func (fetcher *Fetcher) GetListTokenAPI() []tomochain.TokenAPI {
	return fetcher.info.GetTokenAPI()
}

// GetListToken return map token with key is token ID
func (fetcher *Fetcher) GetListToken() map[string]tomochain.Token {
	return fetcher.info.GetListToken()
}

// GetListTokenPriority return map token with key is token ID
func (fetcher *Fetcher) GetListTokenPriority() map[string]tomochain.Token {
	return fetcher.info.GetListTokenPriority()
}

//GetGeneralInfoTokens func
func (fetcher *Fetcher) GetGeneralInfoTokens() map[string]*tomochain.TokenGeneralInfo {
	generalInfo := map[string]*tomochain.TokenGeneralInfo{}
	listTokens := fetcher.GetListToken()
	for _, token := range listTokens {
		if token.CGId != "" {
			result, err := fetcher.marketFetIns.GetGeneralInfo(token.CGId)
			time.Sleep(5 * time.Second)
			if err != nil {
				log.Print(err)
				continue
			}
			generalInfo[token.TokenID] = result
		}
	}

	return generalInfo
}

//GetRateUsdTomo func
func (fetcher *Fetcher) GetRateUsdTomo() (string, error) {
	rateUsd, err := fetcher.marketFetIns.GetRateUsdTomo()
	//rateUsd, err := fetcher.httpFetcher.GetRateUsdTomo()
	if err != nil {
		log.Print(err)
		return "", err
	}
	return rateUsd, nil
}

//GetGasPrice func
func (fetcher *Fetcher) GetGasPrice() (*tomochain.GasPrice, error) {
	result, err := fetcher.httpFetcher.GetGasPrice()
	if err != nil {
		log.Print(err)
		return nil, errors.New("Cannot get gas price")
	}
	return result, nil
}

//GetGeneralInfoTokens func
func (fetcher *Fetcher) GetMaxGasPrice() (string, error) {
	dataAbi, err := fetcher.tomochain.EncodeMaxGasPrice()
	if err != nil {
		log.Print(err)
		return "", err
	}
	for _, fetIns := range fetcher.fetIns {
		result, err := fetIns.TomoCall(fetcher.info.Network, dataAbi)
		if err != nil {
			log.Print(err)
			continue
		}
		gasPrice, err := fetcher.tomochain.ExtractMaxGasPrice(result)
		if err != nil {
			log.Print(err)
			continue
		}
		return gasPrice, nil
	}
	return "", errors.New("Cannot get gas price")
}

//GetGeneralInfoTokens func
func (fetcher *Fetcher) CheckChainTeXEnable() (bool, error) {
	dataAbi, err := fetcher.tomochain.EncodeChainTeXEnable()
	if err != nil {
		log.Print(err)
		return false, err
	}
	for _, fetIns := range fetcher.fetIns {
		result, err := fetIns.TomoCall(fetcher.info.Network, dataAbi)
		if err != nil {
			log.Print(err)
			continue
		}
		enabled, err := fetcher.tomochain.ExtractEnabled(result)
		if err != nil {
			log.Print(err)
			continue
		}
		return enabled, nil
	}
	return false, errors.New("Cannot check kyber enable")
}

func getAmountInWei(amount float64) *big.Int {
	amountFloat := big.NewFloat(amount)
	ethFloat := big.NewFloat(TOMO_TO_WEI)
	weiFloat := big.NewFloat(0).Mul(amountFloat, ethFloat)
	amoutInt, _ := weiFloat.Int(nil)
	return amoutInt
}

func getAmountTokenWithMinTOMO(rate *big.Int, decimal int) *big.Int {
	rFloat := big.NewFloat(0).SetInt(rate)
	ethFloat := big.NewFloat(TOMO_TO_WEI)
	amoutnToken1TOMO := rFloat.Quo(rFloat, ethFloat)
	minAmountWithMinTOMO := amoutnToken1TOMO.Mul(amoutnToken1TOMO, big.NewFloat(MIN_TOMO))
	decimalWei := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)
	amountWithDecimal := big.NewFloat(0).Mul(minAmountWithMinTOMO, big.NewFloat(0).SetInt(decimalWei))
	amountInt, _ := amountWithDecimal.Int(nil)
	return amountInt
}

func tokenWei(decimal int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)
}

func getMapRates(rates []tomochain.Rate) map[string]tomochain.Rate {
	m := make(map[string]tomochain.Rate)
	for _, r := range rates {
		m[r.Dest] = r
	}
	return m
}

//FetchRate7dData func
func (fetcher *Fetcher) FetchRate7dData() (map[string]*tomochain.Rates, error) {
	result, err := fetcher.httpFetcher.GetRate7dData()
	if err != nil {
		log.Print(err)
		// continue
		return nil, errors.New("Cannot get data from tracker")
	}
	return result, nil
}

// GetRate get full rate of list token
func (fetcher *Fetcher) GetRate(currentRate []tomochain.Rate, isNewRate bool, mapToken map[string]tomochain.Token, fallback bool) ([]tomochain.Rate, error) {
	var (
		rates []tomochain.Rate
		err   error
	)
	if !isNewRate {
		initRate := fetcher.getInitRate(mapToken)
		currentRate = initRate
	}
	sourceArr, sourceSymbolArr, destArr, destSymbolArr, amountArr := fetcher.makeDataGetRate(mapToken, currentRate)
	rates, err = fetcher.runFetchRate(sourceArr, destArr, sourceSymbolArr, destSymbolArr, amountArr)

	if err != nil && fallback {
		log.Println("cannot get rate from network proxy, change to get from network")
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return rates, nil
}

//runFetchRate func
func (fetcher *Fetcher) runFetchRate(sourceArr, destArr, sourceSymbolArr, destSymbolArr []string, amountArr []*big.Int) ([]tomochain.Rate, error) {
	var (
		tokenNum = len(sourceArr)
		rates    []tomochain.Rate
	)

	for i := 0; i < tokenNum; i++ {
		var (
			dataAbi string
			err     error
			rate    tomochain.Rate
		)
		dataAbi, err = fetcher.tomochain.EncodeRateData(sourceArr[i], destArr[i], amountArr[i])

		if err != nil {
			log.Print(err)
		} else {
			rate, err = fetcher.GetRateFromAbi(dataAbi, sourceSymbolArr[i], destSymbolArr[i])
			if err != nil {
				log.Print(err)
			}

			rates = append(rates, rate)
		}
	}

	return rates, nil
}

//GetRateFromAbi func get rate from abi string
func (fetcher *Fetcher) GetRateFromAbi(dataAbi string, fromSymbol string, toSymbol string) (tomochain.Rate, error) {
	var rate tomochain.Rate

	for _, fetIns := range fetcher.fetIns {
		if fetIns.GetTypeName() == "tomoscan" {
			continue
		}

		result, err := fetIns.GetRate(fetcher.info.Network, dataAbi)
		if err != nil {
			log.Print(err)
			continue
		}

		rate, err = fetcher.tomochain.ExtractRateData(result, fromSymbol, toSymbol)
		if err != nil {
			log.Print(err)
			continue
		}
		return rate, nil
	}

	return rate, nil
}

//makeDataGetRate func
func (fetcher *Fetcher) makeDataGetRate(listTokens map[string]tomochain.Token, rates []tomochain.Rate) ([]string, []string, []string, []string, []*big.Int) {
	sourceAddr := make([]string, 0)
	sourceSymbol := make([]string, 0)
	destAddr := make([]string, 0)
	destSymbol := make([]string, 0)
	amount := make([]*big.Int, 0)
	amountTOMO := make([]*big.Int, 0)
	ethSymbol := common.TOMOSymbol
	ethAddr := common.TOMOAddr
	minAmountTOMO := getAmountInWei(MIN_TOMO)
	mapRate := getMapRates(rates)

	for _, t := range listTokens {
		decimal := t.Decimal
		amountToken := tokenWei(t.Decimal / 2)
		if t.Symbol == ethSymbol {
			amountToken = minAmountTOMO
		} else {
			if rate, ok := mapRate[t.Symbol]; ok {
				r := new(big.Int)
				r.SetString(rate.Rate, 10)
				if r.Cmp(new(big.Int)) != 0 {
					amountToken = getAmountTokenWithMinTOMO(r, decimal)
				}
			}
		}
		sourceAddr = append(sourceAddr, t.Address)
		destAddr = append(destAddr, ethAddr)
		sourceSymbol = append(sourceSymbol, t.Symbol)
		destSymbol = append(destSymbol, ethSymbol)
		amount = append(amount, amountToken)
		amountTOMO = append(amountTOMO, minAmountTOMO)
	}

	sourceArr := append(sourceAddr, destAddr...)
	sourceSymbolArr := append(sourceSymbol, destSymbol...)
	destArr := append(destAddr, sourceAddr...)
	destSymbolArr := append(destSymbol, sourceSymbol...)
	amountArr := append(amount, amountTOMO...)

	return sourceArr, sourceSymbolArr, destArr, destSymbolArr, amountArr
}

//getInitRate func
func (fetcher *Fetcher) getInitRate(listTokens map[string]tomochain.Token) []tomochain.Rate {
	tomoSymbol := common.TOMOSymbol
	tomoAddr := common.TOMOAddr
	minAmountTOMO := getAmountInWei(MIN_TOMO)

	srcArr := []string{}
	destArr := []string{}
	srcSymbolArr := []string{}
	destSymbolArr := []string{}
	amountArr := []*big.Int{}
	for _, t := range listTokens {
		if t.Symbol == tomoSymbol {
			continue
		}
		srcArr = append(srcArr, tomoAddr)
		destArr = append(destArr, t.Address)
		srcSymbolArr = append(srcSymbolArr, tomoSymbol)
		destSymbolArr = append(destSymbolArr, t.Symbol)
		amountArr = append(amountArr, minAmountTOMO)
	}

	initRate, _ := fetcher.runFetchRate(srcArr, destArr, srcSymbolArr, destSymbolArr, amountArr)
	return initRate
}

//queryRateBlockchain func
func (fetcher *Fetcher) queryRateBlockchain(fromAddr, toAddr, fromSymbol, toSymbol string, amount *big.Int) (tomochain.Rate, error) {
	var rate tomochain.Rate
	dataAbi, err := fetcher.tomochain.EncodeRateData(fromAddr, toAddr, amount)
	if err != nil {
		log.Print(err)
		return rate, err
	}

	for _, fetIns := range fetcher.fetIns {
		if fetIns.GetTypeName() == "tomoscan" {
			continue
		}
		result, err := fetIns.GetRate(fetcher.info.Network, dataAbi)
		if err != nil {
			log.Print(err)
			continue
		}
		rate, err := fetcher.tomochain.ExtractRateData(result, fromSymbol, toSymbol)
		if err != nil {
			log.Print(err)
			continue
		}
		return rate, nil
	}
	return rate, errors.New("cannot get rate")
}
