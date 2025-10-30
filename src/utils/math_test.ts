import { mapNumber } from "./math.ts";
import { assertAlmostEquals } from "@std/assert/almost-equals";

Deno.test("Calculates correct numbers for mapNumber", () => {
  assertAlmostEquals(mapNumber(70, 0, 100, 0, 1), 0.7);
  assertAlmostEquals(mapNumber(70, 100, 0, 0, 1), 0.3);
  assertAlmostEquals(mapNumber(70, 0, 100, 1, 0), 0.3);
  assertAlmostEquals(mapNumber(70, 100, 0, 1, 0), 0.7);
});
