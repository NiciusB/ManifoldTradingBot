import Foundation

struct CpmmMarketUtils {
    // Cpmm Probability formula based on market's P and pool. Formula from https://github.com/manifoldmarkets/manifold/blob/main/common/src/calculate-cpmm.ts#L14
    // swiftlint:disable:next identifier_name
    static func calcMarketProbabilityFromMarketP(_ p: Float, pool: GetMarket.Pool) -> Float {
        return (p * pool.NO) / ((1 - p) * pool.YES + p * pool.NO)
    }
    // P formula based on market's Cpmm Probability and pool. Previous formula solved for p
    static func calcMarketPFromMarketProbability(_ probability: Float, pool: GetMarket.Pool) -> Float {
        // https://www.wolframalpha.com/input?i=b+%3D+(p+*+n)+/+((1+-+p)+*+y+++p+*+n),+solve+p
        return (probability * pool.YES) / (-probability * pool.NO + probability * pool.YES + pool.NO)
    }
    
    static func calculatePseudoNumericMarketplaceValue(_ market: GetMarket.ResDec) -> Float {
        assert(market.outcomeType == "PSEUDO_NUMERIC", "Unexpected outcomeType \(market.outcomeType). Stopping to prevent messing up the formulas and betting wrong")
        assert(market.mechanism == "cpmm-1", "Unexpected mechanism \(market.outcomeType). Stopping to prevent messing up the formulas and betting wrong")
        let marketProbability = market.probability!
        let marketMin: Float = market.min!
        let marketMax = market.max!
        
        // Sanity check
        let calculatedProbability = CpmmMarketUtils.calcMarketProbabilityFromMarketP(market.p, pool: market.pool)
        assert(
            abs(calculatedProbability - marketProbability) < 0.001,
            "Sanity check error: Differing probabilities \(calculatedProbability) vs \(marketProbability), diff \(abs(calculatedProbability - marketProbability)) is greater than 0.001. Formulas might have changed. Stopping to prevent messing up the formulas and betting wrong"
        )
        
        // PSEUDO_NUMERIC value formulas based on market's cpmm probability, min and max. Formulas from https://github.com/manifoldmarkets/manifold/blob/main/common/src/pseudo-numeric.ts#L13
        if market.isLogScale! {
            return pow(10, marketProbability * log10(marketMax - marketMin + 1)) + marketMin - 1
        } else {
            return marketProbability * (marketMax - marketMin) + marketMin
        }
    }
    
    static func calculatePseudoNumericMarketplaceBet(_ market: GetMarket.ResDec, targetValue: Float) -> (String, Int) {
        assert(market.outcomeType == "PSEUDO_NUMERIC", "Unexpected outcomeType \(market.outcomeType). Stopping to prevent messing up the formulas and betting wrong")
        assert(market.mechanism == "cpmm-1", "Unexpected mechanism \(market.outcomeType). Stopping to prevent messing up the formulas and betting wrong")
        let marketProbability = market.probability!
        let marketMin: Float = market.min!
        let marketMax = market.max!
        
        // Manually inversed PSEUDO_NUMERIC value formulas from calculatePseudoNumericMarketplaceValue, to solve for market cpmm probability based on a target value
        let targetProbability: Float
        if market.isLogScale! {
            targetProbability = log(targetValue - marketMin + 1) / log(10) / log10(marketMax - marketMin + 1)
            
            print(
                "First I got from wolframalpha, second I solved myself",
                log(targetValue - marketMin + 1) / log(marketMax - marketMin + 1),
                log(targetValue - marketMin + 1) / log(10) / log10(marketMax - marketMin + 1)
            )
            
            assert(false, "The code should work for logScale with minor fixes but it's not tested so idk, just crash for now")
        } else {
            targetProbability = (targetValue - marketMin) / (marketMax - marketMin)
        }
        
        // Manually inversed formulas from calculatePseudoNumericMarketplaceValue, to solve for YES/NO votes on a target cpmm probability
        let outcome = marketProbability < targetProbability ? "YES" : "NO"
        
        let amount = calculateCpmmAmountToProb(
            state: CpmmMarketUtils.CpmmState(
                p: CpmmMarketUtils.calcMarketPFromMarketProbability(marketProbability, pool: market.pool),
                pool: Pool(YES: market.pool.YES, NO: market.pool.NO)
            ),
            prob: targetProbability,
            outcome: outcome
        )
        
        return (outcome, Int(round(amount)))
    }
    
