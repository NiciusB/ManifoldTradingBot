package modulevelocity

import (
	"log"
	"sync"
)

type loadedCachesType struct {
	market         cachedMarket
	myPosition     cachedMarketPosition
	betCreatorUser cachedUser
	marketVelocity bool
}

func loadCachesForBet(
	bet SupabaseBet,
) loadedCachesType {
	var caches loadedCachesType

	var wg sync.WaitGroup

	// Load in advance all needed data, even for obviously not needed markets, since it warms up the cache
	wg.Add(4)

	go func() {
		var cachedMarket, err = marketsCache.Get(bet.ContractID)
		if err != nil {
			log.Fatalln(err)
		}
		caches.market = *cachedMarket
		wg.Done()
	}()
	go func() {
		var myPosition, err = myMarketPositionCache.Get(bet.ContractID)
		if err != nil {
			log.Fatalln(err)
		}
		caches.myPosition = *myPosition
		wg.Done()
	}()
	go func() {
		var cachedUser, err = usersCache.Get(bet.UserID)
		if err != nil {
			log.Fatalln(err)
		}
		caches.betCreatorUser = *cachedUser
		wg.Done()
	}()
	go func() {
		var marketVelocity, err = marketVelocityCache.Get(bet.ContractID)
		if err != nil {
			log.Fatalln(err)
		}
		caches.marketVelocity = *marketVelocity
		wg.Done()
	}()

	wg.Wait()

	return caches
}
