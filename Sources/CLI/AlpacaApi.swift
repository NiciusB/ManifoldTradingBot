import Foundation
import AnyCodable

class AlpacaApi {
    let apiEndpoint: String
    let apiKey: String
    let apiSecret: String
    // TODO: Migrate to https://swiftpackageindex.com/apple/swift-nio/main/documentation/niowebsocket to allow compiling for Linux
    private let webSocket: URLSessionWebSocketTask
    private let loginTask = Task(priority: .background) {
        while true {
            try await Task.sleep(nanoseconds: 100000)
            await Task.yield()
        }
    }
    private var connectionTask: Task<Void, Error>?
    private var lastTradeValues: [String: Float] = [:]
    
    init(apiEndpoint: String, apiKey: String, apiSecret: String) {
        self.apiEndpoint = apiEndpoint
        self.apiSecret = apiSecret
        self.apiKey = apiKey
        
        let wsUrl = "wss://stream.data.alpaca.markets/v2/iex"
        
        let session = URLSession(configuration: URLSessionConfiguration.default)
        
        webSocket = session.webSocketTask(with: URL(string: wsUrl)!)
    }
    
    func connect() async throws {
        webSocket.resume()
        
        self.connectionTask = Task {
            try await self.receive()
        }
        
        try? await loginTask.value
    }
    
    func waitUntilConnectionEnd() async throws {
        _ = try await self.connectionTask?.value
    }
    
    func subscribe(
        quotes: [String]? = nil,
        trades: [String]? = nil,
        bars: [String]? = nil,
        dailyBars: [String]? = nil,
        statuses: [String]? = nil
    ) async throws {
        let subscribeMsg = AlpacaApiClientMessages.Subscribe(
            trades: trades,
            quotes: quotes,
            bars: bars,
            dailyBars: dailyBars,
            statuses: statuses
        )
        let data = try JSONEncoder().encode(subscribeMsg)
        webSocket.send(URLSessionWebSocketTask.Message.data(data)) { error in
            if error != nil {
                print(error!)
            }
        }
    }
    
    private func sendAuthMsg() throws {
        let authMsg = AlpacaApiClientMessages.Auth(key: self.apiKey, secret: self.apiSecret)
        let data = try JSONEncoder().encode(authMsg)
        webSocket.send(URLSessionWebSocketTask.Message.data(data)) { error in
            if error != nil {
                print(error!)
            }
        }
    }
    
    func getStockLastTradeValue(_ stock: String) -> Float? {
        return lastTradeValues[stock]
    }
    
    private func receive() async throws {
        if self.webSocket.state != .running {
            // Not connected, end receive loop
            return
        }
        
        let result = try await self.webSocket.receive()
        
        switch result {
        case .string(let strMessage):
            let data = strMessage.data(using: .utf8)!
            let decoded = try JSONDecoder().decode(
                AlpacaApiServerMessages.AlpacaApiServerMessageDecoder.self,
                from: data
            )
            
            try decoded.messages.forEach { decodedMsg in
                switch decodedMsg {
                case let .success(data):
                    if data.msg == "connected" {
                        try self.sendAuthMsg()
                    } else if data.msg == "authenticated" {
                        loginTask.cancel()
                    }
                case let .error(data):
                    print(data)
                case .subscription:
                    // subscribed!
                    break
                case let .trade(data):
                    lastTradeValues[data.S] = data.p
                case let .quote(data):
                    print(data)
                }
            }
            
        default:
            throw RuntimeError("Received data instead of string type for WebSocket message")
        }
        
        try await self.receive()
    }
}

// swiftlint:disable identifier_name
// https://alpaca.markets/docs/api-references/market-data-api/stock-pricing-data/realtime/#server-to-client
struct AlpacaApiServerMessages {
    enum AlpacaApiServerMessageType: String, Decodable, CodingKey {
        case success, error, subscription, t, q, b, d, s
    }
    
    struct AlpacaApiServerMessageDecoder: Decodable {
        var messages: [AlpacaApiServerMessages]
        
        init(from decoder: Decoder) throws {
            var container = try decoder.unkeyedContainer()
            
            messages = []
            if let count = container.count {
                messages.reserveCapacity(count)
            }
            
            while !container.isAtEnd {
                if let value = try? container.decode(AlpacaApiServerMessagesSuccess.self) {
                    messages.append(AlpacaApiServerMessages.success(value))
                } else if let value = try? container.decode(AlpacaApiServerMessagesError.self) {
                    messages.append(AlpacaApiServerMessages.error(value))
                } else if let value = try? container.decode(AlpacaApiServerMessagesTrade.self) {
                    messages.append(AlpacaApiServerMessages.trade(value))
                } else if let value = try? container.decode(AlpacaApiServerMessagesQuote.self) {
                    messages.append(AlpacaApiServerMessages.quote(value))
                } else if let value = try? container.decode(AlpacaApiServerMessagesSubscription.self) {
                    messages.append(AlpacaApiServerMessages.subscription(value))
                } else if let value = try? container.decode([String: AnyCodable].self) {
                    let encodedValue = try JSONEncoder().encode(value)
                    let encodedString = String.init(data: encodedValue, encoding: String.Encoding.utf8)!
                    throw DecodingError.dataCorrupted(
                        DecodingError.Context(
                            codingPath: container.codingPath,
                            debugDescription: "Data doesn't match. Received unknown message type: \(value["T"]?.value ?? "Unknown type"). Whole object: \(encodedString)"
                        )
                    )
                } else {
                    throw DecodingError.dataCorrupted(
                        DecodingError.Context(codingPath: container.codingPath, debugDescription: "Data doesn't match any known value")
                    )
                }
            }
            
        }
    }
    
