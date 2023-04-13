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

    let method = HTTPMethod.post
    let path = "/v0/bet"
    let responseDecodable = Root.self
    let body: RequestParams?

    struct Root: Decodable {
        let orderAmount: Float
        let amount: Float
        let shares: Float
        let isFilled: Bool
        let isCancelled: Bool
        let fills: [Fill]
        let contractId: String
        let outcome: String
        let probBefore: Float
        let probAfter: Float
        let loanAmount: Float
        let createdTime: Float
        let fees: Fees
        let isAnte: Bool
        let isRedemption: Bool
        let isChallenge: Bool
        let visibility: String
        let betId: String
    }
 
    struct Fill: Decodable {
        let matchedBetId: String?
        let shares: Float
        let amount: Float
        let timestamp: Float
    } 
 
    struct Fees: Decodable {
        let creatorFee: Float
        let platformFee: Float
        let liquidityFee: Float
    }

}
