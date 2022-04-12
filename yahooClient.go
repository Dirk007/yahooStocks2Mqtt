package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	resty "github.com/go-resty/resty/v2"
)

// TODO: isn't there any equivalent to serde:rename_all(camelCase)?
type YahooStockInfo struct {
	QuoteType          string  `json:"quoteType"`
	Currency           string  `json:"currency"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
	Symbol             string  `json:"symbol"`
}

type YahooQuoteResponse struct {
	Result []YahooStockInfo `json:"result"`
}

type YahooResponse struct {
	QuoteResponse YahooQuoteResponse `json:"quoteResponse"`
}

func requestSymbols(symbols []string) ([]YahooStockInfo, error) {
	symbolsRequest := strings.Join(symbols, ",")

	client := resty.New()
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"symbols": symbolsRequest,
		}).
		SetHeader("Accept", "application/json").
		Get("https://query2.finance.yahoo.com/v7/finance/quote")

	if err != nil {
		return nil, fmt.Errorf("Unable to request quotes: %v", err)
	}

	var yahooResponse YahooResponse
	uerr := json.Unmarshal(resp.Body(), &yahooResponse)
	if uerr != nil {
		return nil, fmt.Errorf("Unable to unmarshal quotes response: %v", uerr)
	}

	return yahooResponse.QuoteResponse.Result, nil
}

// Sleeps the given `duration` and signals the completion via the sleepDone channel
func sleeper(duration time.Duration, sleepDone chan bool) {
	time.Sleep(duration)
	sleepDone <- true
}

func requestLoop(symbols []string, requestPeriod time.Duration, quotesChannel chan YahooStockInfo, kill chan bool) {
	sleepDone := make(chan bool)
	for {
		quotes, err := requestSymbols(symbols)
		if err != nil {
			log.Fatalf("Unable to request symbols: %v", err)
		} else {
			log.Printf("Got quotes: %v\n", quotes)
			for _, quote := range quotes {
				quotesChannel <- quote
			}
		}

		go sleeper(requestPeriod, sleepDone)
		select {
		case _ = <-sleepDone:
			continue
		case _ = <-kill:
			log.Println("Exiting requestLoop for killswitch")
			return
		}
	}
}
