const TOKEN_STORAGE_KEY = "online-checkers.access-token";

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

export function getStoredToken(): string | null {
  return window.localStorage.getItem(TOKEN_STORAGE_KEY);
}

export function storeToken(token: string): void {
  window.localStorage.setItem(TOKEN_STORAGE_KEY, token);
}

export function clearStoredToken(): void {
  window.localStorage.removeItem(TOKEN_STORAGE_KEY);
}

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown;
};

export async function requestJSON<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  const token = getStoredToken();

  if (options.body !== undefined && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  if (token && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(path, {
    ...options,
    credentials: "include",
    headers,
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  if (!response.ok) {
    let code = "request_failed";
    let message = response.statusText || "request failed";

    try {
      const payload = await response.json();
      if (payload?.error?.code) {
        code = payload.error.code;
      }
      if (payload?.error?.message) {
        message = payload.error.message;
      }
    } catch {
      // Keep the HTTP status text when the server does not send JSON.
    }

    throw new ApiError(response.status, code, message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}
