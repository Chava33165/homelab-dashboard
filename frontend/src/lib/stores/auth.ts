import { writable } from "svelte/store";
import { api } from "../api/client";

export type AuthState = "unknown" | "authenticated" | "anonymous";

export const authState = writable<AuthState>("unknown");

export async function checkSession(): Promise<void> {
  try {
    const res = await api<{ authenticated: boolean }>("/session");
    authState.set(res.authenticated ? "authenticated" : "anonymous");
  } catch {
    authState.set("anonymous");
  }
}

export async function login(password: string): Promise<void> {
  await api("/login", { method: "POST", body: JSON.stringify({ password }) });
  authState.set("authenticated");
}

export async function logout(): Promise<void> {
  try {
    await api("/logout", { method: "POST" });
  } finally {
    authState.set("anonymous");
  }
}

// Cualquier llamada que reciba 401 dispara esto (ver api/client.ts) -- cubre el caso de
// que la sesion expire mientras la app ya esta abierta.
window.addEventListener("auth:unauthorized", () => authState.set("anonymous"));
