import { env } from "../env.ts";

interface ApiRequest {
  method: string;
  url: string;
  body?: object;
  responseCallback: (response: string) => void;
}

const apiReqQueue: ApiRequest[] = [];
const reqPerSecond = 8; // Limit is 500 per minute (8.33 per second), use 8 to be safe
export function startConsumingApiCallsQueue() {
  setInterval(() => {
    if (apiReqQueue.length > 0) {
      const apiReq = apiReqQueue.pop();
      if (apiReq) {
        void consumeQueueApiReq(apiReq);
      }
    }
  }, 1000 / reqPerSecond);
}

export function callManifoldApi(
  method: string,
  path: string,
  bodyOrParams?: Record<string, string | number | undefined>,
): Promise<string> {
  return new Promise((resolve) => {
    if (method === "GET" && bodyOrParams) {
      const params = new URLSearchParams(
        Object.fromEntries(
          Object.entries(bodyOrParams).filter(([_, v]) => v !== undefined),
        ) as Record<string, string>,
      ).toString();
      path += `?${params}`;
      bodyOrParams = undefined;
    }

    apiReqQueue.push({
      method,
      url: "https://api.manifold.markets/" + path,
      body: bodyOrParams,
      responseCallback: resolve,
    });
  });
}

async function consumeQueueApiReq(apiReq: ApiRequest): Promise<void> {
  const headers = new Headers({
    "User-Agent": "ManifoldTradingBot/2.0.0 for @NiciusBot",
    "Authorization": `Key ${env.MANIFOLD_API_KEY}`,
  });

  if (apiReq.method === "POST" && apiReq.body) {
    headers.set("Content-Type", "application/json");
  }

  try {
    const response = await fetch(apiReq.url, {
      method: apiReq.method,
      headers,
      body: apiReq.body ? JSON.stringify(apiReq.body) : undefined,
    });

    const text = await response.text();
    apiReq.responseCallback(text);
  } catch (err) {
    console.error("Error in API request:", err);
    throw err;
  }
}

export function getQueueLength(): number {
  return apiReqQueue.length;
}
