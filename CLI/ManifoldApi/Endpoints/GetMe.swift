import Foundation
import Alamofire

struct GetMe: ManifoldApiEndpoint {
    var method = HTTPMethod.get
    var path = "/v0/me"
    var body: String?
    var responseDecodable = Root.self

    struct Root: Decodable {
        let achievements: Achievements
        let avatarUrl: String
        let isBannedFromPosting: Bool
        let streakForgiveness: Float
        let profitCached: ProfitCached
        let creatorTraders: CreatorTraders
        let createdTime: Float
        let id: String
        let nextLoanCached: Float
        let shouldShowWelcome: Bool
        let name: String
        let username: String
        let bio: String
        let currentBettingStreak: Float
        let lastBetTime: Float
        let totalDeposits: Float
        let balance: Float
        let followerCountCached: Float
        let metricsLastUpdated: Float
    }

    struct Achievements: Decodable {}

    struct ProfitCached: Decodable {
        let daily: Float
        let monthly: Float
        let weekly: Float
        let allTime: Float
    }

    struct CreatorTraders: Decodable {
        let daily: Float
        let monthly: Float
        let weekly: Float
        let allTime: Float
    }
}
