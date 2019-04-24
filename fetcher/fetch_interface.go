package fetcher

import (
	bFetcher "github.com/marknguyen85/server-api/fetcher/blockchain-fetcher"
)

type RateUSD struct {
	Symbol   string `json:"symbol"`
	PriceUsd string `json:"price_usd"`
}

type FetcherInterface interface {
	TomoCall(string, string) (string, error)
	GetLatestBlock() (string, error)
	GetTypeName() string

	GetRate(string, string) (string, error)
}

//var transactionPersistent = models.NewTransactionPersister()

func NewFetcherIns(typeName string, endpoint string, apiKey string) (FetcherInterface, error) {
	var fetcher FetcherInterface
	var err error
	switch typeName {
	case "tomoscan":
		fetcher, err = bFetcher.NewTomoScan(typeName, endpoint, apiKey)
		break
	case "node":
		fetcher, err = bFetcher.NewBlockchainFetcher(typeName, endpoint, apiKey)
		break
	}
	return fetcher, err
}
