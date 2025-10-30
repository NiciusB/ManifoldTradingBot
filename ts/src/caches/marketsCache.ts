import { getMarket } from "../api/getMarket.ts";
import { Cache } from "./cache.ts";

export interface CachedMarket {
  creatorId: string;
  volume: number;
  volume24Hours: number;
  totalShares: number;
}

// Markets cache - 1 day TTL, 15 min minimum refresh
export const marketsCache = new Cache<CachedMarket>(
  "marketsCacheV1",
  async (marketId: string): Promise<CachedMarket> => {
    const apiMarket = await getMarket(marketId);

    let totalShares = 0;
    if (apiMarket.pool) {
      totalShares += Object.values(apiMarket.pool).reduce(
        (a, b) => a + b,
        0,
      );
    }
    if (apiMarket.answers) {
      for (const answer of apiMarket.answers) {
        if (answer.pool) {
          totalShares += Object.values(answer.pool).reduce(
            (a, b) => a + b,
            0,
          );
        }
      }
    }

    return {
      creatorId: apiMarket.creatorId,
      volume: apiMarket.volume,
      volume24Hours: apiMarket.volume24Hours,
      totalShares,
    };
  },
  24 * 60 * 60 * 1000,
  15 * 60 * 1000,
);
