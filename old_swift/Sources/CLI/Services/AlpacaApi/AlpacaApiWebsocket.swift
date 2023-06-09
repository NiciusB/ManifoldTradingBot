import Foundation
#if canImport(FoundationNetworking)
    import FoundationNetworking
#endif
import AnyCodable
import Vapor

// I don't think this class is actually thread-safe, but it shouldn't cause problems because we mostly use it sequentially
final class AlpacaApiWebsocket: @unchecked Sendable {
    private let apiKey: String
    private let apiSecret: String
    private var webSocket: WebSocket?
    let loginTask = Task(priority: .background) {
        while true {
            try await Task.sleep(nanoseconds: 1000 * 1000 * 100)
            await Task.yield()
        }
    }
    private var lastTradeValues: [String: Float] = [:]
    
    init(apiKey: String, apiSecret: String) async throws {
        self.apiSecret = apiSecret
        self.apiKey = apiKey
        
        let wsUrl = "wss://stream.data.alpaca.markets/v2/iex"
        let elg  = MultiThreadedEventLoopGroup(numberOfThreads: 1)
        try await WebSocket.connect(to: wsUrl, on: elg ) { ws in
            self.webSocket = ws
            ws.onText({ _, strMessage in
                Task {
                    do {
                        let data = strMessage.data(using: .utf8)!
                        let decoded = try JSONDecoder().decode(
                            AlpacaApiServerMessages.AlpacaApiServerMessageDecoder.self,
                            from: data
                        )
                        
                        for decodedMsg in decoded.messages {
                            try await self.processMessage(decodedMsg)
                        }
                    } catch {
                        printErr(error)
                        exit(1)
                    }
                }
            })
            
            ws.onClose.whenComplete { _ in
                printErr("Alpaca Websocket closed")
                exit(1)
            }
        }
        
        try? await loginTask.value
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
        try await webSocket!.send(String(data: data, encoding: .utf8)!)
    }
    
    private func sendAuthMsg() async throws {
        let authMsg = AlpacaApiClientMessages.Auth(key: self.apiKey, secret: self.apiSecret)
        let data = try JSONEncoder().encode(authMsg)
        try await webSocket!.send(String(data: data, encoding: .utf8)!)
    }
    
    func getStockLastTradeValue(_ stock: String) -> Float? {
        return lastTradeValues[stock]
    }
    
    private func processMessage(_ decodedMsg: AlpacaApiServerMessages.AlpacaApiServerMessages) async throws {
            switch decodedMsg {
            case let .success(data):
                if data.msg == "connected" {
                    try await self.sendAuthMsg()
                } else if data.msg == "authenticated" {
                    loginTask.cancel()
                } else {
                    let dataString = try String(data: JSONEncoder().encode(data), encoding: .utf8)!
                    throw RuntimeError("Unknown Alpaca API success message: " + dataString)
                }
            case let .error(data):
                let dataString = try String(data: JSONEncoder().encode(data), encoding: .utf8)!
                throw RuntimeError("Alpaca API error: " + dataString)
            case .subscription:
                // subscribed!
                break
            case let .trade(data):
                lastTradeValues[data.S] = data.p
            case let .quote(data):
                print(data)
            }
    }
}

// swiftlint:disable identifier_name
// https://alpaca.markets/docs/api-references/market-data-api/stock-pricing-data/realtime/#server-to-client
struct AlpacaApiServerMessages {
    enum AlpacaApiServerMessageType: String, Codable, CodingKey {
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
    
    enum AlpacaApiServerMessages: Codable {
        case success(AlpacaApiServerMessagesSuccess)
        case error(AlpacaApiServerMessagesError)
        case subscription(AlpacaApiServerMessagesSubscription)
        case trade(AlpacaApiServerMessagesTrade)
        case quote(AlpacaApiServerMessagesQuote)
    }
    
    struct AlpacaApiServerMessageBase: Codable {
        var T: AlpacaApiServerMessageType
    }
    
    struct AlpacaApiServerMessagesSuccess: Codable {
        var T: AlpacaApiServerMessageType = .success
        var msg: String
        
        enum CodingKeys: CodingKey {
            case T
            case msg
        }
        
        init(from decoder: Decoder) throws {
            let container: KeyedDecodingContainer<AlpacaApiServerMessagesSuccess.CodingKeys> = try decoder.container(keyedBy: AlpacaApiServerMessagesSuccess.CodingKeys.self)
            self.T = try container.decode(AlpacaApiServerMessageType.self, forKey: AlpacaApiServerMessagesSuccess.CodingKeys.T)
            
            if self.T != .success {
                throw DecodingError.dataCorrupted(
                    DecodingError.Context(
                        codingPath: container.codingPath,
                        debugDescription: "AlpacaApiServerMessagesSubscription must have type \(AlpacaApiServerMessageType.success), found \(self.T)"
                    )
                )
            }
            
            self.msg = try container.decode(String.self, forKey: AlpacaApiServerMessagesSuccess.CodingKeys.msg)
        }
    }
    
    struct AlpacaApiServerMessagesError: Codable {
        var T: AlpacaApiServerMessageType = .error
        var msg: String
        var code: Int
        
        enum CodingKeys: CodingKey {
            case T
            case msg
            case code
        }
        
        init(from decoder: Decoder) throws {
            let container: KeyedDecodingContainer<AlpacaApiServerMessagesError.CodingKeys> = try decoder.container(keyedBy: AlpacaApiServerMessagesError.CodingKeys.self)
            self.T = try container.decode(AlpacaApiServerMessageType.self, forKey: AlpacaApiServerMessagesError.CodingKeys.T)
            
            if self.T != .error {
                throw DecodingError.dataCorrupted(
                    DecodingError.Context(
                        codingPath: container.codingPath,
                        debugDescription: "AlpacaApiServerMessagesSubscription must have type \(AlpacaApiServerMessageType.error), found \(self.T)"
                    )
                )
            }
            
            self.msg = try container.decode(String.self, forKey: AlpacaApiServerMessagesError.CodingKeys.msg)
            self.code = try container.decode(Int.self, forKey: AlpacaApiServerMessagesError.CodingKeys.code)
        }
    }
    
    struct AlpacaApiServerMessagesSubscription: Codable {
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
    
    struct AlpacaApiServerMessagesTrade: Codable {
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
    
    struct AlpacaApiServerMessagesQuote: Codable {
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
