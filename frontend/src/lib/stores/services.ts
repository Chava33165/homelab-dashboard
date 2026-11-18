import { derived, writable } from "svelte/store";
import { api, ApiError } from "../api/client";
import { containers, refreshContainers } from "./containers";
import type { ContainerSummary } from "../api/types";

// Un "servicio" es un proyecto de compose (immich, asistente, ...): se prende y se
// apaga completo desde un solo switch. El backend decide el orden y protege el
// proyecto del propio dashboard -- aqui solo se agrupa. Lo usan la pestaña Servicios y
// el panel de RAM, asi ambos comparten estado (un "Apagando…" se ve en los dos).
export interface Service {
  project: string;
  members: ContainerSummary[];
  running: number;
  memBytes: number;
  protected: boolean;
}

export const services = derived(containers, (list) => {
  const byProject = new Map<string, Service>();
  for (const c of list) {
    let s = byProject.get(c.project);
    if (!s) {
      s = { project: c.project, members: [], running: 0, memBytes: 0, protected: c.protected };
      byProject.set(c.project, s);
    }
    s.members.push(c);
    if (c.state === "running") s.running++;
    s.memBytes += c.stats?.mem_usage_bytes ?? 0;
  }
  // El servicio del sistema (el propio dashboard) va al final: no se toca.
  return [...byProject.values()].sort(
    (a, b) => Number(a.protected) - Number(b.protected) || a.project.localeCompare(b.project)
  );
});

// project -> encendiendo (true) / apagando (false), mientras dura la peticion.
export const pending = writable<Record<string, boolean>>({});
export const powerErrors = writable<Record<string, string>>({});

export async function setPower(project: string, on: boolean): Promise<void> {
  pending.update((p) => ({ ...p, [project]: on }));
  powerErrors.update(({ [project]: _, ...rest }) => rest);
  try {
    await api(`/services/${encodeURIComponent(project)}/${on ? "start" : "stop"}`, { method: "POST" });
    await refreshContainers();
  } catch (err) {
    const msg = err instanceof ApiError ? err.message : "no se pudo completar";
    powerErrors.update((e) => ({ ...e, [project]: msg }));
  } finally {
    pending.update(({ [project]: _, ...rest }) => rest);
  }
}

export function isOn(s: Service, p: Record<string, boolean>): boolean {
  return s.project in p ? p[s.project] : s.running === s.members.length;
}

// Jugadores conectados si el servicio tiene un Terraria encendido; undefined si no es
// un juego.
export function playersOnline(s: Service): number | undefined {
  const game = s.members.find((c) => c.adapter === "terraria" && c.state === "running");
  const data = game?.adapter_data as { online?: string[] } | undefined;
  return game ? (data?.online?.length ?? 0) : undefined;
}

export function prettyName(project: string): string {
  return project.charAt(0).toUpperCase() + project.slice(1).replace(/[-_]/g, " ");
}

export function stopWarning(s: Service): string {
  const players = playersOnline(s);
  if (players) return `¡Hay ${players} jugador${players === 1 ? "" : "es"} conectado${players === 1 ? "" : "s"}! Si lo apagas, se les cae la partida.`;
  const n = s.members.length;
  return `Se ${n === 1 ? "va a detener su contenedor" : `van a detener sus ${n} contenedores`}. Lo puedes volver a prender con el mismo switch.`;
}
