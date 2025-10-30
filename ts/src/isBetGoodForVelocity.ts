import { PlaceBetRequest } from "./api/placeBet.ts";
import { CachedMarket } from "./caches/marketsCache.ts";
import { CachedMarketPosition } from "./caches/myMarketPositionCache.ts";
import { CachedUser } from "./caches/usersCache.ts";

export type IsGoodForVelocityOutcome =
  | "good"
  | "bad"
  | "ignored-bot"
  | "ignored-redemption"
  | "ignored-unfilled-limit-order"
  | "ignored-scary-user"
  | "ignored-market-creator"
  | "ignored-low-volume"
  | "ignored-old-bet"
  | "ignored-limit-order-fill"
  | "ignored-extreme-values-decent-user"
  | "ignored-extreme-values-good-user"
  | "ignored-insufficient-granularity"
  | "ignored-small-prob-change"
  | "ignored-too-invested";

export function isBetGoodForVelocity(
  bet: {
    probBefore?: number;
    probAfter?: number;
    userId: string;
    limitProb?: number;
    amount: number;
    isRedemption?: boolean;
    isApi?: boolean;
    createdTime: number;
    fills?: {
      timestamp: number;
    }[];
  },
  caches: {
    market: CachedMarket;
    user: CachedUser;
    myPosition: CachedMarketPosition;
  },
  betRequest: PlaceBetRequest,
): IsGoodForVelocityOutcome {
  if (bet.probBefore === undefined || bet.probAfter === undefined) {
    throw new Error(
      "Bet event must have probBefore and probAfter for our logic",
    );
  }
  if (betRequest.limitProb === undefined) {
    throw new Error(
      "Bet request must have limitProb for our logic",
    );
  }

  // Ignore redemptions, as they are always the opposite of another bet
  if (bet.isRedemption) return "ignored-redemption";

  // Ignore bots, mainly to prevent infinite loops of one reacting to another
  if (bet.isApi) return "ignored-bot";

  // Ignore unfilled limit orders
  if (bet.amount === 0 || bet.probBefore === bet.probAfter) {
    return "ignored-unfilled-limit-order";
  }

  // Ignore bets by market creator
  if (caches.market.creatorId === bet.userId) {
    return "ignored-market-creator";
  }

  // Ignore old bets. Sometimes supabase gives old bets for some reason
  if (bet.createdTime < Date.now() - 5000) return "ignored-old-bet";

  // Ignore markets with low volume
  if (caches.market.volume < 5000 || caches.market.volume24Hours < 100) {
    return "ignored-low-volume";
  }

  // Ignore banned user ids
  if (SCARY_USER_IDS.includes(bet.userId)) return "ignored-scary-user";

  // Ignore fills of limit orders, other than the initial one
  for (const fill of bet.fills ?? []) {
    if (fill.timestamp !== bet.createdTime) {
      return "ignored-limit-order-fill";
    }
  }

  // Ignore extreme values for decent users
  if (
    caches.user.profitCachedAllTime > 1000 &&
    (betRequest.limitProb >= 0.99 || betRequest.limitProb <= 0.01)
  ) {
    return "ignored-extreme-values-decent-user";
  }

  // Ignore extreme values for good users
  if (
    caches.user.profitCachedAllTime > 5000 &&
    (betRequest.limitProb >= 0.97 || betRequest.limitProb <= 0.03)
  ) {
    return "ignored-extreme-values-good-user";
  }

  // Ignore probabilities without enough granularity
  const smallestProb = Math.min(bet.probBefore, bet.probAfter);
  const largestProb = Math.max(bet.probBefore, bet.probAfter);
  if (
    betRequest.limitProb <= smallestProb ||
    betRequest.limitProb >= largestProb
  ) {
    // We do not have enough granularity on the limit order probabilities: We would bounce even more than the original probs
    // This only works for betting against the latest best, if we wanted to sometimes follow it, we would need to rework this check
    return "ignored-insufficient-granularity";
  }

  const betLogOdds = Math.log(bet.probAfter / (1 - bet.probAfter));
  const betRequestLogOdds = Math.log(
    betRequest.limitProb / (1 - betRequest.limitProb),
  );
  const logOddsDiff = Math.abs(betRequestLogOdds - betLogOdds);
  const marketVolumeFactor = Math.min(caches.market.totalShares, 50000) / 50000;
  const minLogOddsSwing = 0.4 - marketVolumeFactor * 0.39;
  if (logOddsDiff < minLogOddsSwing) {
    // Ignore small prob changes
    return "ignored-small-prob-change";
  }

  // Ignore markets where I am too invested on one side. This could be increased in the future to allow larger positions
  const myPosition = caches.myPosition.positions.find((x) =>
    x.answerId === betRequest.answerId
  );
  if (
    myPosition && myPosition.invested > 60 &&
    ((betRequest.outcome == "YES" && myPosition.hasYesShares) ||
      (betRequest.outcome == "NO" && !myPosition.hasYesShares))
  ) {
    return "ignored-too-invested";
  }

  return "good";
}

const SCARY_USER_IDS = [
  "w1knZ6yBvEhRThYPEYTwlmGv7N33",
  "BhNkw088bMNwIFF2Aq5Gg9NTPzz1",
  "dNgcgrHGn8ZB30hyDARNtbjvGPm1",
  "kzjQhRJ4GINn5umiq2ee1QvaMcE2",
  "P7PV13rynzOHyxm8AiXIN568bmF2",
  "jOl1FMKpFbXkoaDGp2qlakUxAiJ3",
  "MxdyEeVgrFMTDDsPbXwAe9W1CLs2",
  "IEVDP2LTpgMYaka38r1TVZcabWS2",
  "prSlKwvKkRfHCY43txO4pG1sFMT2",
  "XebdFvo6vqO5WGXTsWYVdSH3WNc2",
  "Y8xXwCCYe3cBCW5XeU8MxykuPAY2",
  "ymezf2YMJ9aaILxT95uWJj7gnx83",
  "Y96HJoD5tQaPgbKi5JEt5JuQJLN2",
  "ffwIBb255DhSsJRh3VWZ4RY2pxz2",
  "wjbOTRRJ7Ee5mjSMMYrtwoWuiCp2",
  "EFzCw6YhqTYCJpeWHUG6p9JsDy02",
  "UN5UGCJRQdfB3eQSnadiAxjkmRp2",
  "9B5QsPTDAAcWOBW8NJNS7YdUjpO2",
  "KIpsyUwgKmO1YXv2EJPXwGxaO533",
  "VI8Htwx9JYeKeT6cUnH66XvBAv73",
  "n820DjHGX9dKsrv0jHIJV8xmDgr2",
  "w07LrYnLg8XDHySwrKxmAYAnLJH2",
  "U7KQfJgJp1fa35k9EXpQCgvmmjh1",
  "rVaQiGT7qCRfAD9QDQQ8SHxvvuu2",
  "wuOtYy52f4Sx4JFfT85LpizVGsx1",
  "I8VZW5hGw9cfIeWs7oQJaNdFwhL2",
  "kydVkcfg7TU4zrrMBRx1Csipwkw2",
  "QQodxPUTIFdQWJiIzVUW2ztF43e2",
  "K2BeNvRj4beTBafzKLRCnxjgRlv1",
  "zgCIqq8AmRUYVu6AdQ9vVEJN8On1",
  "BB5ZIBNqNKddjaZQUnqkFCiDyTs2",
];
