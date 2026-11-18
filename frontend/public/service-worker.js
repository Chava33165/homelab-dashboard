// Service worker minimo: el objetivo es carga instantanea del shell (JS/CSS/fuentes),
// NUNCA cachear /api/* -- un panel de control mostrando metricas o estado de
// contenedores desactualizado como si fuera en vivo seria activamente enganoso, peor
// que no tener cache en absoluto.
//
// Estrategia:
//  - /api/*             -> siempre red (nunca cache). Es la unica ruta con datos vivos.
//  - /assets/*           -> cache-first. Vite les pone un hash en el nombre segun su
//                           contenido, asi que un archivo con esa URL nunca cambia --
//                           cachearlo agresivo es seguro y es lo que da la carga instantanea.
//  - todo lo demas (html, /) -> red primero, con la cache como respaldo si no hay red.
//
// JS plano a proposito (no TS/build) -- Vite lo copia tal cual desde public/ a
// dist/service-worker.js, sin bundling, para que la URL sea predecible y estable.

const CACHE_NAME = "homelab-shell-v1";

self.addEventListener("install", (event) => {
  event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    (async () => {
      const keys = await caches.keys();
      await Promise.all(keys.filter((k) => k !== CACHE_NAME).map((k) => caches.delete(k)));
      await self.clients.claim();
    })()
  );
});

self.addEventListener("fetch", (event) => {
  const url = new URL(event.request.url);

  if (event.request.method !== "GET" || url.origin !== self.location.origin) return;

  if (url.pathname.startsWith("/api/")) {
    return; // no interceptar -- va directo a red, sin cache
  }

  if (url.pathname.startsWith("/assets/")) {
    event.respondWith(cacheFirst(event.request));
    return;
  }

  event.respondWith(networkFirst(event.request));
});

async function cacheFirst(request) {
  const cached = await caches.match(request);
  if (cached) return cached;

  const response = await fetch(request);
  if (response.ok) {
    const cache = await caches.open(CACHE_NAME);
    cache.put(request, response.clone());
  }
  return response;
}

async function networkFirst(request) {
  try {
    const response = await fetch(request);
    if (response.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, response.clone());
    }
    return response;
  } catch (err) {
    const cached = await caches.match(request);
    if (cached) return cached;
    throw err;
  }
}
