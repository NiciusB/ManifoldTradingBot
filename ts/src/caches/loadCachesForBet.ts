import { WsBet } from "../api/ws/parseManifoldWsNewBetEvent.ts";
import { marketsCache } from "./marketsCache.ts";
import { myMarketPositionCache } from "./myMarketPositionCache.ts";
import { usersCache } from "./usersCache.ts";

export async function loadCachesForBet(bet: WsBet) {
  const results = await Promise.all(
    [
      marketsCache.get(bet.contractId),
      usersCache.get(bet.userId),
      myMarketPositionCache.get(bet.contractId),
    ] as const,
  );

  return {
    market: results[0],
    user: results[1],
    myPosition: results[2],
  };
}
