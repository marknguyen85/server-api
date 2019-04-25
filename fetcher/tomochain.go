package fetcher

import (
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
	averageBlockTime int64
}

//NewTomoChain contruct func
func NewTomoChain(network string, networkAbiStr string, tradeTopic string, averageBlockTime int64) (*TomoChain, error) {

	networkAbi, err := abi.JSON(strings.NewReader(networkAbiStr))
	if err != nil {
		log.Print(err)
		return nil, err
	}

	tomochain := &TomoChain{
		network, networkAbi, tradeTopic, averageBlockTime,
	}

	return tomochain, nil
}

//EncodeRateData func
func (tomoChain *TomoChain) EncodeRateData(source, dest string, quantity *big.Int) (string, error) {
	srcAddr := common.HexToAddress(source)
	destAddr := common.HexToAddress(dest)

	encodedData, err := tomoChain.networkAbi.Pack("getExpectedRate", srcAddr, destAddr, quantity)
	if err != nil {
		log.Print(err)
		return "", err
	}

	return common.Bytes2Hex(encodedData), nil
}

//EncodeChainTeXEnable func
func (tomoChain *TomoChain) EncodeChainTeXEnable() (string, error) {
	encodedData, err := tomoChain.networkAbi.Pack("enabled")
	if err != nil {
		log.Print(err)
		return "", err
	}
	return common.Bytes2Hex(encodedData), nil
}

func (tomoChain *TomoChain) ExtractEnabled(result string) (bool, error) {
	enabledByte, err := hexutil.Decode(result)
	if err != nil {
		log.Print(err)
		return false, err
	}
	var enabled bool
	err = tomoChain.networkAbi.Unpack(&enabled, "enabled", enabledByte)
	if err != nil {
		log.Print(err)
		return false, err
	}
	return enabled, nil
}

func (tomoChain *TomoChain) EncodeMaxGasPrice() (string, error) {
	encodedData, err := tomoChain.networkAbi.Pack("maxGasPrice")
	if err != nil {
		log.Print(err)
		return "", err
	}
	return common.Bytes2Hex(encodedData), nil
}

func (tomoChain *TomoChain) ExtractMaxGasPrice(result string) (string, error) {
	gasByte, err := hexutil.Decode(result)
	if err != nil {
		log.Print(err)
		return "", err
	}
	var gasPrice *big.Int
	err = tomoChain.networkAbi.Unpack(&gasPrice, "maxGasPrice", gasByte)
	if err != nil {
		log.Print(err)
		return "", err
	}
	return gasPrice.String(), nil
}

//ExtractRateData func
func (tomoChain *TomoChain) ExtractRateData(result string, sourceSymbol, destSymbol string) (tomochain.Rate, error) {
	var rate tomochain.Rate
	rateByte, err := hexutil.Decode(result)

	if err != nil {
		log.Print("**********************Decode error******************", err)
		return rate, err
	}
	var rateNetwork RateNetwork
	err = tomoChain.networkAbi.Unpack(&rateNetwork, "getExpectedRate", rateByte)
	if err != nil {
		log.Print("**********************Unpack getExpectedRate******************", err)
		return rate, err
	}

	return tomochain.Rate{
		Source:  sourceSymbol,
		Dest:    destSymbol,
		Rate:    rateNetwork.ExpectedRate.String(),
		Minrate: rateNetwork.SlippageRate.String(),
	}, nil
}

func (tomoChain *TomoChain) ReadEventsWithBlockNumber(eventRaw *[]tomochain.EventRaw, latestBlock string) (*[]tomochain.EventHistory, error) {
	//get latestBlock to calculate timestamp
	events, err := tomoChain.ReadEvents(eventRaw, "node", latestBlock)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return events, nil
}

func (tomoChain *TomoChain) ReadEventsWithTimeStamp(eventRaw *[]tomochain.EventRaw) (*[]tomochain.EventHistory, error) {
	//get latestBlock to calculate timestamp
	events, err := tomoChain.ReadEvents(eventRaw, "tomoscan", "0")
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

func (tomoChain *TomoChain) ReadEvents(listEventAddr *[]tomochain.EventRaw, typeFetch string, latestBlock string) (*[]tomochain.EventHistory, error) {
	listEvent := *listEventAddr
	endIndex := len(listEvent) - 1

	index := 0
	events := make([]tomochain.EventHistory, 0)
	for i := endIndex; i >= 0; i-- {
		if index >= 5 {
			break
		}
		//filter amount
		isSmallAmount, err := tomoChain.IsSmallAmount(listEvent[i])
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
			timestamp, err = tomoChain.Gettimestamp(blockNumber.String(), latestBlock, tomoChain.averageBlockTime)
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
		err = tomoChain.networkAbi.Unpack(&logData, "ExecuteTrade", data)
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

func (tomoChain *TomoChain) IsSmallAmount(eventRaw tomochain.EventRaw) (bool, error) {
	data, err := hexutil.Decode(eventRaw.Data)
	if err != nil {
		log.Print(err)
		return true, err
	}
	var logData LogData
	err = tomoChain.networkAbi.Unpack(&logData, "ExecuteTrade", data)
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

func (tomoChain *TomoChain) Gettimestamp(block string, latestBlock string, averageBlockTime int64) (string, error) {
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
