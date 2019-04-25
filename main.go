package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/marknguyen85/server-api/common"
	"github.com/marknguyen85/server-api/fetcher"
	"github.com/marknguyen85/server-api/http"
	persister "github.com/marknguyen85/server-api/persister"
	"github.com/marknguyen85/server-api/tomochain"
)

type fetcherFunc func(persister persister.Persister, boltIns persister.BoltInterface, fetcher *fetcher.Fetcher)

func enableLogToFile() (*os.File, error) {
	const logFileName = "log/error.log"
	f, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	//clear error log file
	if err = f.Truncate(0); err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)
	return f, nil
}

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	//set log for server
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if os.Getenv("LOG_TO_STDOUT") != "true" {
		f, err := enableLogToFile()
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}

	chainTexENV := os.Getenv("CHAINTEX_ENV")
	persisterIns, _ := persister.NewPersister("ram")
	boltIns, err := persister.NewBoltStorage()
	// boltIns, err := persister.NewInfluxStorage()
	if err != nil {
		log.Println("cannot init db: ", err.Error())
	}
	fertcherIns, err := fetcher.NewFetcher(chainTexENV)
	if err != nil {
		log.Fatal(err)
	}

	err = fertcherIns.TryUpdateListToken()
	if err != nil {
		log.Println(err)
	}

	tickerUpdateToken := time.NewTicker(300 * time.Second)
	go func() {
		for {
			<-tickerUpdateToken.C
			fertcherIns.TryUpdateListToken()
		}
	}()
	var (
		initRate   []tomochain.Rate
		tomoSymbol = common.TOMOSymbol
	)

	for symbol := range fertcherIns.GetListToken() {
		if symbol == tomoSymbol {
			tomoRate := tomochain.Rate{
				Source:  tomoSymbol,
				Dest:    tomoSymbol,
				Rate:    "0",
				Minrate: "0",
			}
			initRate = append(initRate, tomoRate, tomoRate)
		} else {
			buyRate := tomochain.Rate{
				Source:  tomoSymbol,
				Dest:    symbol,
				Rate:    "0",
				Minrate: "0",
			}
			sellRate := tomochain.Rate{
				Source:  symbol,
				Dest:    tomoSymbol,
				Rate:    "0",
				Minrate: "0",
			}
			initRate = append(initRate, buyRate, sellRate)
		}
	}
	persisterIns.SaveRate(initRate, 0)
	tokenNum := fertcherIns.GetNumTokens()
	bonusTimeWait := 900
	if tokenNum > 200 {
		bonusTimeWait = 60
	}
	intervalFetchGeneralInfoTokens := time.Duration((tokenNum * 7) + bonusTimeWait)

	runFetchData(persisterIns, boltIns, fetchRateUSD, fertcherIns, 300) //5 minutes

	runFetchData(persisterIns, boltIns, fetchGeneralInfoTokens, fertcherIns, intervalFetchGeneralInfoTokens)

	runFetchData(persisterIns, boltIns, fetchRate7dData, fertcherIns, 300) //5 minutes

	runFetchData(persisterIns, boltIns, fetchRate, fertcherIns, 15) //15 seconds
	runFetchData(persisterIns, boltIns, fetchRateWithFallback, fertcherIns, 300)
	//run server
	server := http.NewHTTPServer(":3001", persisterIns, fertcherIns)
	server.Run(chainTexENV)

}

func runFetchData(persister persister.Persister, boltIns persister.BoltInterface, fn fetcherFunc, fertcherIns *fetcher.Fetcher, interval time.Duration) {
	ticker := time.NewTicker(interval * time.Second)
	go func() {
		for {
			fn(persister, boltIns, fertcherIns)
			<-ticker.C
		}
	}()
}

func fetchRateUSD(persister persister.Persister, boltIns persister.BoltInterface, fetcher *fetcher.Fetcher) {
	rateUSD, err := fetcher.GetRateUsdTomo()
	if err != nil {
		log.Print(err)
		persister.SetNewRateUSD(false)
		return
	}

	if rateUSD == "" {
		persister.SetNewRateUSD(false)
		return
	}

	err = persister.SaveRateUSD(rateUSD)
	if err != nil {
		log.Print(err)
		persister.SetNewRateUSD(false)
		return
	}
}

