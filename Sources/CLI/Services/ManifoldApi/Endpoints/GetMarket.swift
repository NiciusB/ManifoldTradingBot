import Foundation
import Alamofire

struct GetMarket: ManifoldApiEndpoint {
    init(_ id: String) {
        path = "/v0/market/" + id // + "?cacheBust=" + String(Int(Date().timeIntervalSince1970))
    }

    let method = HTTPMethod.post
    let path: String
    let responseDecodable = Root.self
    let body: String? = nil
    
    // swiftlint:disable identifier_name
    struct Root: Decodable {
      let id: String
      let creatorId: String
      let creatorUsername: String
      let creatorName: String
      let createdTime: Float
      let creatorAvatarUrl: String
      let closeTime: Float
      let question: String
      let tags: [String]
      let url: String
      let pool: Pool
      let probability: Float?
      let p: Float
      let totalLiquidity: Float
      let outcomeType: String
      let mechanism: String
      let volume: Float
      let volume24Hours: Float
      let isResolved: Bool
      let lastUpdatedTime: Float
      let value: Float?
      let min: Float?
      let max: Float?
      let isLogScale: Bool?
      let description: Description
      let coverImageUrl: String
      let textDescription: String
    }

    struct Pool: Decodable {
      let YES: Float
      let NO: Float
    }

    struct Description: Decodable {
      let content: [Content]
      let type: String
    }

    struct Content: Decodable {
      let type: String
      let content: [Content2]?
      let attrs: Attrs?
    }

    struct Content2: Decodable {
      let text: String?
      let type: String
    }

    struct Attrs: Decodable {
      let contractIds: String
    }
    // swiftlint:enable identifier_name

}
