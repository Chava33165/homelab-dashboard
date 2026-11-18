import { writable } from "svelte/store";
import { api } from "../api/client";
import type { FinMovimiento, FinNomina, FinResumen } from "../api/types";

// Estado de la pestaña Dinero. task-hub es el dueño de los datos: aqui solo se piden y, tras
// cualquier cambio, se vuelve a pedir el resumen (las reglas viven en el servidor, no se duplican).
export const resumen = writable<FinResumen | null>(null);
export const movimientos = writable<FinMovimiento[]>([]);
export const nominas = writable<FinNomina[]>([]);
export const finError = writable<string | null>(null);

export async function refrescar(): Promise<void> {
  try {
    const [r, m] = await Promise.all([
      api<FinResumen>("/finanzas/resumen"),
      api<FinMovimiento[]>("/finanzas/movimientos?limit=80"),
    ]);
    resumen.set(r);
    movimientos.set(m);
    finError.set(null);
  } catch (e) {
    finError.set(e instanceof Error ? e.message : "No se pudo cargar tu dinero.");
  }
}

export async function cargarNominas(): Promise<void> {
  nominas.set(await api<FinNomina[]>("/finanzas/nominas"));
}

/** Manda un cambio y refresca. Devuelve la respuesta (p. ej. el `mensaje`) o lanza el error de task-hub. */
export async function accion<T = { mensaje?: string }>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await api<T>(`/finanzas/${path}`, { method, body: body === undefined ? undefined : JSON.stringify(body) });
  await refrescar();
  return res;
}

export function pesos(n: number, decimals?: number): string {
  const d = decimals ?? (Math.round(n * 100) % 100 === 0 ? 0 : 2);
  return (n < 0 ? "−$" : "$") + Math.abs(n).toLocaleString("es-MX", { minimumFractionDigits: d, maximumFractionDigits: d });
}

const MESES = ["ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"];

export function fecha(iso: string): string {
  const [y, m, d] = iso.split("-").map(Number);
  return `${d} ${MESES[m - 1]}${y !== new Date().getFullYear() ? ` ${y}` : ""}`;
}
