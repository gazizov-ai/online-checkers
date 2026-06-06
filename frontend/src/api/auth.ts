import { requestJSON, storeToken } from "./http";
import type { User } from "../types/domain";

const AUTH_BASE = import.meta.env.VITE_AUTH_API_BASE ?? "/auth";

type AuthUserResponse = {
  user: User;
};

type LoginResponse = AuthUserResponse & {
  access_token: string;
  id_token: string;
  token_type: string;
  expires_in: number;
};

export type RegisterInput = {
  username: string;
  email?: string;
  password: string;
};

export type LoginInput = {
  username: string;
  password: string;
};

export async function register(input: RegisterInput): Promise<User> {
  const response = await requestJSON<AuthUserResponse>(`${AUTH_BASE}/api/v1/register`, {
    method: "POST",
    body: {
      username: input.username,
      email: input.email || undefined,
      password: input.password,
    },
  });

  return response.user;
}

export async function login(input: LoginInput): Promise<User> {
  const response = await requestJSON<LoginResponse>(`${AUTH_BASE}/api/v1/login`, {
    method: "POST",
    body: input,
  });

  storeToken(response.access_token);

  return response.user;
}

export async function me(): Promise<User> {
  const response = await requestJSON<AuthUserResponse>(`${AUTH_BASE}/api/v1/me`);

  return response.user;
}
