import Foundation
import Alamofire
import SwiftSoup

class SharesOutstandingApi {
    var cachedOutstandingSharesValues: [String: (Date, Int?)] = [:]
    
    func getSymbolOutstandingShares(_ symbol: String) async -> Int? {
        let cached = cachedOutstandingSharesValues[symbol]
        if cached != nil {
            let cacheSetTime = cached!.0
            let components = Calendar.current.dateComponents([.second], from: cacheSetTime, to: Date())
            if components.second! < 60 * 10 {
                return cached!.1
            }
        }
        
        do {
            let urlPath = "https://stockanalysis.com/stocks/\(symbol)/statistics/"
            let headers = HTTPHeaders([
                "User-Agent": "ManifoldTradingBot/1.0.0 for @NiciusBot"
            ])
            let request = AF.request(
                urlPath,
                method: .get,
                headers: headers
            )
            
            let dataTask = request.serializingData()
            let dataResponse = await dataTask.response
            let resData = try dataResponse.result.get()
            let htmlString = String(data: resData, encoding: String.Encoding.utf8)!
            let doc = try SwiftSoup.parse(htmlString)
            let sharesOutstandingRow = try doc.getElementsMatchingOwnText("Shares Outstanding").first()?.parent()?.nextElementSibling()
            let titleAttr = try sharesOutstandingRow!.attr("title")
            let parsedAmount = Int(titleAttr.replacingOccurrences(of: ",", with: ""))
            self.cachedOutstandingSharesValues[symbol] = (Date(), parsedAmount)
            return parsedAmount
        } catch {
            self.cachedOutstandingSharesValues[symbol] = (Date(), nil)
            return nil
        }
    }
}
