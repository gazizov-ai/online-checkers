import { requestJSON } from "./http";
import type { GameRecord } from "../types/domain";

const GAME_BASE = import.meta.env.VITE_GAME_API_BASE ?? "/game";

export function getGame(gameId: string): Promise<GameRecord> {
  return requestJSON<GameRecord>(`${GAME_BASE}/api/v1/games/${gameId}`);
}
