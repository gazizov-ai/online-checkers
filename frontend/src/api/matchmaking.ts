import { requestJSON } from "./http";
import type { MatchmakingResponse } from "../types/domain";

const MATCHMAKING_BASE = import.meta.env.VITE_MATCHMAKING_API_BASE ?? "/matchmaking";

export function searchMatch(): Promise<MatchmakingResponse> {
  return requestJSON<MatchmakingResponse>(`${MATCHMAKING_BASE}/api/v1/matchmaking/search`, {
    method: "POST",
  });
}

export function getMatchmakingStatus(): Promise<MatchmakingResponse> {
  return requestJSON<MatchmakingResponse>(`${MATCHMAKING_BASE}/api/v1/matchmaking/status`);
}

export function cancelMatchmaking(): Promise<{ status: string }> {
  return requestJSON<{ status: string }>(`${MATCHMAKING_BASE}/api/v1/matchmaking/cancel`, {
    method: "POST",
  });
}
