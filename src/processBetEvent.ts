import {
  PlaceBetRequest,
  placeInstantlyCancelledLimitOrder,
} from "./api/placeBet.ts";
import { WsBet } from "./api/ws/parseManifoldWsNewBetEvent.ts";
import { loadCachesForBet } from "./caches/loadCachesForBet.ts";
import { myMarketPositionCache } from "./caches/myMarketPositionCache.ts";
import { isBetGoodForVelocity } from "./isBetGoodForVelocity.ts";
import { mapNumber, oddsToProb, probToOdds } from "./utils/math.ts";

export async function processBetEvent(bet: WsBet): Promise<void> {
  if (bet.probBefore === undefined || bet.probAfter === undefined) {
    throw new Error(
      "Bet event must have probBefore and probAfter for our logic",
    );
  }

  const caches = await loadCachesForBet(bet);

  // This used to be a better metric: https://github.com/NiciusB/ManifoldTradingBot/blob/b4051ea662d5ea9fbdd2d9de23ac3d6bdfc7d9e8/go/ModuleVelocity/caches.go#L74
  const marketVelocity = caches.market.volume24Hours /
    (caches.market.volume + 1);

  // [0, 1] The bigger, the more we correct
  const alpha = 0.2 +
    mapNumber(caches.user.skillEstimate, 1, 0, 0, 0.5) +
    mapNumber(marketVelocity, 0, 1, -0.2, 0.3);

  const beforeOdds = probToOdds(bet.probBefore);
  const afterOdds = probToOdds(bet.probAfter);
  const correctedOdds = beforeOdds * alpha + afterOdds * (1 - alpha);
  const correctedProb = oddsToProb(correctedOdds);
  const limitProb = Math.round(correctedProb * 100) / 100; // round for manifold's limit order accuracy

  const outcome = bet.probBefore > bet.probAfter ? "YES" : "NO";

  // [10, 50] How much mana to bet
  const amount = Math.round(
    10 +
      mapNumber(caches.user.skillEstimate, 1, 0, 0, 15) +
      mapNumber(marketVelocity, 0, 1, 0, 25),
  );

  const betRequest: PlaceBetRequest = {
    contractId: bet.contractId,
    answerId: bet.answerId,
    outcome,
    amount,
    limitProb,
  };

  const isGoodOutcome = isBetGoodForVelocity(bet, caches, betRequest);
  if (isGoodOutcome !== "good") {
    // Bet is no good for velocity, ignore
    return;
  }

  await placeInstantlyCancelledLimitOrder(betRequest);
  await myMarketPositionCache.renew(bet.contractId);
}
