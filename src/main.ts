import { startConsumingApiCallsQueue } from "./api/callManifoldApi.ts";
import {
  addManifoldWebsocketEventListener,
  connectManifoldApiWebsocket,
} from "./api/ws/manifoldApiWebsocket.ts";
import { parseManifoldWsNewBetEvent } from "./api/ws/parseManifoldWsNewBetEvent.ts";
import { loadMe } from "./me.ts";
import { processBetEvent } from "./processBetEvent.ts";

async function main() {
  startConsumingApiCallsQueue();

  addManifoldWebsocketEventListener((event) => {
    if (event.topic === "global/new-bet") {
      const parsedEvent = parseManifoldWsNewBetEvent(event.data);
      parsedEvent.bets.forEach((bet) => {
        processBetEvent(bet);
      });
    }
  });
  await loadMe();
  connectManifoldApiWebsocket();
}

main();
