import { requestJSON } from "./http";
import type { GameRecord, UserGameHistoryPage } from "../types/domain";

const GAME_BASE = import.meta.env.VITE_GAME_API_BASE ?? "/game";

export function getGame(gameId: string): Promise<GameRecord> {
  return requestJSON<GameRecord>(`${GAME_BASE}/api/v1/games/${gameId}`);
}

export function getUserGameHistory(
  userId: string,
  cursor?: string,
  limit = 20,
): Promise<UserGameHistoryPage> {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) {
    params.set("cursor", cursor);
  }

  return requestJSON<UserGameHistoryPage>(
    `${GAME_BASE}/api/v1/users/${userId}/games?${params.toString()}`,
  );
}
