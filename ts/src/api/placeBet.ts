import z from "zod";
import { callManifoldApi } from "./callManifoldApi.ts";
import { cancelBet } from "./cancelBet.ts";

const BetFillSchema = z.object({
  matchedBetId: z.string().nullable(),
  amount: z.number(),
  shares: z.number(),
  timestamp: z.number(),
});

const FeesSchema = z.object({
  creatorFee: z.number().optional(),
  platformFee: z.number().optional(),
  liquidityFee: z.number().optional(),
});

const PlaceBetRequestSchema = z.object({
  contractId: z.string(),
  answerId: z.string().optional(),
  outcome: z.enum(["YES", "NO"]),
  amount: z.number(),
  limitProb: z.number().optional(),
  expiresAt: z.string().optional(),
});

const PlaceBetResponseSchema = z.object({
  message: z.string().optional(),
  orderAmount: z.number().optional(),
  amount: z.number().optional(),
  shares: z.number().optional(),
  isFilled: z.boolean().optional(),
  isCancelled: z.boolean().optional(),
  fills: z.array(BetFillSchema).optional(),
  contractId: z.string().optional(),
  outcome: z.string().optional(),
  probBefore: z.number().optional(),
  probAfter: z.number().optional(),
  loanAmount: z.number().optional(),
  createdTime: z.number().optional(),
  fees: FeesSchema.optional(),
  isAnte: z.boolean().optional(),
  isRedemption: z.boolean().optional(),
  isChallenge: z.boolean().optional(),
  visibility: z.string().optional(),
  betId: z.string().optional(),
});
export type PlaceBetRequest = z.infer<typeof PlaceBetRequestSchema>;
export type PlaceBetResponse = z.infer<typeof PlaceBetResponseSchema>;

export async function placeBet(
  bet: PlaceBetRequest,
): Promise<PlaceBetResponse> {
  // Validate request
  const validatedBet = PlaceBetRequestSchema.parse(bet);
  // callManifoldApi queues the request and will JSON.stringify the body when sending
  const sb = await callManifoldApi("POST", "v0/bet", validatedBet);
  const response = PlaceBetResponseSchema.parse(JSON.parse(sb));

  if (response.message) {
    throw new Error(response.message);
  }

  return response;
}

export async function placeInstantlyCancelledLimitOrder(
  betRequest: PlaceBetRequest,
): Promise<PlaceBetResponse> {
  const placedBet = await placeBet(betRequest);

  if (!placedBet.isFilled) {
    // Cancel instantly if it didn't fully fill
    if (placedBet.betId) {
      await cancelBet(placedBet.betId);
    }
  }

  return placedBet;
}
