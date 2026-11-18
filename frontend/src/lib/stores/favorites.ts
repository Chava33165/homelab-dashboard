import { writable } from "svelte/store";

// Favoritos: preferencia puramente del navegador (no hay "usuario" que sincronizar
// entre dispositivos en este proyecto -- ver "sin base de datos" en CLAUDE.md), asi que
// vive en localStorage igual que el tema, nunca en el backend.

const STORAGE_KEY = "homelab-favorite-containers";

function readStored(): Set<string> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return new Set(JSON.parse(raw));
  } catch {
    /* localStorage puede fallar en modo privado -- se cae a sin favoritos */
  }
  return new Set();
}

export const favorites = writable<Set<string>>(readStored());

favorites.subscribe((value) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify([...value]));
  } catch {
    /* ignorar si no hay storage disponible */
  }
});

export function toggleFavorite(id: string): void {
  favorites.update((current) => {
    const next = new Set(current);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    return next;
  });
}
