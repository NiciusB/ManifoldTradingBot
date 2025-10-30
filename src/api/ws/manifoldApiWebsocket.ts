export interface ManifoldSocketEvent {
  type: string;
  topic: string;
  data: Record<string, unknown>;
}

type ManifoldEventCallback = (event: ManifoldSocketEvent) => void;

const url = "wss://api.manifold.markets/ws";
let websocketConn: WebSocket;
let txid = 0;
const callbacks: ManifoldEventCallback[] = [];

export function addManifoldWebsocketEventListener(
  callback: ManifoldEventCallback,
): void {
  callbacks.push(callback);
}

export function sendManifoldApiWebsocketMessage(msg: string): void {
  websocketConn.send(msg);
}

export function connectManifoldApiWebsocket(): void {
  websocketConn = new WebSocket(url);

  websocketConn.onopen = () => {
    console.debug("Manifold api websocket connected");
    sendConnectionMessage();
    startHeartbeatMessages();
  };

  websocketConn.onmessage = (event) => {
    const message = JSON.parse(event.data) as ManifoldSocketEvent;

    // Execute callbacks asynchronously
    for (const callback of callbacks) {
      queueMicrotask(() => callback(message));
    }
  };

  websocketConn.onerror = (error) => {
    console.error("Manifold api websocket connection error:", error);
    Deno.exit(1);
  };

  websocketConn.onclose = () => {
    console.error("Manifold api websocket connection closed unexpectedly");
    Deno.exit(1);
  };
}

function sendConnectionMessage(): void {
  const msg = JSON.stringify({
    type: "subscribe",
    txid: txid++,
    topics: ["global/new-bet"],
  });

  try {
    sendManifoldApiWebsocketMessage(msg);
  } catch (err) {
    console.error("sendConnectionMessage error:", err);
  }
}

function startHeartbeatMessages(): void {
  setInterval(() => {
    const msg = JSON.stringify({
      type: "ping",
      txid: txid++,
    });

    try {
      sendManifoldApiWebsocketMessage(msg);
    } catch (err) {
      console.error("sendHeartbeatMessages error:", err);
    }
  }, 50000); // 50 seconds
}
