import Dotenv
import Alamofire
import Foundation

let mainApp = try MainApp()
try await mainApp.start()

class MainApp {
    private var manifoldApi: ManifoldApi
    private var alpacaApi: AlpacaApi
    private var sharesOutstandingApi: SharesOutstandingApi
    private var marketsDb: [Market]
    
    private struct Market {
        var id: String
        var realStockSymbol: String
    }
    
    init() throws {
        let dotenv = try Dotenv()
        let manifoldApi = ManifoldApi(
            apiKey: dotenv.get("MANIFOLD_API_KEY")!
        )
        let alpacaApi = AlpacaApi(
            apiEndpoint: dotenv.get("ALPACA_API_ENDPOINT")!,
            apiKey: dotenv.get("ALPACA_API_KEY")!,
            apiSecret: dotenv.get("ALPACA_API_SECRET")!
        )
        let sharesOutstandingApi = SharesOutstandingApi()
        
        self.manifoldApi = manifoldApi
        self.alpacaApi = alpacaApi
        self.sharesOutstandingApi = sharesOutstandingApi
        self.marketsDb = [
            Market(id: "aZn4kn9dIv5wjQSbVzdk", realStockSymbol: "AAPL"),
            Market(id: "qy4Pujoc7k2G03cb7Vnh", realStockSymbol: "AMZN"),
            Market(id: "RnzTxpnUSsbfPG8Ec6BO", realStockSymbol: "GOOG"),
            Market(id: "78LK7lYi6fgGHMWvCG8j", realStockSymbol: "MSFT")
        ]
    }

    func start() async throws {
        let stocksToTrack = marketsDb.map { market in market.realStockSymbol}
        
        try await alpacaApi.connect()
        try await alpacaApi.subscribe(
            trades: stocksToTrack
        )
        
        print("Tracking: \(stocksToTrack)")

        Task {
            while true {
                try await self.runAppLogicLoop()
                try await Task.sleep(nanoseconds: 1000 * 1000 * 1000 * 30)
            }
        }
        
        try await alpacaApi.waitUntilConnectionEnd()
    }
    
    private func runAppLogicLoop() async throws {
        let realTradeValues = self.marketsDb.compactMap { market -> (Market, Float)? in
            let lastStockTradeValue = alpacaApi.getStockLastTradeValue(market.realStockSymbol)
            if lastStockTradeValue == nil {
                return nil
            } else {
                return (market, lastStockTradeValue!)
            }
        }
        
        if realTradeValues.isEmpty {
            print("We have no stock data for any market. That's probably because the markets closed right now, so we don't get realtime data. And we don't poll for historic data for now, so just wait until they open")
            return
        }
        
        let manifoldMarkets = try await getManifoldMarkets(realTradeValues.map { (market, _) in market.id })
        let outstandingShares = try await getOutstandingShares(realTradeValues.map { (market, _) in market.realStockSymbol })
        
        for (market, lastStockTradeValue) in realTradeValues {
            let manifoldMarket = manifoldMarkets[market.id]!
            let stockOutstandingShares = outstandingShares[market.realStockSymbol]!!
            
            let targetMarketValue: Float = lastStockTradeValue * Float(stockOutstandingShares) / 1000 / 1000 / 1000 // Hardcoded division by 1B, as the markets I'm betting in use that
            let currentMarketValue = CpmmMarketUtils.calculatePseudoNumericMarketplaceValue(manifoldMarket)

            if targetMarketValue / currentMarketValue >  1.2 || targetMarketValue / currentMarketValue <  0.8 {
                print("\(market.realStockSymbol) (\(manifoldMarket.url)): Value difference too high, something must be wrong. (Found current value \(currentMarketValue) VS expected value \(targetMarketValue))")
                continue
            }

            let (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(manifoldMarket, targetValue: targetMarketValue)
            
            print("\(market.realStockSymbol) (\(manifoldMarket.url)): Betting $\(betAmount) on \(outcome) (Found current value \(currentMarketValue) VS expected value \(targetMarketValue))")
            if betAmount >= 1 {
                _  = try await manifoldApi.placeBet(amount: betAmount, contractId: manifoldMarket.id, outcome: outcome)
            }
        }
    }
    
    private func getManifoldMarkets(_ manifoldMarketIds: [String]) async throws -> [String: GetMarket.ResDec] {
        return try await withThrowingTaskGroup(of: (String, GetMarket.ResDec).self, body: { group in
            for marketId in manifoldMarketIds {
                group.addTask {
                    return try await (marketId, self.manifoldApi.getMarket(marketId))
                }
            }
            
            var result: [String: GetMarket.ResDec] = [:]
            for try await (marketId, manifoldMarket) in group {
                result[marketId] = manifoldMarket
            }
            return result
        })
    }
    
    private func getOutstandingShares(_ realStockSymbols: [String]) async throws -> [String: Int?] {
        return try await withThrowingTaskGroup(of: (String, Int?).self, body: { group in
            for symbol in realStockSymbols {
                group.addTask {
                    return await (symbol, self.sharesOutstandingApi.getSymbolOutstandingShares(symbol))
                }
            }
            
            var result: [String: Int?] = [:]
            for try await (marketId, outstandingShares) in group {
                result[marketId] = outstandingShares
            }
            return result
        })
    }
}
