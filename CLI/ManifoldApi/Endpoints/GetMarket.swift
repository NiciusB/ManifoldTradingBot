import Foundation
import Alamofire

struct GetMarket: ManifoldApiEndpoint {
    init(_ id: String) {
        path = "/v0/market/" + id // + "?cacheBust=" + String(Int(Date().timeIntervalSince1970))
    }

    var method = HTTPMethod.post
    var path: String
    var responseDecodable = Root.self
    var body: String?
    
    // swiftlint:disable identifier_name
    struct Root: Decodable {
      var id: String
      var creatorId: String
      var creatorUsername: String
      var creatorName: String
      var createdTime: Float
      var creatorAvatarUrl: String
      var closeTime: Float
      var question: String
      var tags: [String]
      var url: String
      var pool: Pool
      var probability: Float
      var p: Float
      var totalLiquidity: Float
      var outcomeType: String
      var mechanism: String
      var volume: Float
      var volume24Hours: Float
      var isResolved: Bool
      var lastUpdatedTime: Float
      var value: Float
      var min: Float
      var max: Float
      var isLogScale: Bool
      var description: Description
      var coverImageUrl: String
      var textDescription: String
    }

    struct Pool: Decodable {
      var YES: Float
      var NO: Float
    }

    struct Description: Decodable {
      var content: [Content]
      var type: String
    }

    struct Content: Decodable {
      var type: String
      var content: [Content2]?
      var attrs: Attrs?
    }

    struct Content2: Decodable {
      var text: String?
      var type: String
    }

    struct Attrs: Decodable {
      var contractIds: String
    }
    // swiftlint:enable identifier_name

}
