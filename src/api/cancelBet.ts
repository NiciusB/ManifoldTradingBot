import z from "zod";
import { callManifoldApi } from "./callManifoldApi.ts";

const BetSchema = z.object({
  id: z.string(),
  isCancelled: z.boolean().nullable(),
});

export async function cancelBet(betId: string): Promise<void> {
  // The Go implementation POSTs to v0/bet/cancel/{betId} and parses a Bet response
  const sb = await callManifoldApi("POST", `v0/bet/cancel/${betId}`);
  const response = BetSchema.parse(JSON.parse(sb));

  if (!response.id || !response.isCancelled) {
    throw new Error(`failed to cancel bet: ${JSON.stringify(sb)}`);
  }
}
