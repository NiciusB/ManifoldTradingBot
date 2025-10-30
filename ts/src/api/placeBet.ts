import z from "zod";
import { callManifoldApi } from "./callManifoldApi.ts";
import { cancelBet } from "./cancelBet.ts";

const PlaceBetResponseSchema = z.object({
  message: z.string().optional(),
  isFilled: z.boolean().optional(),
  betId: z.string().optional(),
});
export type PlaceBetResponse = z.infer<typeof PlaceBetResponseSchema>;

export type PlaceBetRequest = {
  contractId: string;
  outcome: "YES" | "NO";
  amount: number;
  answerId?: string | undefined;
  limitProb?: number | undefined;
};

export async function placeBet(
  bet: PlaceBetRequest,
) {
  const sb = await callManifoldApi("POST", "v0/bet", bet);
  const response = PlaceBetResponseSchema.parse(JSON.parse(sb));

  if (response.message) {
    throw new Error(response.message);
  }

  return response;
}

export async function placeInstantlyCancelledLimitOrder(
  betRequest: PlaceBetRequest,
) {
  const placedBet = await placeBet(betRequest);

  if (!placedBet.isFilled) {
    // Cancel instantly if it didn't fully fill
    if (placedBet.betId) {
      await cancelBet(placedBet.betId);
    }
  }
}
