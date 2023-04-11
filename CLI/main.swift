import Dotenv
import Alamofire
import Foundation

let dotenv = try Dotenv()
let manifoldApi = ManifoldApi(
    apiKey: dotenv.get("MANIFOLD_API_KEY")!
)

let result = try await manifoldApi.placeBet(PlaceBet.RequestParams(amount: 312231231, contractId: "abc", outcome: "x"))

print(result)
