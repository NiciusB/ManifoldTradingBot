import { callManifoldApi } from "./callManifoldApi.ts";
import { z } from "zod";

const LiteMarketSchema = z.object({
  id: z.string(),

  creatorId: z.string(),
  creatorUsername: z.string(),
  creatorName: z.string(),
  creatorAvatarUrl: z.string().optional(),

  createdTime: z.number(),
  closeTime: z.number().optional(),
  question: z.string(),
  url: z.string(),

  outcomeType: z.string(),
  mechanism: z.string(),

  probability: z.number().optional(),
  pool: z.record(z.string(), z.number()).optional(),
  p: z.number().optional(),
  totalLiquidity: z.number().optional(),

  value: z.number().optional(),
  min: z.number().optional(),
  max: z.number().optional(),
  isLogScale: z.boolean().optional(),

  volume: z.number(),
  volume24Hours: z.number(),

  isResolved: z.boolean(),
  resolutionTime: z.number().optional(),
  resolution: z.string().optional(),
  resolutionProbability: z.number().optional(),
  uniqueBettorCount: z.number(),

  lastUpdatedTime: z.number().optional(),
  lastBetTime: z.number().optional(),

  token: z.enum(["MANA", "CASH"]).optional(),
  siblingContractId: z.string().optional(),
});

const AnswerSchema = z.object({
  createdTime: z.number(),
  avatarUrl: z.string().url().optional(), // algunos pueden no tener o cambiar formato
  id: z.string(),
  username: z.string().optional(),
  number: z.number().optional(),
  name: z.string().optional(),
  contractId: z.string(),
  text: z.string(),
  userId: z.string(),
  probability: z.number(),
  pool: z.record(z.string(), z.number()).optional(),
});

const FullMarketSchema = LiteMarketSchema.extend({
  answers: z.array(AnswerSchema).optional(),
  shouldAnswersSumToOne: z.boolean().optional(),
  addAnswersMode: z.enum(["ANYONE", "ONLY_CREATOR", "DISABLED"]).optional(),

  options: z
    .array(
      z.object({
        text: z.string(),
        votes: z.number(),
      }),
    )
    .optional(),

  totalBounty: z.number().optional(),
  bountyLeft: z.number().optional(),

  // description: z.any(), // not needed for now
  textDescription: z.string(),
  coverImageUrl: z.string().optional().nullable(),
  groupSlugs: z.array(z.string()).optional(),
});

export type FullMarket = z.infer<typeof FullMarketSchema>;

export async function getMarket(marketId: string): Promise<FullMarket> {
  const sb = await callManifoldApi("GET", `v0/market/${marketId}`);
  return FullMarketSchema.parse(JSON.parse(sb));
}
