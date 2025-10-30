import { z } from "zod";
import { load } from "@std/dotenv";

const envRaw = await load({});

export const env = z.object({
  MANIFOLD_API_KEY: z.string().min(1, "MANIFOLD_API_KEY is required"),
  MANIFOLD_API_DEBUG: z.string().optional().default("false"),
}).parse(envRaw);
