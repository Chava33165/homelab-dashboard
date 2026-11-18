// Wrapper de fetch: agrega el header que exige el backend en mutaciones (defensa CSRF,
// ver internal/auth/middleware.go) y credentials:"include" para que viaje el cookie de
// sesion. Un 401 dispara un evento global "auth:unauthorized" -- el store de auth
// escucha eso para volver a la pantalla de login sin que cada llamada tenga que
// manejarlo por separado.

const MUTATING = new Set(["POST", "PUT", "PATCH", "DELETE"]);

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const method = (init.method ?? "GET").toUpperCase();
  const headers = new Headers(init.headers);
  if (MUTATING.has(method)) {
    headers.set("X-Requested-With", "homelab-dashboard");
  }
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const res = await fetch(`/api${path}`, {
    ...init,
    method,
    headers,
    credentials: "include",
  });

  if (res.status === 401) {
    window.dispatchEvent(new CustomEvent("auth:unauthorized"));
    throw new ApiError(401, "no autenticado");
  }

  const text = await res.text();
  const data = text ? JSON.parse(text) : undefined;

  if (!res.ok) {
    throw new ApiError(res.status, errorMessage(data) || res.statusText);
  }
  return data as T;
}

// El backend propio responde {error}; task-hub (FastAPI, via /api/finanzas) responde {detail},
// que en errores de validacion es una lista [{msg}].
function errorMessage(data: unknown): string | undefined {
  if (!data || typeof data !== "object") return undefined;
  const d = data as { error?: string; detail?: string | { msg?: string }[] };
  if (d.error) return d.error;
  if (typeof d.detail === "string") return d.detail;
  if (Array.isArray(d.detail) && d.detail[0]?.msg) return d.detail[0].msg.replace(/^Value error, /, "");
  return undefined;
}
