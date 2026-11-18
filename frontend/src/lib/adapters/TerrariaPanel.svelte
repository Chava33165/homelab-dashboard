<script lang="ts">
  import { onMount } from "svelte";
  import { mainStream } from "../api/sse";
  import GlassCard from "../components/GlassCard.svelte";

  interface PlayerEvent {
    name: string;
    type: "joined" | "left";
    time: string;
  }
  interface TerrariaSnapshot {
    online: string[];
    recent: PlayerEvent[];
  }
  interface AdapterEventMsg {
    container_id: string;
    adapter: string;
    type: string;
    data: PlayerEvent;
    time: string;
  }

  let { data, containerId }: { data: TerrariaSnapshot | null | undefined; containerId: string } = $props();

  // svelte-ignore state_referenced_locally
  let online = $state<string[]>(data?.online ?? []);
  // svelte-ignore state_referenced_locally
  let recent = $state<PlayerEvent[]>(data?.recent ?? []);

  $effect(() => {
    if (data) {
      online = data.online;
      recent = data.recent;
    }
  });

  onMount(() => {
    return mainStream.on<AdapterEventMsg>("adapter_event", (ev) => {
      if (ev.container_id !== containerId || ev.adapter !== "terraria") return;

      if (ev.type === "player_joined") {
        online = [...new Set([...online, ev.data.name])];
      } else if (ev.type === "player_left") {
        online = online.filter((n) => n !== ev.data.name);
      }
      recent = [...recent, ev.data].slice(-30);
    });
  });

  function relativeTime(iso: string): string {
    const diff = (Date.now() - new Date(iso).getTime()) / 1000;
    if (diff < 60) return "hace instantes";
    if (diff < 3600) return `hace ${Math.floor(diff / 60)}m`;
    return `hace ${Math.floor(diff / 3600)}h`;
  }
</script>

<GlassCard comic>
  <div class="head">
    <div>
      <span class="comic-chip">TERRARIA SERVER</span>
      <h3 class="title">Jugadores Online</h3>
    </div>
    <span class="count">{online.length}</span>
  </div>

  {#if online.length > 0}
    <div class="players">
      {#each online as name (name)}
        <span class="player-pill">
          <span class="dot"></span>
          <span class="player-name">{name}</span>
        </span>
      {/each}
    </div>
  {:else}
    <p class="empty">Nadie conectado en este momento.</p>
  {/if}

  {#if recent.length > 0}
    <div class="history">
      <span class="history-label">HISTORIAL RECIENTE</span>
      <ul>
        {#each [...recent].reverse() as ev, i (ev.time + ev.name + i)}
          <li>
            <span class="hist-dot" class:left={ev.type === "left"}></span>
            <span class="hist-name">{ev.name}</span>
            <span class="hist-action">{ev.type === "joined" ? "se unió" : "salió"}</span>
            <span class="hist-time">{relativeTime(ev.time)}</span>
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</GlassCard>

<style>
  .head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 14px;
  }
  .comic-chip {
    font-size: 9px;
    font-weight: 800;
    letter-spacing: 0.08em;
    color: var(--c-accent-secondary);
    text-transform: uppercase;
  }
  .title {
    font-size: 20px;
    font-weight: 800;
    margin-top: 2px;
  }
  .count {
    font-family: var(--font-display);
    font-size: 34px;
    font-weight: 800;
    color: var(--c-accent-secondary);
    text-shadow: 0 0 10px color-mix(in srgb, var(--c-accent-secondary) 40%, transparent);
  }
  .players {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 6px;
  }
  .player-pill {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 6px 14px;
    border-radius: 99em;
    background: color-mix(in srgb, var(--c-glass) 14%, transparent);
    border: 1px solid var(--border-strong);
    font-size: 13px;
    font-weight: 600;
    box-shadow: inset 0 1px 0 var(--glass-specular);
    animation: comicBadgePop 260ms var(--ease-out);
  }
  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--c-accent-secondary);
    box-shadow: 0 0 8px var(--c-accent-secondary);
  }
  .empty {
    font-size: 13px;
    color: var(--text-tertiary);
    margin: 4px 0 0;
  }
  .history {
    margin-top: 18px;
    padding-top: 14px;
    border-top: 1px solid var(--border);
  }
  .history-label {
    font-size: 10px;
    letter-spacing: 0.06em;
    color: var(--text-tertiary);
    font-weight: 700;
  }
  ul {
    list-style: none;
    margin: 10px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-height: 200px;
    overflow-y: auto;
  }
  li {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
  }
  .hist-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--c-accent-secondary);
    flex-shrink: 0;
    box-shadow: 0 0 6px var(--c-accent-secondary);
  }
  .hist-dot.left {
    background: var(--text-tertiary);
    box-shadow: none;
  }
  .hist-name {
    font-weight: 600;
  }
  .hist-action {
    color: var(--text-secondary);
  }
  .hist-time {
    margin-left: auto;
    color: var(--text-tertiary);
    font-size: 12px;
    flex-shrink: 0;
  }
</style>
