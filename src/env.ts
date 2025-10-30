import { z } from "zod";

export const env = z.object({
  MANIFOLD_API_KEY: z.string().min(1, "MANIFOLD_API_KEY is required"),
  MANIFOLD_API_DEBUG: z.string().optional().default("false"),
}).parse(Deno.env.toObject());
