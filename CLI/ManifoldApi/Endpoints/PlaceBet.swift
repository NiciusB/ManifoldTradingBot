import Foundation
import Alamofire

struct PlaceBet: ManifoldApiEndpoint {
    struct RequestParams: Encodable {
        let amount: Int
        let contractId: String
        let outcome: String
        let limitProb: Float?
    }

    init(_ req: RequestParams) {
        body = req
    }

    var method = HTTPMethod.post
    var path = "/v0/bet"
    var responseDecodable = Root.self
    var body: RequestParams?

    struct Root: Decodable {
       var TODODefineDecodable: Data
    }
}
