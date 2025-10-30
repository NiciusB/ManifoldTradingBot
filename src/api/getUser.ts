import z from "zod";
import { callManifoldApi } from "./callManifoldApi.ts";

export const UserSchema = z.object({
  id: z.string(),
  createdTime: z.number(),
  name: z.string().optional(),
  username: z.string().optional(),
  balance: z.number().optional(),
  totalDeposits: z.number().optional(),
  isBot: z.boolean().optional(),
  isAdmin: z.boolean().optional(),
  isTrustworthy: z.boolean().optional(),
  currentBettingStreak: z.number().optional(),
  lastBetTime: z.number().optional(),
});

export type User = z.infer<typeof UserSchema>;

export async function getUser(userId: string): Promise<User> {
  const sb = await callManifoldApi("GET", `v0/user/by-id/${userId}`);
  return UserSchema.parse(JSON.parse(sb));
}
