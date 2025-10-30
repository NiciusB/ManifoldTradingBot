import { getMarketPositionsForUser } from "../api/getMarketPositions.ts";
import { getMyUserId } from "../me.ts";
import { Cache } from "./cache.ts";

export interface CachedMarketPosition {
  positions: {
    answerId?: string;
    hasYesShares: boolean;
    invested: number;
  }[];
}

// My market position cache - 5 days TTL, 1 hour minimum refresh
export const myMarketPositionCache = new Cache<CachedMarketPosition>(
  "myMarketPositionCacheV1",
  async (marketId: string): Promise<CachedMarketPosition> => {
    const apiMarketPositions = await getMarketPositionsForUser(
      marketId,
      getMyUserId(),
    );
    return {
      positions: apiMarketPositions.map((x) => ({
        answerId: x.answerId ?? undefined,
        hasYesShares: x.hasYesShares ?? false,
        invested: x.invested ?? 0,
      })),
    };
  },
  5 * 24 * 60 * 60 * 1000, // 5 days in ms
  60 * 60 * 1000,
);
