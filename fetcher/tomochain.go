package fetcher

import (
	"errors"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/marknguyen85/server-api/tomochain"
	"github.com/tomochain/tomochain/accounts/abi"
	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/common/hexutil"
)

type RateNetwork struct {
	ExpectedRate *big.Int `json:"expectedRate"`
	SlippageRate *big.Int `json:"slippageRate"`
}

type RateWrapper struct {
	ExpectedRate []*big.Int `json:"expectedRate"`
	SlippageRate []*big.Int `json:"slippageRate"`
}

type TomoChain struct {
	network          string
	networkAbi       abi.ABI
	tradeTopic       string
	wrapper          string
	wrapperAbi       abi.ABI
	averageBlockTime int64
}

func NewTomoChain(network string, networkAbiStr string, tradeTopic string, wrapper string, wrapperAbiStr string, averageBlockTime int64) (*TomoChain, error) {

	networkAbi, err := abi.JSON(strings.NewReader(networkAbiStr))
	if err != nil {
		log.Print(err)
		return nil, err
	}

	wrapperAbi, err := abi.JSON(strings.NewReader(wrapperAbiStr))
	if err != nil {
		log.Print(err)
		return nil, err
	}

	tomochain := &TomoChain{
		network, networkAbi, tradeTopic, wrapper, wrapperAbi, averageBlockTime,
	}

	return tomochain, nil
}

func getOrAmount(amount *big.Int) *big.Int {
	orNumber := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(255), nil)
	orAmount := big.NewInt(0).Or(amount, orNumber)
	return orAmount
}

func (self *TomoChain) EncodeRateData(source, dest string, quantity *big.Int) (string, error) {
	srcAddr := common.HexToAddress(source)
	destAddr := common.HexToAddress(dest)

	encodedData, err := self.networkAbi.Pack("getExpectedRate", srcAddr, destAddr, getOrAmount(quantity))
	if err != nil {
		log.Print(err)
		return "", err
	}

	return common.Bytes2Hex(encodedData), nil
}

func (self *TomoChain) EncodeRateDataWrapper(source, dest []string, quantity []*big.Int) (string, error) {
	sourceList := make([]common.Address, 0)
	for _, sourceItem := range source {
		sourceList = append(sourceList, common.HexToAddress(sourceItem))
	}
	destList := make([]common.Address, 0)
	for _, destItem := range dest {
		destList = append(destList, common.HexToAddress(destItem))
	}

	quantityList := make([]*big.Int, 0)
	for _, quanItem := range quantity {
		quantityList = append(quantityList, getOrAmount(quanItem))
	}

	encodedData, err := self.wrapperAbi.Pack("getExpectedRates", common.HexToAddress(self.network), sourceList, destList, quantityList)
	if err != nil {
		log.Print(err)
		return "", err
	}

	return common.Bytes2Hex(encodedData), nil
}

func (self *TomoChain) EncodeKyberEnable() (string, error) {
	encodedData, err := self.networkAbi.Pack("enabled")
	if err != nil {
		log.Print(err)
		return "", err
	}
	return common.Bytes2Hex(encodedData), nil
}

func (self *TomoChain) ExtractEnabled(result string) (bool, error) {
	enabledByte, err := hexutil.Decode(result)
	if err != nil {
		log.Print(err)
		return false, err
	}
	var enabled bool
	err = self.networkAbi.Unpack(&enabled, "enabled", enabledByte)
	if err != nil {
		log.Print(err)
		return false, err
	}
	return enabled, nil
}

func (self *TomoChain) EncodeMaxGasPrice() (string, error) {
	encodedData, err := self.networkAbi.Pack("maxGasPrice")
	if err != nil {
		log.Print(err)
		return "", err
	}
	return common.Bytes2Hex(encodedData), nil
}

func (self *TomoChain) ExtractMaxGasPrice(result string) (string, error) {
	gasByte, err := hexutil.Decode(result)
	if err != nil {
		log.Print(err)
		return "", err
	}
	var gasPrice *big.Int
	err = self.networkAbi.Unpack(&gasPrice, "maxGasPrice", gasByte)
	if err != nil {
		log.Print(err)
		return "", err
	}
	return gasPrice.String(), nil
}

func (self *TomoChain) ExtractRateData(result string, sourceSymbol, destSymbol string) (tomochain.Rate, error) {
	var rate tomochain.Rate
	rateByte, err := hexutil.Decode(result)
	if err != nil {
		log.Print(err)
		return rate, err
	}
	var rateNetwork RateNetwork
	err = self.networkAbi.Unpack(&rateNetwork, "getExpectedRate", rateByte)
	if err != nil {
		log.Print(err)
		return rate, err
	}

	return tomochain.Rate{
		Source:  sourceSymbol,
		Dest:    destSymbol,
		Rate:    rateNetwork.ExpectedRate.String(),
		Minrate: rateNetwork.SlippageRate.String(),
	}, nil
}

