import { requestJSON } from "./http";
import type { Rating } from "../types/domain";

const RATING_BASE = import.meta.env.VITE_RATING_API_BASE ?? "/rating";

export async function getRating(userId: string): Promise<Rating | null> {
  try {
    return await requestJSON<Rating>(`${RATING_BASE}/api/v1/ratings/${userId}`);
  } catch (error) {
    if (error instanceof Error && "status" in error && error.status === 404) {
      return null;
    }

    throw error;
  }
}

export async function getLeaderboard(limit = 10): Promise<Rating[]> {
  const response = await requestJSON<{ items: Rating[] }>(
    `${RATING_BASE}/api/v1/leaderboard?limit=${limit}`,
  );

  return response.items;
}
