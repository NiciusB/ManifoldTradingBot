import z from "zod";
import { callManifoldApi } from "./callManifoldApi.ts";

const PortfolioMetricsSchema = z.object({
  investmentValue: z.number(),
  cashInvestmentValue: z.number(),
  balance: z.number(),
  cashBalance: z.number(),
  spiceBalance: z.number(),
  totalDeposits: z.number(),
  totalCashDeposits: z.number(),
  loanTotal: z.number(),
  timestamp: z.number(),
  profit: z.number().optional(),
  userId: z.string(),
});

export type PortfolioMetrics = z.infer<typeof PortfolioMetricsSchema>;

export async function getUserPortfolioHistory(
  userId: string,
  period: "daily" | "weekly" | "monthly" | "allTime",
): Promise<PortfolioMetrics[]> {
  const res = await callManifoldApi("GET", `v0/get-user-portfolio-history`, {
    userId,
    period,
  });
  return z.array(PortfolioMetricsSchema).parse(JSON.parse(res));
}
