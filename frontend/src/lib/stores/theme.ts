import { writable } from "svelte/store";

// Dos temas: Tinta (oscuro, default) y Papel (claro). Los valores viejos guardados en
// el navegador ("spider", "system") se migran a Tinta.
export type Theme = "dark" | "light";

const STORAGE_KEY = "homelab-theme";

function readStored(): Theme {
  try {
    if (localStorage.getItem(STORAGE_KEY) === "light") return "light";
  } catch {
    /* localStorage puede fallar en modo privado -- se cae al default */
  }
  return "dark";
}

export const theme = writable<Theme>(readStored());

theme.subscribe((value) => {
  document.documentElement.setAttribute("data-theme", value);
  const meta = document.querySelector('meta[name="theme-color"]');
  meta?.setAttribute("content", value === "light" ? "#f4f2ec" : "#0b0b0c");
  try {
    localStorage.setItem(STORAGE_KEY, value);
  } catch {
    /* ignorar si no hay storage disponible */
  }
});

export function toggleTheme(): void {
  theme.update((current) => (current === "dark" ? "light" : "dark"));
}
