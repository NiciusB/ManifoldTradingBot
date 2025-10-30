import { assertEquals } from "@std/assert";
import { isBetGoodForVelocity } from "./isBetGoodForVelocity.ts";

Deno.test("isBetGoodForVelocity - should reject missing probabilities", () => {
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: Date.now(),
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.5,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  try {
    isBetGoodForVelocity(bet, caches, betRequest);
    throw new Error("Expected function to throw");
  } catch (error) {
    assertEquals(
      (error as Error).message,
      "Bet event must have probBefore and probAfter for our logic",
    );
  }
});

Deno.test("isBetGoodForVelocity - should ignore redemptions", () => {
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.6,
    isRedemption: true,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.5,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-redemption");
});

Deno.test("isBetGoodForVelocity - should ignore bots", () => {
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.6,
    isApi: true,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.5,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-bot");
});

Deno.test("isBetGoodForVelocity - should ignore unfilled limit orders", () => {
  const bet = {
    userId: "test-user",
    amount: 0, // Zero amount
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.5,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.5,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-unfilled-limit-order");
});

Deno.test("isBetGoodForVelocity - should ignore market creator", () => {
  const creatorId = "creator-user";
  const bet = {
    userId: creatorId,
    amount: 100,
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.6,
  };
  const caches = {
    market: {
      creatorId,
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.5,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-market-creator");
});

Deno.test("isBetGoodForVelocity - should ignore old bets", () => {
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: Date.now() - 6000, // 6 seconds ago
    probBefore: 0.5,
    probAfter: 0.6,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.5,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-old-bet");
});

Deno.test("isBetGoodForVelocity - should ignore low volume markets", () => {
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.6,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 4000, // Low volume
      volume24Hours: 50, // Low 24h volume
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.5,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-low-volume");
});

Deno.test("isBetGoodForVelocity - should ignore scary users", () => {
  const bet = {
    userId: "w1knZ6yBvEhRThYPEYTwlmGv7N33", // First scary user from the list
    amount: 100,
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.6,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.5,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-scary-user");
});

Deno.test("isBetGoodForVelocity - should ignore limit order fills", () => {
  const currentTime = Date.now();
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: currentTime,
    probBefore: 0.5,
    probAfter: 0.6,
    fills: [
      { timestamp: currentTime }, // Original fill
      { timestamp: currentTime + 1000 }, // Later fill
    ],
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.5,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-limit-order-fill");
});

Deno.test("isBetGoodForVelocity - should ignore extreme values for decent users", () => {
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.6,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 2000, // Decent user
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.995, // Extreme value
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-extreme-values-decent-user");
});

Deno.test("isBetGoodForVelocity - should ignore extreme values for good users", () => {
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.6,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 6000, // Good user
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.98, // Extreme value
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-extreme-values-good-user");
});

Deno.test("isBetGoodForVelocity - should ignore insufficient granularity", () => {
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: Date.now(),
    probBefore: 0.4,
    probAfter: 0.6,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.3, // Below smallest prob
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-insufficient-granularity");
});

Deno.test("isBetGoodForVelocity - should ignore small probability changes", () => {
  const bet = {
    userId: "test-user",
    amount: 10,
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.52,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.51,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-small-prob-change");
});

Deno.test("isBetGoodForVelocity - should ignore too invested positions", () => {
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.7,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [{
        answerId: "1",
        invested: 70,
        hasYesShares: true,
      }],
    },
  };
  const betRequest = {
    limitProb: 0.6,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "ignored-too-invested");
});

Deno.test("isBetGoodForVelocity - should accept good bets", () => {
  const bet = {
    userId: "test-user",
    amount: 100,
    createdTime: Date.now(),
    probBefore: 0.5,
    probAfter: 0.7,
  };
  const caches = {
    market: {
      creatorId: "other-user",
      volume: 10000,
      volume24Hours: 1000,
      totalShares: 20000,
      uniqueBettorCount: 10,
    },
    user: {
      profitCachedAllTime: 0,
      skillEstimate: 0,
    },
    myPosition: {
      positions: [],
    },
  };
  const betRequest = {
    limitProb: 0.6,
    answerId: "1",
    outcome: "YES" as const,
    contractId: "test-contract",
    amount: 1,
  };

  const result = isBetGoodForVelocity(
    bet,
    caches,
    betRequest,
  );
  assertEquals(result, "good");
});
