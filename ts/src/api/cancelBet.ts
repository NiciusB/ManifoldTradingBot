import z from "zod";
import { callManifoldApi } from "./callManifoldApi.ts";

const BetSchema = z.object({
  id: z.string(),
  amount: z.number().optional(),
  isAnte: z.boolean().optional(),
  shares: z.number().optional(),
  userId: z.string().optional(),
  outcome: z.string().optional(),
  answerId: z.string().optional(),
  probAfter: z.number().optional(),
  contractId: z.string().optional(),
  loanAmount: z.number().optional(),
  probBefore: z.number().optional(),
  visibility: z.string().optional(),
  createdTime: z.number().optional(),
  isChallenge: z.boolean().optional(),
  isRedemption: z.boolean().optional(),
  isApi: z.boolean().nullable(),
  isFilled: z.union([z.boolean(), z.null()]).optional(),
  userName: z.string().optional(),
  isCancelled: z.boolean().nullable(),
  orderAmount: z.number().optional(),
  userUsername: z.string().optional(),
  userAvatarUrl: z.string().optional(),
  limitProb: z.number().optional(),
});

export async function cancelBet(betId: string): Promise<void> {
  // The Go implementation POSTs to v0/bet/cancel/{betId} and parses a Bet response
  const sb = await callManifoldApi("POST", `v0/bet/cancel/${betId}`);
  const response = BetSchema.parse(JSON.parse(sb));

  if (!response.id || !response.isCancelled) {
    throw new Error(`failed to cancel bet: ${JSON.stringify(response)}`);
  }
}
