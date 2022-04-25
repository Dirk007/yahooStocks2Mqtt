package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

// TODO: isn't there any equivalent to serde:rename_all(camelCase)?
type YahooStockInfo struct {
	QuoteType          string  `json:"quoteType"`
	Currency           string  `json:"currency"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
	Symbol             string  `json:"symbol"`
}

func (stock YahooStockInfo) Serialize() (*string, error) {
	result, err := json.Marshal(stock)
	if err != nil {
		return nil, err
	}

	strResult := string(result)
	return &strResult, nil
}

type YahooQuoteResponse struct {
	Result []YahooStockInfo `json:"result"`
}

type YahooResponse struct {
	QuoteResponse YahooQuoteResponse `json:"quoteResponse"`
}

func requestSymbols(symbols []string) ([]YahooStockInfo, error) {
	symbolsList := strings.Join(symbols, ",")

	client := resty.New()
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"symbols": symbolsList,
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
func sleeper(ctx context.Context, duration time.Duration, sleepDone chan bool) {
	timer := time.NewTimer(duration)
	select {
	case _ = <-ctx.Done():
		log.Debug("Timer done")
		if !timer.Stop() {
			log.Info("Timer could not be stopped. Draining.")
			<-timer.C
		}
	case _ = <-timer.C:
	}
	log.Debug("Timer signalling finish")
	sleepDone <- true
}

func requestLoop(symbols []string, requestPeriod time.Duration, quotesChannel chan YahooStockInfo, kill chan bool) {
	sleepDone := make(chan bool)
	for {
		quotes, err := requestSymbols(symbols)
		if err != nil {
			log.Fatalf("Unable to request symbols: %v", err)
		} else {
			log.Debugf("Got quotes: %v", quotes)
			for _, quote := range quotes {
				quotesChannel <- quote
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		go sleeper(ctx, requestPeriod, sleepDone)
		select {
		case _ = <-sleepDone:
			continue
		case _ = <-kill:
			log.Info("Exiting requestLoop for killswitch, canceling the sleep for next request")
			cancel()
			return
		}
	}
}
