import Foundation
import Alamofire

// I don't think this class is actually thread-safe, but it shouldn't cause problems because we mostly use it sequentially
final class AlpacaApi: @unchecked Sendable {
    private let apiEndpoint: String
    private let apiKey: String
    private let apiSecret: String
    private let alpacaApiWebsocket: AlpacaApiWebsocket
    
    init(apiEndpoint: String, apiKey: String, apiSecret: String) async throws {
        self.apiEndpoint = apiEndpoint
        self.apiSecret = apiSecret
        self.apiKey = apiKey
        
        self.alpacaApiWebsocket = try await AlpacaApiWebsocket(apiKey: apiKey, apiSecret: apiSecret)
    }
    
    func getHistoricLatestTrade(
        _ symbol: String
    ) async throws -> AlpacaApiLatestTradeResponseRoot {
        let request = AF.request(
            self.apiEndpoint + "/v2/stocks/\(symbol)/trades/latest",
            method: .get,
            headers: HTTPHeaders([
                "APCA-API-KEY-ID": self.apiKey,
                "APCA-API-SECRET-KEY": self.apiSecret,
                "User-Agent": "ManifoldTradingBot/1.0.0 for @NiciusBot",
                "Accept": "application/json"
            ]),
            requestModifier: { $0.timeoutInterval = 10 }
        )
        
        let dataTask = request.serializingData()
        let dataResponse = await dataTask.response
        let resData = try dataResponse.result.get()
        
        return try JSONDecoder().decode(AlpacaApiLatestTradeResponseRoot.self, from: resData)
    }
    
    func subscribe(
        quotes: [String]? = nil,
        trades: [String]? = nil,
        bars: [String]? = nil,
        dailyBars: [String]? = nil,
        statuses: [String]? = nil
    ) async throws {
        return try await self.alpacaApiWebsocket.subscribe(
            quotes: quotes,
            trades: trades,
            bars: bars,
            dailyBars: dailyBars,
            statuses: statuses
        )
    }
    
    func getStockLastTradeValue(_ stock: String) async throws -> Float? {
        let websocketLatestValue = self.alpacaApiWebsocket.getStockLastTradeValue(stock)
        if websocketLatestValue != nil {
            return  websocketLatestValue
        }
        
        // Fallback to Historic API until we have values from Websocket
        let historicApiLatestValue = try await self.getHistoricLatestTrade(stock)
        return historicApiLatestValue.trade.p
    }
}

// swiftlint:disable identifier_name
struct AlpacaApiLatestTradeResponseRoot: Decodable {
    let symbol: String
    let trade: AlpacaApiLatestTradeResponseTrade
}
struct AlpacaApiLatestTradeResponseTrade: Decodable {
    let t: String
    let x: String
    let p: Float
    let s: Float
    let c: [String]
    let i: Float
    let z: String
}
// swiftlint:enable identifier_name
