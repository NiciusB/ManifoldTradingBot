import { z } from "zod";

export const BetFeesSchema = z.object({
  creatorFee: z.number(),
  platformFee: z.number(),
  liquidityFee: z.number(),
});

export const FillSchema = z.object({
  matchedBetId: z.string().nullable(),
  shares: z.number(),
  amount: z.number(),
  timestamp: z.number(),
  fees: BetFeesSchema,
});

export const BetSchema = z.object({
  id: z.string(),
  userId: z.string(),
  contractId: z.string(),
  outcome: z.string(),

  amount: z.number(),
  shares: z.number(),
  loanAmount: z.number(),

  // opcionales
  orderAmount: z.number().optional(),
  isApi: z.boolean().optional(),
  silent: z.boolean().optional(),
  betGroupId: z.string().optional(),
  answerId: z.string().optional(),
  limitProb: z.number().optional(),
  expiresAt: z.number().optional(),

  isFilled: z.boolean().optional(),
  isCancelled: z.boolean().optional(),
  isRedemption: z.boolean().optional(),

  probBefore: z.number().optional(),
  probAfter: z.number().optional(),

  createdTime: z.number(),
  visibility: z.string().optional(),

  fees: BetFeesSchema,
  fills: z.array(FillSchema).optional(),
});

export const ReceivedBetSchema = z.object({
  bets: z.array(BetSchema),
});

// Tipo inferido (por si lo necesitas)
export type WsNetBetEvent = z.infer<typeof ReceivedBetSchema>;
export type WsBet = z.infer<typeof BetSchema>;

export function parseManifoldWsNewBetEvent(data: unknown): WsNetBetEvent {
  return ReceivedBetSchema.parse(data);
}