    // Following functions copied from https://github.com/manifoldmarkets/manifold/blob/0a71fdd0e3d684145022dcb9f27bb8bb14835d50/common/src/calculate-cpmm.ts and translated to Swift
    // swiftlint:disable identifier_name
    
    private static let CREATOR_FEE: Float = 0
    private static let PLATFORM_FEE: Float = 0
    private static let LIQUIDITY_FEE: Float = 0
    
    static func calculateCpmmShares(
        pool: Pool,
        p: Float,
        bet: Float,
        outcome: String
    ) -> Float {
        let k = pow(pool.YES, p) * pow(pool.NO, (1 - p))
        
        if outcome == "YES" {
            return pool.YES + bet - pow(k * pow(bet + pool.NO, (p - 1)), (1 / p))
        } else {
            return pool.NO + bet - pow(k * pow(bet + pool.YES, -p), (1 / (1 - p)))
        }
    }
    
    static func calculateCpmmAmountToProb(
        state: CpmmState,
        prob argProb: Float,
        outcome: String
    ) -> Float {
        var prob = argProb
        if prob <= 0 || prob >= 1 {
            return Float.infinity
        }
        if outcome == "NO" {
            prob = 1 - prob
        }
        
        // First, find an upper bound that leads to a more extreme probability than prob.
        var maxGuess: Float = 10
        var maxProb: Float = 0
        repeat {
            maxGuess *= 10
            maxProb = getCpmmOutcomeProbabilityAfterBet(state: state, outcome: outcome, bet: maxGuess)
        } while (maxProb < prob)
        
        // Then, binary search for the amount that gets closest to prob.
        let amount = binarySearch(min: 0, max: maxGuess, comparator: { amount in
            let newProb = getCpmmOutcomeProbabilityAfterBet(state: state, outcome: outcome, bet: amount)
            
            return newProb - prob
        })
        
        return amount
    }
    
    static func getCpmmOutcomeProbabilityAfterBet(
        state: CpmmState,
        outcome: String,
        bet: Float
    ) -> Float {
        let newPool = calculateCpmmPurchase(state: state, bet: bet, outcome: outcome).newPool
        let p = getCpmmProbability(pool: newPool, p: state.p)
        return outcome == "YES" ? p : 1 - p
    }
    
    struct PurchaseInfo {
        var shares: Float
        var newPool: Pool
        var newP: Float
        var fees: Fees
    }
    static func calculateCpmmPurchase(
        state: CpmmState,
        bet: Float,
        outcome: String
    ) -> PurchaseInfo {
        let feesInfo = getCpmmFees(state: state, bet: bet, outcome: outcome)
        
        let shares = calculateCpmmShares(pool: state.pool, p: state.p, bet: feesInfo.remainingBet, outcome: outcome)
        
        let newY: Float
        let newN: Float
        if outcome == "YES" {
            newY = state.pool.YES - shares + feesInfo.remainingBet + feesInfo.fees.liquidityFee
            newN = state.pool.NO + feesInfo.remainingBet + feesInfo.fees.liquidityFee
        } else {
            newY = state.pool.YES + feesInfo.remainingBet + feesInfo.fees.liquidityFee
            newN = state.pool.NO - shares + feesInfo.remainingBet + feesInfo.fees.liquidityFee
        }
        
        let postBetPool = Pool(YES: newY, NO: newN)
        
        let addInfo = addCpmmLiquidity(pool: postBetPool, p: state.p, amount: feesInfo.fees.liquidityFee)
        
        return PurchaseInfo(shares: shares, newPool: addInfo.newPool, newP: addInfo.newP, fees: feesInfo.fees)
    }
    
