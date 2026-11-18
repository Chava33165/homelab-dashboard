import { writable, get } from "svelte/store";
import { api } from "../api/client";
import { mainStream } from "../api/sse";
import type { ContainerSummary, DockerEvent } from "../api/types";

export const containers = writable<ContainerSummary[]>([]);
export const containersLoaded = writable(false);

export async function refreshContainers(): Promise<void> {
  const list = await api<ContainerSummary[]>("/containers");
  containers.set(list);
  containersLoaded.set(true);
}

let started = false;

export function startContainersStream(): void {
  if (started) return;
  started = true;

  // El snapshot periodico (cada ~5s) trae CPU/RAM actualizados de todos.
  mainStream.on<ContainerSummary[]>("containers", (list) => {
    containers.set(list);
    containersLoaded.set(true);
  });

  // Un evento discreto (create/start/stop/destroy/...) refresca de inmediato en vez de
  // esperar hasta 5s al siguiente snapshot -- asi un contenedor nuevo aparece al toque.
  mainStream.on<DockerEvent>("docker_event", () => {
    refreshContainers().catch(() => void 0);
  });

  mainStream.connect();
}

export function findContainer(id: string): ContainerSummary | undefined {
  return get(containers).find((c) => c.id === id);
}
