package CpmmMarketUtils

import (
	"ManifoldTradingBot/ManifoldApi"
	"log"
	"math"
)

func assert(condition bool, message string) {
	if !condition {
		log.Fatalln(message)
	}
}

// Cpmm Probability formula based on market's P and  pool. Formula from https://github.com/manifoldmarkets/manifold/blob/main/common/src/calculate-cpmm.ts#L14
func calcMarketProbabilityFromMarketP(p float64, pool pool) float64 {
	return (p * pool.NO) / ((1-p)*pool.YES + p*pool.NO)
}

// P formula based on market's Cpmm Probability and  pool. Previous formula solved for p
func calcMarketPFromMarketProbability(probability float64, pool pool) float64 {
	// https://www.wolframalpha.com/input?i=b+%3D+(p+*+n)+/+((1+-+p)+*+y+++p+*+n),+solve+p
	return (probability * pool.YES) / (-probability*pool.NO + probability*pool.YES + pool.NO)
}

func CalculatePseudoNumericMarketplaceValue(market ManifoldApi.Market) float64 {
	assert(market.OutcomeType == "PSEUDO_NUMERIC", "Unexpected outcomeType (market.OutcomeType). Stopping to prevent messing up the formulas and betting wrong")
	assert(market.Mechanism == "cpmm-1", "Unexpected mechanism (market.OutcomeType). Stopping to prevent messing up the formulas and betting wrong")
	var marketProbability = market.Probability
	var marketMin float64 = market.Min
	var marketMax = market.Max

	// Sanity check
	var calculatedProbability = calcMarketProbabilityFromMarketP(market.P, market.Pool)
	assert(math.Abs(calculatedProbability-marketProbability) < 0.001, "Sanity check error: Differing probabilities (calculatedProbability) vs (marketProbability), diff (math.Abs(calculatedProbability - marketProbability)) is greater than 0.001. Formulas might have changed. Stopping to prevent messing up the formulas and betting wrong")

	// PSEUDO_NUMERIC value formulas based on market's cpmm probability, min and max. Formulas from https://github.com/manifoldmarkets/manifold/blob/main/common/src/pseudo-numeric.ts#L13
	if market.IsLogScale {
		return math.Pow(10, marketProbability*math.Log10(marketMax-marketMin+1)) + marketMin - 1
	} else {
		return marketProbability*(marketMax-marketMin) + marketMin
	}
}

func ConvertValueToProbability(market ManifoldApi.Market, value float64) float64 {
	var targetProbability float64
	// Manually inversed PSEUDO_NUMERIC value formulas from calculatePseudoNumericMarketplaceValue, to solve for market cpmm probability based on a target value
	if market.IsLogScale {
		targetProbability = math.Log(value-market.Min+1) / math.Log(10) / math.Log10(market.Max-market.Min+1)

		log.Printf(
			"First I got from wolframalpha, second I solved myself: %v, %v",
			math.Log(value-market.Min+1)/math.Log(market.Max-market.Min+1),
			math.Log(value-market.Min+1)/math.Log(10)/math.Log10(market.Max-market.Min+1),
		)

		assert(false, "The code should work for logScale with minor fixes but it's not tested so idk, just crash for now")
	} else {
		targetProbability = (value - market.Min) / (market.Max - market.Min)
	}
	return targetProbability
}

func CalculatePseudoNumericMarketplaceBet(market ManifoldApi.Market, targetValue float64) (string, int64) {
	assert(market.OutcomeType == "PSEUDO_NUMERIC", "Unexpected outcomeType (market.OutcomeType). Stopping to prevent messing up the formulas and betting wrong")
	assert(market.Mechanism == "cpmm-1", "Unexpected mechanism (market.OutcomeType). Stopping to prevent messing up the formulas and betting wrong")
	var marketProbability = market.Probability
	var targetProbability float64 = ConvertValueToProbability(market, targetValue)

	// Manually inversed formulas from calculatePseudoNumericMarketplaceValue, to solve for YES/NO votes on a target cpmm probability
	var outcome string
	if marketProbability < targetProbability {
		outcome = "YES"
	} else {
		outcome = "NO"
	}

	var amount = calculateCpmmAmountToProb(
		CpmmState{
			p:    calcMarketPFromMarketProbability(marketProbability, market.Pool),
			pool: market.Pool,
		},
		targetProbability,
		outcome,
	)

	return outcome, int64(math.Round(amount))
}

// Following functions copied from https://github.com/manifoldmarkets/manifold/blob/0a71fdd0e3d684145022dcb9f27bb8bb14835d50/common/src/calculate-cpmm.ts and translated to Swift and then to Go

const CREATOR_FEE float64 = 0
const PLATFORM_FEE float64 = 0
const LIQUIDITY_FEE float64 = 0

func calculateCpmmShares(
	pool pool,
	p float64,
	bet float64,
	outcome string,
) float64 {
	var k = math.Pow(pool.YES, p) * math.Pow(pool.NO, (1-p))

	if outcome == "YES" {
		return pool.YES + bet - math.Pow(k*math.Pow(bet+pool.NO, (p-1)), (1/p))
	} else {
		return pool.NO + bet - math.Pow(k*math.Pow(bet+pool.YES, -p), (1/(1-p)))
	}
}

