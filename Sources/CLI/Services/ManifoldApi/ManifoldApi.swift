import Foundation
import Alamofire

struct ManifoldApi: Sendable {
    let apiKey: String
    
    private func request<E: ManifoldApiEndpoint>(_ endpoint: E, logResponse: Bool = false) async throws -> E.ResDec {
        let urlPath = "https://manifold.markets/api" + endpoint.path
        let headers = HTTPHeaders([
            "Authorization": "Key " + self.apiKey,
            "User-Agent": "ManifoldTradingBot/1.0.0 for @NiciusBot",
            "Accept": "application/json"
        ])
        let parameters = endpoint.body
        let request = AF.request(
            urlPath,
            method: endpoint.method,
            parameters: parameters,
            encoder: JSONParameterEncoder.default,
            headers: headers,
            requestModifier: { $0.timeoutInterval = 10 }
        )
        
        if logResponse {
            request.responseString { response in
                print(response)
            }
        }
        
        let dataTask = request.serializingData()
        let dataResponse = await dataTask.response
        let resData = try dataResponse.result.get()
        
        let errorBody = try JSONDecoder().decode(Error400Response.self, from: resData)
        if dataResponse.response?.statusCode == 400 || errorBody.error != nil {
            throw ManifoldApiError.invalidRequest(
                message: errorBody.error ??
                errorBody.message ??
                ("Unknown error: " + String(decoding: resData, as: UTF8.self))
            )
        }
        
        do {
            return try JSONDecoder().decode(endpoint.responseDecodable, from: resData)
        } catch let error as DecodingError {
            throw ManifoldApiError.unableToDecodeResponse(
                decodingError: error,
                response: String(decoding: resData, as: UTF8.self)
            )
        }
    }
    
    func getMe() async throws -> GetMe.ResDec {
        return try await self.request(GetMe())
    }
    
    @discardableResult
    func placeBet(
        amount: Int, contractId: String, outcome: String, limitProb: Float? = nil
    ) async throws -> PlaceBet.ResDec {
        return try await self.request(PlaceBet(PlaceBet.RequestParams(
            amount: amount,
            contractId: contractId,
            outcome: outcome,
            limitProb: limitProb
        )))
    }
    
    func getMarket(_ id: String) async throws -> GetMarket.ResDec {
        return try await self.request(GetMarket(id))
    }
    
    struct Error400Response: Decodable {
        let message: String?
        let error: String?
    }

    enum ManifoldApiError: Error {
        case invalidRequest(message: String)
        case unableToDecodeResponse(decodingError: DecodingError, response: String)
    }
}

protocol ManifoldApiEndpoint {
    associatedtype ResDec: Decodable
    associatedtype BodyEncodable: Encodable
    
    var method: HTTPMethod { get }
    var path: String { get }
    var body: BodyEncodable? { get }
    var responseDecodable: ResDec.Type { get }
}
