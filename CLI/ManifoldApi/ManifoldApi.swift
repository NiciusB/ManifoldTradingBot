import Foundation
import Alamofire

struct ManifoldApi {
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
            headers: headers
        )

        if logResponse {
            request.responseString { response in
                print(response)
            }
        }

        request
            .validate()
            .responseData { response in
                switch response.result {
                case .success:
                    print("Validation Successful")
                case let .failure(error):
                    print("Error:", error)
                }
            }

        let dataTask = request.serializingDecodable(endpoint.responseDecodable)
        return try await dataTask.value
    }

    func getMe() async throws -> GetMe.ResDec {
        return try await self.request(GetMe())
    }

    func placeBet(_ req: PlaceBet.RequestParams) async throws -> PlaceBet.ResDec {
        return try await self.request(PlaceBet(req))
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
