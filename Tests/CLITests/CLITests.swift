import XCTest
@testable import CLI

final class CLITests: XCTestCase {
    // XCTest Documenation
    // https://developer.apple.com/documentation/xctest
    
    // Defining Test Cases and Test Methods
    // https://developer.apple.com/documentation/xctest/defining_test_cases_and_test_methods
    
    func testPRemainsStable() throws {
        for numbers in 0...100 {
            let p: Float = Float(numbers) / 100
            let pool = GetMarket.Pool(YES: 123, NO: 123)
            let probability = CpmmMarketUtils.calcMarketProbabilityFromMarketP(p, pool: pool)
            let newP = CpmmMarketUtils.calcMarketPFromMarketProbability(probability, pool: pool)
            
            XCTAssertLessThanOrEqual(abs(p - newP), 0.0000001, "P should remain semi-stable")
        }
    }
    
    func testCalculatesCorrectSharesToBuy() throws {
        let market: GetMarket.ResDec = GetMarket.ResDec(
            id: "id-1234",
            creatorId: "id-1234",
            creatorUsername: "creator",
            creatorName: "Creator",
            createdTime: Float(Date().timeIntervalSince1970.magnitude),
            creatorAvatarUrl: "https://example.com/avatar.jpg",
            closeTime: Float(Date().timeIntervalSince1970.magnitude) + 1000 * 60 * 24,
            question: "Example question",
            tags: ["Bets"],
            url: "https://example.com/market/XYZ",
            pool: GetMarket.Pool(YES: 300, NO: 400),
            probability: 0.45,
            p: 0.4,
            totalLiquidity: 300,
            outcomeType: "PSEUDO_NUMERIC",
            mechanism: "cpmm-1",
            volume: 800,
            volume24Hours: 123,
            isResolved: false,
            lastUpdatedTime: Float(Date().timeIntervalSince1970.magnitude) - 1000 * 60 * 24,
            value: 4000,
            min: 0,
            max: 10000,
            isLogScale: false,
            description: GetMarket.Description(content: [], type: ""),
            coverImageUrl: "https://example.com/cover.jpg",
            textDescription: "Example description"
        )
        
        var (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(market, targetValue: 100)
        
        XCTAssertEqual(outcome, "NO", "Bet should be yes")
        XCTAssertEqual(betAmount, 5161, "Bet should be specific number")
        
        (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(market, targetValue: 1234)
        
        XCTAssertEqual(outcome, "NO", "Bet should be yes")
        XCTAssertEqual(betAmount, 767, "Bet should be specific number")
        
        (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(market, targetValue: 5000)
        
        XCTAssertEqual(outcome, "YES", "Bet should be yes")
        XCTAssertEqual(betAmount, 0, "Bet should be specific number")
        
        (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(market, targetValue: 9000)
        
        XCTAssertEqual(outcome, "YES", "Bet should be yes")
        XCTAssertEqual(betAmount, 492, "Bet should be specific number")
    }
}
