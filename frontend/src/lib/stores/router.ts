import { writable } from "svelte/store";

// Router minimo a mano -- la app tiene un puñado de vistas, no justifica una libreria
// de rutas. Cada cambio de vista pasa por la View Transitions API (con fallback
// silencioso si el navegador no la soporta) para las transiciones suaves entre pantallas.

export type Route =
  | { name: "dashboard" }
  | { name: "containers" }
  | { name: "container"; id: string }
  | { name: "korra" }
  | { name: "dinero" }
  | { name: "settings" };

function parse(path: string): Route {
  const containerMatch = path.match(/^\/containers\/([^/]+)$/);
  if (containerMatch) return { name: "container", id: containerMatch[1] };
  if (path === "/containers") return { name: "containers" };
  if (path === "/settings") return { name: "settings" };
  if (path === "/korra") return { name: "korra" };
  if (path === "/dinero") return { name: "dinero" };
  return { name: "dashboard" };
}

function toPath(route: Route): string {
  switch (route.name) {
    case "containers":
      return "/containers";
    case "container":
      return `/containers/${route.id}`;
    case "settings":
      return "/settings";
    case "korra":
      return "/korra";
    case "dinero":
      return "/dinero";
    default:
      return "/";
  }
}

export const route = writable<Route>(parse(window.location.pathname));

window.addEventListener("popstate", () => {
  route.set(parse(window.location.pathname));
});

export function navigate(r: Route): void {
  const path = toPath(r);
  if (path === window.location.pathname) return;

  const apply = () => {
    history.pushState(null, "", path);
    route.set(r);
  };

  if (document.startViewTransition && !window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
    document.startViewTransition(apply);
  } else {
    apply();
  }
}
