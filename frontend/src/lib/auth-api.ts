import { apiRequest } from "@/lib/api";
import type { AuthResponse, LoginRequest } from "@/types/api";

export function login(payload: LoginRequest): Promise<AuthResponse> {
  return apiRequest<AuthResponse>("/api/auth/login", {
    method: "POST",
    body: payload,
  });
}

export function logout(): Promise<{ ok: boolean }> {
  return apiRequest<{ ok: boolean }>("/api/auth/logout", {
    method: "POST",
  });
}

export function getCurrentUser(): Promise<AuthResponse> {
  return apiRequest<AuthResponse>("/api/me");
}

