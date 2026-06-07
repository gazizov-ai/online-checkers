import type { Profile, UpdateProfileInput } from "../types/domain";
import { requestJSON } from "./http";

const PROFILE_BASE = import.meta.env.VITE_PROFILE_API_BASE ?? "/profile";

export function getMyProfile(): Promise<Profile> {
  return requestJSON<Profile>(`${PROFILE_BASE}/api/v1/profiles/me`);
}

export function getProfile(userId: string): Promise<Profile> {
  return requestJSON<Profile>(`${PROFILE_BASE}/api/v1/profiles/${userId}`);
}

export async function getProfiles(userIds: string[]): Promise<Profile[]> {
  const uniqueIds = [...new Set(userIds)].filter(Boolean);
  if (uniqueIds.length === 0) {
    return [];
  }

  const response = await requestJSON<{ profiles: Profile[] }>(
    `${PROFILE_BASE}/api/v1/profiles/batch`,
    {
      method: "POST",
      body: { user_ids: uniqueIds },
    },
  );

  return response.profiles;
}

export function updateMyProfile(input: UpdateProfileInput): Promise<Profile> {
  return requestJSON<Profile>(`${PROFILE_BASE}/api/v1/profiles/me`, {
    method: "PATCH",
    body: input,
  });
}

export function profileName(profile?: Profile | null): string {
  return profile?.display_name || profile?.username || "Игрок";
}