func (self *TomoChain) ExtractRateDataWrapper(result string, sourceArr, destAddr []string) ([]tomochain.Rate, error) {
	rateByte, err := hexutil.Decode(result)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	var rateWapper RateWrapper
	err = self.wrapperAbi.Unpack(&rateWapper, "getExpectedRates", rateByte)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var lenArr = len(sourceArr)
	if (len(rateWapper.ExpectedRate) != lenArr) || (len(rateWapper.SlippageRate) != lenArr) {
		errorLength := errors.New("Length of expected for slippage rate is not enough")
		log.Print(errorLength)
		return nil, errorLength
	}

	rateReturn := make([]tomochain.Rate, 0)
	for i := 0; i < lenArr; i++ {
		source := sourceArr[i]
		dest := destAddr[i]
		rate := rateWapper.ExpectedRate[i]
		minRate := rateWapper.SlippageRate[i]
		rateReturn = append(rateReturn, tomochain.Rate{
			source, dest, rate.String(), minRate.String(),
		})
	}
	return rateReturn, nil
}

func (self *TomoChain) ReadEventsWithBlockNumber(eventRaw *[]tomochain.EventRaw, latestBlock string) (*[]tomochain.EventHistory, error) {
	//get latestBlock to calculate timestamp
	events, err := self.ReadEvents(eventRaw, "node", latestBlock)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return events, nil
}

func (self *TomoChain) ReadEventsWithTimeStamp(eventRaw *[]tomochain.EventRaw) (*[]tomochain.EventHistory, error) {
	//get latestBlock to calculate timestamp
	events, err := self.ReadEvents(eventRaw, "tomoscan", "0")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return events, nil
}

type LogData struct {
	Source           common.Address `json:"source"`
	Dest             common.Address `json:"dest"`
	ActualSrcAmount  *big.Int       `json:"actualSrcAmount"`
	ActualDestAmount *big.Int       `json:"actualDestAmount"`
}

func (self *TomoChain) ReadEvents(listEventAddr *[]tomochain.EventRaw, typeFetch string, latestBlock string) (*[]tomochain.EventHistory, error) {
	listEvent := *listEventAddr
	endIndex := len(listEvent) - 1
	// var beginIndex = 0
	// if endIndex > 4 {
	// 	beginIndex = endIndex - 4
	// }

	index := 0
	events := make([]tomochain.EventHistory, 0)
	for i := endIndex; i >= 0; i-- {
		if index >= 5 {
			break
		}
		//filter amount
		isSmallAmount, err := self.IsSmallAmount(listEvent[i])
		if err != nil {
			log.Print(err)
			continue
		}
		if isSmallAmount {
			continue
		}

		txHash := listEvent[i].Txhash

		blockNumber, err := hexutil.DecodeBig(listEvent[i].BlockNumber)
		if err != nil {
			log.Print(err)
			continue
		}

		var timestamp string
		if typeFetch == "tomoscan" {
			timestampHex, err := hexutil.DecodeBig(listEvent[i].Timestamp)
			if err != nil {
				log.Print(err)
				continue
			}
			timestamp = timestampHex.String()
			//fmt.Println(timestamp)
		} else {
			timestamp, err = self.Gettimestamp(blockNumber.String(), latestBlock, self.averageBlockTime)
			if err != nil {
				log.Print(err)
				continue
			}
		}

		var logData LogData
		data, err := hexutil.Decode(listEvent[i].Data)
		if err != nil {
			log.Print(err)
			continue
		}
		//fmt.Print(listEvent[i].Data)
		err = self.networkAbi.Unpack(&logData, "ExecuteTrade", data)
		if err != nil {
			log.Print(err)
			continue
		}

		actualDestAmount := logData.ActualDestAmount.String()
		actualSrcAmount := logData.ActualSrcAmount.String()
		dest := logData.Dest.String()
		source := logData.Source.String()

		events = append(events, tomochain.EventHistory{
			actualDestAmount, actualSrcAmount, dest, source, blockNumber.String(), txHash, timestamp,
		})
		index++
	}
	return &events, nil
}

func (self *TomoChain) IsSmallAmount(eventRaw tomochain.EventRaw) (bool, error) {
	data, err := hexutil.Decode(eventRaw.Data)
	if err != nil {
		log.Print(err)
		return true, err
	}
	var logData LogData
	err = self.networkAbi.Unpack(&logData, "ExecuteTrade", data)
	if err != nil {
		log.Print(err)
		return true, err
	}

	source := logData.Source
	var amount *big.Int
	if strings.ToLower(source.String()) == "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" {
		amount = logData.ActualSrcAmount
	} else {
		amount = logData.ActualDestAmount
	}

	//fmt.Print(amount.String())
	//check amount is greater than episol
	// amountBig, ok := new(big.Int).SetString(amount)
	// if !ok {
	// 	errorAmount := errors.New("Cannot read amount as number")
	// 	log.Print(errorAmount)
	// 	return false, errorAmount
	// }
	var episol, weight = big.NewInt(10), big.NewInt(15)
	episol.Exp(episol, weight, nil)
	//fmt.Print(episol.String())

	//fmt.Print(amount.Cmp(episol))
	if amount.Cmp(episol) == -1 {
		return true, nil
	}
	return false, nil
}

func (self *TomoChain) Gettimestamp(block string, latestBlock string, averageBlockTime int64) (string, error) {
	fromBlock, err := strconv.ParseInt(block, 10, 64)
	if err != nil {
		log.Print(err)
		return "", err
	}
	toBlock, err := strconv.ParseInt(latestBlock, 10, 64)
	if err != nil {
		log.Print(err)
		return "", err
	}
	timeNow := time.Now().Unix()
	timeStamp := timeNow - averageBlockTime*(toBlock-fromBlock)/1000

	timeStampBig := big.NewInt(timeStamp)
	return timeStampBig.String(), nil
}
