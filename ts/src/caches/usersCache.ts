import { getUser } from "../api/getUser.ts";
import { getUserPortfolioHistory } from "../api/getUserPortfolioHistory.ts";
import { mapNumber } from "../utils/math.ts";
import { Cache } from "./cache.ts";

export interface CachedUser {
  profitCachedAllTime: number;
  skillEstimate: number; // [0-1], our own formula that estimates skill
}

// Users cache - 5 days TTL, 30 minutes minimum refresh
export const usersCache = new Cache<CachedUser>(
  "usersCacheV1",
  async (userId: string): Promise<CachedUser> => {
    const [apiUser, monthlyPortfolioHistory] = await Promise.all(
      [
        getUser(userId),
        getUserPortfolioHistory(userId, "monthly"),
      ] as const,
    );

    const profitAtStartOfMonth = monthlyPortfolioHistory.length > 0
      ? monthlyPortfolioHistory[0].profit ?? 0
      : 0;
    const profitAtEndOfMonth = monthlyPortfolioHistory.length > 0
      ? monthlyPortfolioHistory[monthlyPortfolioHistory.length - 1].profit ?? 0
      : 0;

    const allTimeProfit = profitAtEndOfMonth;
    const monthlyProfit = profitAtEndOfMonth - profitAtStartOfMonth;

    // [0-1]
    const skillEstimate = 0.5 +
      mapNumber(allTimeProfit, -5000, 40000, -0.1, 0.3) +
      mapNumber(monthlyProfit ?? 0, -2000, 10000, -0.1, 0.2) +
      mapNumber(
        Date.now() - apiUser.createdTime,
        0,
        24 * 30 * 60 * 60 * 1000,
        -0.2,
        0,
      );

    return {
      profitCachedAllTime: allTimeProfit,
      skillEstimate,
    };
  },
  5 * 24 * 60 * 60 * 1000, // 5 days in ms
  30 * 60 * 1000,
);