    struct AddCpmmLiquidityInfo {
        var newPool: Pool
        var newP: Float
        var liquidity: Float
    }
    
    static func addCpmmLiquidity(
        pool: Pool,
        p: Float,
        amount: Float
    ) -> AddCpmmLiquidityInfo {
        let prob = getCpmmProbability(pool: pool, p: p)
        
        // https://www.wolframalpha.com/input?i=p%28n%2Bb%29%2F%28%281-p%29%28y%2Bb%29%2Bp%28n%2Bb%29%29%3Dq%2C+solve+p
        let numerator = prob * (amount + pool.YES)
        let denominator = amount - pool.NO * (prob - 1) + prob * pool.YES
        let newP = numerator / denominator
        
        let newPool = Pool(YES: pool.YES + amount, NO: pool.NO + amount)
        
        let oldLiquidity = getCpmmLiquidity(pool: pool, p: newP)
        let newLiquidity = getCpmmLiquidity(pool: newPool, p: newP)
        let liquidity = newLiquidity - oldLiquidity
        
        return AddCpmmLiquidityInfo(newPool: newPool, newP: liquidity, liquidity: newP)
    }
    
    static func getCpmmLiquidity(
        pool: Pool,
        p: Float
    ) -> Float {
        return pow(pool.YES, p) *  pow(pool.NO, (1 - p))
    }
    
    struct FeesInfo {
        var remainingBet: Float
        var totalFees: Float
        var fees: Fees
    }
    static func getCpmmFees(state: CpmmState, bet: Float, outcome: String) -> FeesInfo {
        let prob = getCpmmProbabilityAfterBetBeforeFees(state: state, outcome: outcome, bet: bet)
        let betP = outcome == "YES" ? 1 - prob : prob
        
        let liquidityFee = LIQUIDITY_FEE * betP * bet
        let platformFee = PLATFORM_FEE * betP * bet
        let creatorFee = CREATOR_FEE * betP * bet
        let fees: Fees = Fees(liquidityFee: liquidityFee, platformFee: platformFee, creatorFee: creatorFee )
        
        let totalFees = liquidityFee + platformFee + creatorFee
        let remainingBet = bet - totalFees
        
        return FeesInfo(remainingBet: remainingBet, totalFees: totalFees, fees: fees)
    }
    
    static func getCpmmProbabilityAfterBetBeforeFees(
        state: CpmmState,
        outcome: String,
        bet: Float
    ) -> Float {
        let shares = calculateCpmmShares(pool: state.pool, p: state.p, bet: bet, outcome: outcome)
        
        let newY: Float
        let newN: Float
        if outcome == "YES" {
            newY = state.pool.YES - shares + bet
            newN = state.pool.NO + bet
        } else {
            newY = state.pool.YES + bet
            newN = state.pool.NO - shares + bet
        }
        
        return getCpmmProbability(pool: Pool(YES: newY, NO: newN), p: state.p)
    }
    
    static func getCpmmProbability (
        pool: Pool,
        p: Float
    ) -> Float {
        return (p * pool.NO) / ((1 - p) * pool.YES + p * pool.NO)
    }
    
    struct Fees {
        var liquidityFee: Float
        var platformFee: Float
        var creatorFee: Float
    }
    struct Pool {
        var YES: Float
        var NO: Float
    }
    struct CpmmState {
        var p: Float
        var pool: Pool
    }
    
    static func binarySearch(
        min argMin: Float,
        max argMax: Float,
        comparator: (Float) -> Float
    ) -> Float {
        var min = argMin
        var max = argMax
        var mid: Float = 0
        while true {
            mid = min + (max - min) / 2
            
            // Break once we've reached max precision.
            if mid == min || mid == max {
                break
            }
            
            let comparison = comparator(mid)
            if comparison == 0 {
                break
            } else if comparison > 0 {
                max = mid
            } else {
                min = mid
            }
        }
        
        return mid
    }
    // swiftlint:enable identifier_name
}
