import { writable } from "svelte/store";
import { mainStream } from "../api/sse";
import type { MetricsSnapshot } from "../api/types";

const HISTORY_LEN = 60; // ~2min a 2s/tick, suficiente para un sparkline legible

export const latestMetrics = writable<MetricsSnapshot | null>(null);
export const cpuHistory = writable<number[]>([]);
export const tempHistory = writable<number[]>([]);
export const netRxHistory = writable<number[]>([]);
export const netTxHistory = writable<number[]>([]);

function pushHistory(store: typeof cpuHistory, value: number) {
  store.update((arr) => {
    const next = [...arr, value];
    return next.length > HISTORY_LEN ? next.slice(next.length - HISTORY_LEN) : next;
  });
}

// El backend reporta TODAS las interfaces del host (incluye pid:host, ver CLAUDE.md),
// pero sumarlas todas cuenta el mismo trafico varias veces (bridge + veth del mismo
// paquete). Para el total "en vivo" solo interesan las interfaces reales de salida.
const REAL_IFACE = /^(eth|wlan|en|tailscale)/;

function sumReal(bytesByIface: Record<string, number>): number {
  let total = 0;
  for (const [iface, v] of Object.entries(bytesByIface)) {
    if (REAL_IFACE.test(iface)) total += v;
  }
  return total;
}

let started = false;

export function startMetricsStream(): void {
  if (started) return;
  started = true;

  mainStream.on<MetricsSnapshot>("metrics", (snap) => {
    latestMetrics.set(snap);
    pushHistory(cpuHistory, snap.cpu_percent);
    pushHistory(tempHistory, snap.temp_c);
    pushHistory(netRxHistory, sumReal(snap.net.rx_bytes_per_sec));
    pushHistory(netTxHistory, sumReal(snap.net.tx_bytes_per_sec));
  });

  mainStream.connect();
}