func calculateCpmmAmountToProb(
	state CpmmState,
	prob float64,
	outcome string,
) float64 {
	if prob <= 0 || prob >= 1 {
		return math.Inf(1)
	}
	if outcome == "NO" {
		prob = 1 - prob
	}

	// First, find an upper bound that leads to a more extreme probability than prob.
	var maxGuess float64 = 10
	var maxProb float64 = 0
	for {
		maxGuess *= 10
		maxProb = getCpmmOutcomeProbabilityAfterBet(state, outcome, maxGuess)

		if maxProb >= prob {
			break
		}
	}

	// Then, binary search for the amount that gets closest to prob.
	var amount = binarySearch(0, maxGuess, func(amount float64) float64 {
		var newProb = getCpmmOutcomeProbabilityAfterBet(state, outcome, amount)
		return newProb - prob
	})

	return amount
}

func getCpmmOutcomeProbabilityAfterBet(
	state CpmmState,
	outcome string,
	bet float64,
) float64 {
	var newPool = calculateCpmmPurchase(state, bet, outcome).newPool
	var p = getCpmmProbability(newPool, state.p)
	if outcome == "YES" {
		return p
	} else {
		return 1 - p
	}
}

type PurchaseInfo struct {
	shares  float64
	newPool pool
	newP    float64
	fees    Fees
}

func calculateCpmmPurchase(
	state CpmmState,
	bet float64,
	outcome string,
) PurchaseInfo {
	var feesInfo = getCpmmFees(state, bet, outcome)

	var shares = calculateCpmmShares(state.pool, state.p, feesInfo.remainingBet, outcome)

	var newY float64
	var newN float64
	if outcome == "YES" {
		newY = state.pool.YES - shares + feesInfo.remainingBet + feesInfo.fees.liquidityFee
		newN = state.pool.NO + feesInfo.remainingBet + feesInfo.fees.liquidityFee
	} else {
		newY = state.pool.YES + feesInfo.remainingBet + feesInfo.fees.liquidityFee
		newN = state.pool.NO - shares + feesInfo.remainingBet + feesInfo.fees.liquidityFee
	}

	var postBet pool = pool{YES: newY, NO: newN}

	var addInfo = addCpmmLiquidity(postBet, state.p, feesInfo.fees.liquidityFee)

	return PurchaseInfo{
		shares:  shares,
		newPool: addInfo.newPool,
		newP:    addInfo.newP,
		fees:    feesInfo.fees,
	}
}

type AddCpmmLiquidityInfo struct {
	newPool   pool
	newP      float64
	liquidity float64
}

func addCpmmLiquidity(
	pool pool,
	p float64,
	amount float64,
) AddCpmmLiquidityInfo {
	var prob = getCpmmProbability(pool, p)

	// https://www.wolframalpha.com/input?i=p%28n%2Bb%29%2F%28%281-p%29%28y%2Bb%29%2Bp%28n%2Bb%29%29%3Dq%2C+solve+p
	var numerator = prob * (amount + pool.YES)
	var denominator = amount - pool.NO*(prob-1) + prob*pool.YES
	var newP = numerator / denominator

	var newPool = ManifoldApi.MarketPool{YES: pool.YES + amount, NO: pool.NO + amount}

	var oldLiquidity = getCpmmLiquidity(pool, newP)
	var newLiquidity = getCpmmLiquidity(newPool, newP)
	var liquidity = newLiquidity - oldLiquidity

	return AddCpmmLiquidityInfo{
		newPool:   newPool,
		liquidity: liquidity,
		newP:      newP,
	}
}

func getCpmmLiquidity(
	pool pool,
	p float64,
) float64 {
	return math.Pow(pool.YES, p) * math.Pow(pool.NO, (1-p))
}

type FeesInfo struct {
	remainingBet float64
	totalFees    float64
	fees         Fees
}

func getCpmmFees(state CpmmState, bet float64, outcome string) FeesInfo {
	var prob = getCpmmProbabilityAfterBetBeforeFees(state, outcome, bet)

	var betP float64
	if outcome == "YES" {
		betP = 1 - prob

	} else {
		betP = prob
	}

	var liquidityFee = LIQUIDITY_FEE * betP * bet
	var platformFee = PLATFORM_FEE * betP * bet
	var creatorFee = CREATOR_FEE * betP * bet
	var fees Fees = Fees{liquidityFee: liquidityFee, platformFee: platformFee, creatorFee: creatorFee}

	var totalFees = liquidityFee + platformFee + creatorFee
	var remainingBet = bet - totalFees

	return FeesInfo{remainingBet: remainingBet, totalFees: totalFees, fees: fees}
}

func getCpmmProbabilityAfterBetBeforeFees(
	state CpmmState,
	outcome string,
	bet float64,
) float64 {
	var shares = calculateCpmmShares(state.pool, state.p, bet, outcome)

	var newY float64
	var newN float64
	if outcome == "YES" {
		newY = state.pool.YES - shares + bet
		newN = state.pool.NO + bet
	} else {
		newY = state.pool.YES + bet
		newN = state.pool.NO - shares + bet
	}

	return getCpmmProbability(pool{YES: newY, NO: newN}, state.p)
}

func getCpmmProbability(
	pool pool,
	p float64,
) float64 {
	return (p * pool.NO) / ((1-p)*pool.YES + p*pool.NO)
}

type Fees struct {
	liquidityFee float64
	platformFee  float64
	creatorFee   float64
}
type pool = ManifoldApi.MarketPool
type CpmmState struct {
	p    float64
	pool pool
}

func binarySearch(
	min float64,
	max float64,
	comparator func(float64) float64,
) float64 {
	var mid float64 = 0
	for {
		mid = min + (max-min)/2

		// Break once we've reached max precision.
		if mid == min || mid == max {
			break
		}

		var comparison = comparator(mid)
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
