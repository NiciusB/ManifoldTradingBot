import { callManifoldApi } from "./callManifoldApi.ts";
import { z } from "zod";

export const ContractMetricSchema = z.object({
  contractId: z.string(),
  answerId: z.string().nullable(),

  from: z
    .record(
      z.string(),
      z.object({
        profit: z.number(),
        profitPercent: z.number(),
        invested: z.number(),
        prevValue: z.number(),
        value: z.number(),
      }),
    )
    .optional(),

  hasNoShares: z.boolean(),
  hasShares: z.boolean(),
  hasYesShares: z.boolean(),

  invested: z.number(),
  loan: z.number(),

  maxSharesOutcome: z.string().nullable(),

  payout: z.number(),
  profit: z.number(),
  profitPercent: z.number(),

  totalShares: z.record(z.string(), z.number()),

  userId: z.string(),
  userUsername: z.string().optional(),
  userName: z.string().optional(),
  userAvatarUrl: z.string().optional(),

  lastBetTime: z.number(),
});

export type MarketPosition = z.infer<typeof ContractMetricSchema>;

export async function getMarketPositionsForUser(
  marketId: string,
  userId: string,
): Promise<MarketPosition[]> {
  const sb = await callManifoldApi(
    "GET",
    `v0/market/${marketId}/positions`,
    {
      userId,
    },
  );
  return z.array(ContractMetricSchema).parse(JSON.parse(sb));
}
