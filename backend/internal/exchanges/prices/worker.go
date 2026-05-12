package prices

import (
	"context"
	"cryptofund/internal/exchanges/clients"
	"log"
	"net/http"
	"time"
)

func (p *PriceCache) Worker(ctx context.Context, httpClient *http.Client) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	mexcClient := clients.NewMexcClient(httpClient)
	gateClient := clients.NewGateClient(httpClient)
	binanceClient := clients.NewBinanceClient(httpClient)

	p.updateMexcPrices(ctx, mexcClient)
	p.updateGatePrices(ctx, gateClient)
	p.updateBinancePrices(ctx, binanceClient)

	for {
		select {
		case <-ctx.Done():
			log.Println("Redis price worker stopped")
			return

		case <-ticker.C:
			p.updateMexcPrices(ctx, mexcClient)
			p.updateGatePrices(ctx, gateClient)
			p.updateBinancePrices(ctx, binanceClient)
		}
	}
}

func (p *PriceCache) updateMexcPrices(ctx context.Context, mexcClient *clients.MexcClient) {
	pairs := PopularPairs()

	for _, pair := range pairs {
		price, err := mexcClient.GetAssetValueUSDT(ctx, pair)
		if err != nil {
			log.Printf("error getting mexc price for %s: %v", pair, err)
			continue
		}

		key := PriceCacheKey("mexc", pair)

		err = p.SetPrice(ctx, key, price)
		if err != nil {
			log.Printf("error saving mexc price for %s to redis: %v", pair, err)
			continue
		}
	}
}

func (p *PriceCache) updateGatePrices(ctx context.Context, gateClient *clients.GateClient) {
	pairs := PopularPairs()

	for _, pair := range pairs {
		price, err := gateClient.GetAssetPriceUSDT(ctx, pair)
		if err != nil {
			log.Printf("error getting gate price for %s: %v", pair, err)
			continue
		}

		key := PriceCacheKey("gate", pair)

		err = p.SetPrice(ctx, key, price)
		if err != nil {
			log.Printf("error saving gate price for %s to redis: %v", pair, err)
			continue
		}
	}
}

func (p *PriceCache) updateBinancePrices(ctx context.Context, binanceClient *clients.BinanceClient) {
	pairs := PopularPairs()

	for _, pair := range pairs {
		price, err := binanceClient.GetAssetPriceUSDT(ctx, pair)
		if err != nil {
			log.Printf("error getting binance price for %s: %v", pair, err)
			continue
		}

		key := PriceCacheKey("binance", pair)

		err = p.SetPrice(ctx, key, price)
		if err != nil {
			log.Printf("error saving binance price for %s to redis: %v", pair, err)
			continue
		}
	}
}