    enum AlpacaApiServerMessages: Decodable {
        case success(AlpacaApiServerMessagesSuccess)
        case error(AlpacaApiServerMessagesError)
        case subscription(AlpacaApiServerMessagesSubscription)
        case trade(AlpacaApiServerMessagesTrade)
        case quote(AlpacaApiServerMessagesQuote)
    }
    
    struct AlpacaApiServerMessageBase: Decodable {
        var T: AlpacaApiServerMessageType
    }
    
    struct AlpacaApiServerMessagesSuccess: Decodable {
        var T: AlpacaApiServerMessageType = .success
        var msg: String
    }
    
    struct AlpacaApiServerMessagesError: Decodable {
        var T: AlpacaApiServerMessageType = .error
        var msg: String
        var code: Int
    }
    
    struct AlpacaApiServerMessagesSubscription: Decodable {
        var T: AlpacaApiServerMessageType = .subscription
        var trades: [String]?
        var quotes: [String]?
        var bars: [String]?
        var updatedBars: [String]?
        var dailyBars: [String]?
        var statuses: [String]?
        var lulds: [String]?
        var corrections: [String]?
        var cancelErrors: [String]?
        
        enum CodingKeys: CodingKey {
            case T
            case trades
            case quotes
            case bars
            case updatedBars
            case dailyBars
            case statuses
            case lulds
            case corrections
            case cancelErrors
        }
        
        init(from decoder: Decoder) throws {
            let container: KeyedDecodingContainer<AlpacaApiServerMessagesSubscription.CodingKeys> = try decoder.container(keyedBy: AlpacaApiServerMessagesSubscription.CodingKeys.self)
            self.T = try container.decode(AlpacaApiServerMessageType.self, forKey: CodingKeys.T)
            
            if self.T != .subscription {
                throw DecodingError.dataCorrupted(
                    DecodingError.Context(
                        codingPath: container.codingPath,
                        debugDescription: "AlpacaApiServerMessagesSubscription must have type \(AlpacaApiServerMessageType.subscription), found \(self.T)"
                    )
                )
            }

            self.trades = try container.decodeIfPresent([String].self, forKey: CodingKeys.trades)
            self.quotes = try container.decodeIfPresent([String].self, forKey: CodingKeys.quotes)
            self.bars = try container.decodeIfPresent([String].self, forKey: CodingKeys.bars)
            self.updatedBars = try container.decodeIfPresent([String].self, forKey: CodingKeys.updatedBars)
            self.dailyBars = try container.decodeIfPresent([String].self, forKey: CodingKeys.dailyBars)
            self.statuses = try container.decodeIfPresent([String].self, forKey: CodingKeys.statuses)
            self.lulds = try container.decodeIfPresent([String].self, forKey: CodingKeys.lulds)
            self.corrections = try container.decodeIfPresent([String].self, forKey: CodingKeys.corrections)
            self.cancelErrors = try container.decodeIfPresent([String].self, forKey: CodingKeys.cancelErrors)
        }
    }
    
    struct AlpacaApiServerMessagesTrade: Decodable {
        var T: AlpacaApiServerMessageType = .t
        var i: Int
        var S: String
        var x: String
        var p: Float
        var s: Int
        var t: String
        var c: [String]
        var z: String
    }
    
    struct AlpacaApiServerMessagesQuote: Decodable {
        var T: AlpacaApiServerMessageType = .q
        var bx: String
        var ap: Float
        var ax: String
        var `as`: Int
        var t: String
        var bs: Int
        var z: String
        var S: String
        var c: [String]
        var bp: Float
    }
}

// https://alpaca.markets/docs/api-references/market-data-api/stock-pricing-data/realtime/#client-to-server
struct AlpacaApiClientMessages {
    struct Auth: Encodable {
        var action = "auth"
        var key: String
        var secret: String
    }
    
    struct Subscribe: Encodable {
        var action = "subscribe"
        var trades: [String]?
        var quotes: [String]?
        var bars: [String]?
        var dailyBars: [String]?
        var statuses: [String]?
    }
}

// swiftlint:enable identifier_name
