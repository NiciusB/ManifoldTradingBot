import Dotenv
import Alamofire
import Foundation

final class MainApp: Sendable {
    private let manifoldApi: ManifoldApi
    let alpacaApi: AlpacaApi
    private let sharesOutstandingApi: SharesOutstandingApi
    private let marketsDb: [Market]
    
    private struct Market {
        var id: String
        var realStockSymbol: String
    }
    
    init() async throws {
        let dotenv = try Dotenv()
        let manifoldApi = ManifoldApi(
            apiKey: dotenv.get("MANIFOLD_API_KEY")!
        )
        let alpacaApi = try await AlpacaApi(
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
            Market(id: "1IBrgJ6IlwBIaJ7xdQ5c", realStockSymbol: "MSFT")
        ]

        let stocksToTrack = marketsDb.map { market in market.realStockSymbol}
        
        print("Connecting to Alpaca API to track: \(stocksToTrack)...")
        
        try await alpacaApi.subscribe(
            trades: stocksToTrack
        )
        
        print("Connected to Alpaca API")
    }

    func startAppLogicLoopTimer() {
        _ = Timer.scheduledTimer(withTimeInterval: 30.0, repeats: true, block: { _ in
            Task {
                do {
                    try await self.runAppLogicLoop()
                } catch {
                    printErr(error)
                    exit(1)
                }
            }
        })
    }
    
    private func runAppLogicLoop() async throws {
        var realTradeValues: [(Market, Float)] = []
        for market in self.marketsDb {
                let lastStockTradeValue = try await alpacaApi.getStockLastTradeValue(market.realStockSymbol)
                if lastStockTradeValue != nil {
                    realTradeValues.append((market, lastStockTradeValue!))
                }
        }
        
        if realTradeValues.isEmpty {
            print("We have no stock data for any market. That's probably because the markets closed right now, so we don't get realtime data. And we don't poll for historic data for now, so just wait until they open")
            return
        }

        print("Getting data from manifold API for \(realTradeValues.count) markets...")
        let manifoldMarkets = try await getManifoldMarkets(realTradeValues.map { (market, _) in market.id })

        print("Getting data from outstanding shares website...")
        let outstandingShares = await getOutstandingShares(realTradeValues.map { (market, _) in market.realStockSymbol })
        
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
            
            if betAmount >= 1 {
                print("\(market.realStockSymbol) (\(manifoldMarket.url)): Found current value \(currentMarketValue) VS expected value \(targetMarketValue). Betting $\(betAmount) on \(outcome)")
                try await manifoldApi.placeBet(amount: betAmount, contractId: manifoldMarket.id, outcome: outcome)
            }
        }

        print("Betting round done!")
    }
    
    private func getManifoldMarkets(_ manifoldMarketIds: [String]) async throws -> [String: GetMarket.ResDec] {
        var result: [String: GetMarket.ResDec] = [:]
        for marketId in manifoldMarketIds {
            result[marketId] = try await self.manifoldApi.getMarket(marketId)
        }
        return result
    }
    
    private func getOutstandingShares(_ realStockSymbols: [String]) async -> [String: Int?] {
        var result: [String: Int?] = [:]
        for symbol in realStockSymbols {
            result[symbol] = await self.sharesOutstandingApi.getSymbolOutstandingShares(symbol)
        }
        return result
    }
}
