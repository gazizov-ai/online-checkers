import type { GameStatePayload, Position } from "../types/domain";

export type SocketStatus = "connecting" | "open" | "closed";

type GameSocketMessage =
  | {
      type: "game_state";
      payload: GameStatePayload;
    }
  | {
      type: "game_finished";
      payload: {
        game_id: string;
        winner_id?: string;
        result?: "white_win" | "black_win" | "draw";
        finish_reason?: "checkers_rules" | "resignation" | "draw_agreement";
      };
    }
  | {
      type: "draw_offered";
      payload: {
        game_id: string;
        offered_by: string;
      };
    }
  | {
      type: "draw_declined";
      payload: {
        game_id: string;
        declined_by: string;
      };
    }
  | {
      type: "error";
      message: string;
    };

export type GameSocketHandlers = {
  onState: (state: GameStatePayload) => void;
  onFinished: (payload: {
    game_id: string;
    winner_id?: string;
    result?: "white_win" | "black_win" | "draw";
    finish_reason?: "checkers_rules" | "resignation" | "draw_agreement";
  }) => void;
  onDrawOffered: (payload: { game_id: string; offered_by: string }) => void;
  onDrawDeclined: (payload: { game_id: string; declined_by: string }) => void;
  onError: (message: string) => void;
  onStatusChange: (status: SocketStatus) => void;
};

function buildGameSocketURL(gameId: string): string {
  const configuredBase = import.meta.env.VITE_GAME_WS_BASE;

  if (configuredBase) {
    return `${configuredBase.replace(/\/$/, "")}/api/v1/games/${gameId}/ws`;
  }

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/game/api/v1/games/${gameId}/ws`;
}

export function connectGameSocket(gameId: string, handlers: GameSocketHandlers): WebSocket {
  handlers.onStatusChange("connecting");

  const socket = new WebSocket(buildGameSocketURL(gameId));

  socket.addEventListener("open", () => handlers.onStatusChange("open"));
  socket.addEventListener("close", () => handlers.onStatusChange("closed"));
  socket.addEventListener("error", () => handlers.onError("websocket connection error"));
  socket.addEventListener("message", (event) => {
    let message: GameSocketMessage;

    try {
      message = JSON.parse(event.data) as GameSocketMessage;
    } catch {
      handlers.onError("invalid websocket message");
      return;
    }

    switch (message.type) {
      case "game_state":
        handlers.onState(message.payload);
        break;
      case "game_finished":
        handlers.onFinished(message.payload);
        break;
      case "draw_offered":
        handlers.onDrawOffered(message.payload);
        break;
      case "draw_declined":
        handlers.onDrawDeclined(message.payload);
        break;
      case "error":
        handlers.onError(message.message);
        break;
      default:
        handlers.onError("unknown websocket message");
    }
  });

  return socket;
}

export function sendMove(socket: WebSocket | null, from: Position, to: Position): void {
  sendGameMessage(socket, {
    type: "move",
    payload: {
      from,
      to,
    },
  });
}

export function sendResign(socket: WebSocket | null): void {
  sendGameMessage(socket, {
    type: "resign",
    payload: {},
  });
}

export function sendDrawOffer(socket: WebSocket | null): void {
  sendGameMessage(socket, {
    type: "draw_offer",
    payload: {},
  });
}

export function sendDrawResponse(socket: WebSocket | null, accepted: boolean): void {
  sendGameMessage(socket, {
    type: "draw_response",
    payload: {
      accepted,
    },
  });
}

function sendGameMessage(socket: WebSocket | null, message: unknown): void {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    throw new Error("websocket is not connected");
  }

  socket.send(JSON.stringify(message));
}
