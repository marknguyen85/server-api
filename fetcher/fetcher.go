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

	"github.com/marknguyen85/server-api/tomochain"
	// nFetcher "github.com/marknguyen85/server-api/fetcher/normal-fetcher"
)

const (
	TOMO_TO_WEI = 1000000000000000000
	MIN_TOMO    = 0.1
	KEY         = "kybersecret"

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

func (self *InfoData) GetListToken() map[string]tomochain.Token {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.Tokens
}

func (self *InfoData) UpdateByBackupToken() {
	self.mu.RLock()
	backupToken := self.BackupTokens
	self.mu.RUnlock()
	listPriority := make(map[string]tomochain.Token)
	for _, t := range backupToken {
		if t.Priority {
			listPriority[t.Symbol] = t
		}
	}
	self.mu.Lock()
	defer self.mu.Unlock()
	self.Tokens = backupToken
	self.TokenPriority = listPriority
}

func (self *InfoData) UpdateListToken(tokens, tokenPriority map[string]tomochain.Token) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.Tokens = tokens
	self.BackupTokens = tokens
	self.TokenPriority = tokenPriority
}

func (self *InfoData) GetTokenAPI() []tomochain.TokenAPI {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.TokenAPI
}

func (self *InfoData) GetListTokenPriority() map[string]tomochain.Token {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.TokenPriority
}

type Fetcher struct {
	info         *InfoData
	tomochain    *TomoChain
	fetIns       []FetcherInterface
	marketFetIns MarketFetcherInterface
	httpFetcher  *HTTPFetcher
}

func (self *Fetcher) GetNumTokens() int {
	listTokens := self.GetListToken()
	return len(listTokens)
}

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
		log.Print("===================connection", connection)
		newFetcher, err := NewFetcherIns(connection.Type, connection.Endpoint, connection.Apikey)
		if err != nil {
			log.Print(err)
		} else {
			fetIns = append(fetIns, newFetcher)
		}
	}

	log.Print("===================fetIns", fetIns)

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

func (self *Fetcher) TryUpdateListToken() error {
	var err error
	for i := 0; i < 3; i++ {
		err = self.UpdateListToken()
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}
		return nil
	}
	self.info.UpdateByBackupToken()
	return nil
}

func (self *Fetcher) UpdateListToken() error {
	var (
		err    error
		result []tomochain.Token
	)
	result, err = self.httpFetcher.GetListToken()
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
	self.info.UpdateListToken(listToken, listPriority)
	return nil
}

// api to get config token
func (self *Fetcher) GetListTokenAPI() []tomochain.TokenAPI {
	return self.info.GetTokenAPI()
}

// GetListToken return map token with key is token ID
func (self *Fetcher) GetListToken() map[string]tomochain.Token {
	return self.info.GetListToken()
}

// GetListToken return map token with key is token ID
func (self *Fetcher) GetListTokenPriority() map[string]tomochain.Token {
	return self.info.GetListTokenPriority()
}

func (self *Fetcher) GetGeneralInfoTokens() map[string]*tomochain.TokenGeneralInfo {
	generalInfo := map[string]*tomochain.TokenGeneralInfo{}
	listTokens := self.GetListToken()
	for _, token := range listTokens {
		if token.CGId != "" {
			result, err := self.marketFetIns.GetGeneralInfo(token.CGId)
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

func (self *Fetcher) GetRateUsdTomo() (string, error) {
	rateUsd, err := self.marketFetIns.GetRateUsdTomo()
	//rateUsd, err := self.httpFetcher.GetRateUsdTomo()
	if err != nil {
		log.Print(err)
		return "", err
	}
	return rateUsd, nil
}

func (self *Fetcher) GetGasPrice() (*tomochain.GasPrice, error) {
	result, err := self.httpFetcher.GetGasPrice()
	if err != nil {
		log.Print(err)
		return nil, errors.New("Cannot get gas price")
	}
	return result, nil
}

func (self *Fetcher) GetMaxGasPrice() (string, error) {
	dataAbi, err := self.tomochain.EncodeMaxGasPrice()
	if err != nil {
		log.Print(err)
		return "", err
	}
	for _, fetIns := range self.fetIns {
		result, err := fetIns.EthCall(self.info.Network, dataAbi)
		if err != nil {
			log.Print(err)
			continue
		}
		gasPrice, err := self.tomochain.ExtractMaxGasPrice(result)
		if err != nil {
			log.Print(err)
			continue
		}
		return gasPrice, nil
	}
	return "", errors.New("Cannot get gas price")
}

func (self *Fetcher) CheckKyberEnable() (bool, error) {
	dataAbi, err := self.tomochain.EncodeKyberEnable()
	if err != nil {
		log.Print(err)
		return false, err
	}
	for _, fetIns := range self.fetIns {
		result, err := fetIns.EthCall(self.info.Network, dataAbi)
		if err != nil {
			log.Print(err)
			continue
		}
		enabled, err := self.tomochain.ExtractEnabled(result)
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

func (self *Fetcher) FetchRate7dData() (map[string]*tomochain.Rates, error) {
	result, err := self.httpFetcher.GetRate7dData()
	if err != nil {
		log.Print(err)
		// continue
		return nil, errors.New("Cannot get data from tracker")
	}
	return result, nil
}