func makeMapRate(rates []tomochain.Rate) map[string]tomochain.Rate {
	mapRate := make(map[string]tomochain.Rate)
	for _, r := range rates {
		mapRate[fmt.Sprintf("%s_%s", r.Source, r.Dest)] = r
	}
	return mapRate
}

func fetchGeneralInfoTokens(persister persister.Persister, boltIns persister.BoltInterface, fetcher *fetcher.Fetcher) {
	generalInfo := fetcher.GetGeneralInfoTokens()
	persister.SaveGeneralInfoTokens(generalInfo)
	err := boltIns.StoreGeneralInfo(generalInfo)
	if err != nil {
		log.Println(err.Error())
	}
}

func fetchRate7dData(persister persister.Persister, boltIns persister.BoltInterface, fetcher *fetcher.Fetcher) {
	data, err := fetcher.FetchRate7dData()
	if err != nil {
		log.Print(err)
		if !persister.IsFailedToFetchTracker() {
			return
		}
		persister.SetIsNewTrackerData(false)
	} else {
		persister.SetIsNewTrackerData(true)
	}
	mapToken := fetcher.GetListToken()
	currentGeneral, err := boltIns.GetGeneralInfo(mapToken)
	if err != nil {
		log.Println(err.Error())
		currentGeneral = make(map[string]*tomochain.TokenGeneralInfo)
	}
	persister.SaveMarketData(data, currentGeneral, mapToken)
	// persister.SetIsNewMarketInfo(true)
}

func fetchRate(persister persister.Persister, boltIns persister.BoltInterface, fetcher *fetcher.Fetcher) {
	timeNow := time.Now().UTC().Unix()
	var result []tomochain.Rate
	currentRate := persister.GetRate()
	tokenPriority := fetcher.GetListTokenPriority()
	rates, err := fetcher.GetRate(currentRate, persister.GetIsNewRate(), tokenPriority, false)
	if err != nil {
		log.Print(err)
		persister.SetIsNewRate(false)
		return
	}
	mapRate := makeMapRate(rates)
	for _, cr := range currentRate {
		keyRate := fmt.Sprintf("%s_%s", cr.Source, cr.Dest)
		if r, ok := mapRate[keyRate]; ok {
			result = append(result, r)
			delete(mapRate, keyRate)
		} else {
			result = append(result, cr)
		}
	}
	// add new token to current rate
	if len(mapRate) > 0 {
		for _, nr := range mapRate {
			result = append(result, nr)
		}
	}
	persister.SaveRate(result, timeNow)
	persister.SetIsNewRate(true)
}

func fetchRateWithFallback(persister persister.Persister, boltIns persister.BoltInterface, fetcher *fetcher.Fetcher) {
	var result []tomochain.Rate
	currentRate := persister.GetRate()
	listToken := fetcher.GetListToken()
	newList := make(map[string]tomochain.Token)
	for _, t := range listToken {
		if !t.Priority {
			newList[t.Symbol] = t
		}
	}
	rates, err := fetcher.GetRate(currentRate, persister.GetIsNewRate(), newList, true)
	if err != nil {
		log.Print(err)
		persister.SetIsNewRate(false)
		return
	}
	mapRate := makeMapRate(rates)
	for _, cr := range currentRate {
		keyRate := fmt.Sprintf("%s_%s", cr.Source, cr.Dest)
		if r, ok := mapRate[keyRate]; ok {
			result = append(result, r)
			if keyRate != "TOMO_TOMO" {
				delete(mapRate, keyRate)
			}
		} else {
			result = append(result, cr)
		}
	}
	// add new token to current rate
	if len(mapRate) > 1 {
		for _, nr := range mapRate {
			result = append(result, nr)
		}
	}
	persister.SaveRate(result, 0)
}
