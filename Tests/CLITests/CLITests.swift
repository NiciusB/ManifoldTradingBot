import XCTest
@testable import CLI

final class CLITests: XCTestCase {
    // XCTest Documenation
    // https://developer.apple.com/documentation/xctest
    
    // Defining Test Cases and Test Methods
    // https://developer.apple.com/documentation/xctest/defining_test_cases_and_test_methods
    
    func testPRemainsStable() throws {
        for numbers in 0...100 {
            let origP: Float = Float(numbers) / 100
            let pool = GetMarket.Pool(YES: 123, NO: 123)
            let probability = CpmmMarketUtils.calcMarketProbabilityFromMarketP(origP, pool: pool)
            let newP = CpmmMarketUtils.calcMarketPFromMarketProbability(probability, pool: pool)
            
            XCTAssertLessThanOrEqual(abs(origP - newP), 0.0000001, "P should remain semi-stable")
        }
    }
    
    // swiftlint:disable:next function_body_length
    func testCalculatesCorrectSharesToBuy() throws {
        let googMarket: GetMarket.ResDec = GetMarket.ResDec(
            id: "id-1234",
            creatorId: "id-1234",
            creatorUsername: "creator",
            creatorName: "Creator",
            createdTime: Float(Date().timeIntervalSince1970.magnitude),
            creatorAvatarUrl: "https://example.com/avatar.jpg",
            closeTime: Float(Date().timeIntervalSince1970.magnitude) + 1000 * 60 * 24,
            question: "Example question",
            tags: ["Bets"],
            url: "https://manifold.markets/PatMyron/current-alphabet-google-market-cap",
            pool: GetMarket.Pool(YES: 427.39884559678626, NO: 469.26700725062585),
            probability: 0.1381288306058536,
            p: 0.12733858779397111,
            totalLiquidity: 481,
            outcomeType: "PSEUDO_NUMERIC",
            mechanism: "cpmm-1",
            volume: 509.2376203441685,
            volume24Hours: 21.497127860332313,
            isResolved: false,
            lastUpdatedTime: Float(Date().timeIntervalSince1970.magnitude) - 1000 * 60 * 24,
            value: 1381.90,
            min: 0,
            max: 10000,
            isLogScale: false,
            description: GetMarket.Description(content: [], type: ""),
            coverImageUrl: "https://example.com/cover.jpg",
            textDescription: "Example description"
        )
        
        var (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(googMarket, targetValue: 1372.9418)
        
        XCTAssertEqual(outcome, "NO", "Bet should be no")
        XCTAssertEqual(betAmount, 3, "Bet should be specific number")

        (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(googMarket, targetValue: 1390)
        
        XCTAssertEqual(outcome, "YES", "Bet should be yes")
        XCTAssertEqual(betAmount, 0, "Bet should be specific number")

        (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(googMarket, targetValue: 1450)
        
        XCTAssertEqual(outcome, "YES", "Bet should be yes")
        XCTAssertEqual(betAmount, 3, "Bet should be specific number")

        (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(googMarket, targetValue: 2910)
        
        XCTAssertEqual(outcome, "YES", "Bet should be yes")
        XCTAssertEqual(betAmount, 60, "Bet should be specific number")

        let syntheticFakeMarket: GetMarket.ResDec = GetMarket.ResDec(
            id: "id-1234",
            creatorId: "id-1234",
            creatorUsername: "creator",
            creatorName: "Creator",
            createdTime: Float(Date().timeIntervalSince1970.magnitude),
            creatorAvatarUrl: "https://example.com/avatar.jpg",
            closeTime: Float(Date().timeIntervalSince1970.magnitude) + 1000 * 60 * 24,
            question: "Example question",
            tags: ["Bets"],
            url: "https://example.com/syntheticFakeMarket/XYZ",
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
        
        (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(syntheticFakeMarket, targetValue: 100)
        
        XCTAssertEqual(outcome, "NO", "Bet should be no")
        XCTAssertEqual(betAmount, 4269, "Bet should be specific number")
        
        (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(syntheticFakeMarket, targetValue: 1234)
        
        XCTAssertEqual(outcome, "NO", "Bet should be no")
        XCTAssertEqual(betAmount, 593, "Bet should be specific number")
        
        (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(syntheticFakeMarket, targetValue: 5000)
        
        XCTAssertEqual(outcome, "YES", "Bet should be yes")
        XCTAssertEqual(betAmount, 32, "Bet should be specific number")
        
        (outcome, betAmount) = CpmmMarketUtils.calculatePseudoNumericMarketplaceBet(syntheticFakeMarket, targetValue: 9000)
        
        XCTAssertEqual(outcome, "YES", "Bet should be yes")
        XCTAssertEqual(betAmount, 596, "Bet should be specific number")
    }
}
