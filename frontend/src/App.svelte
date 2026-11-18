<script lang="ts">
  import { onMount } from "svelte";
  import { authState, checkSession } from "./lib/stores/auth";
  import { route } from "./lib/stores/router";
  import Login from "./lib/views/Login.svelte";
  import Dashboard from "./lib/views/Dashboard.svelte";
  import Containers from "./lib/views/Containers.svelte";
  import ContainerDetail from "./lib/views/ContainerDetail.svelte";
  import Settings from "./lib/views/Settings.svelte";
  import Korra from "./lib/views/Korra.svelte";
  import Dinero from "./lib/views/Dinero.svelte";
  import NavTabBar from "./lib/components/NavTabBar.svelte";

  onMount(() => {
    checkSession();
  });
</script>

{#if $authState === "unknown"}
  <div class="splash">
    <div class="splash-badge">
      <svg viewBox="0 0 24 24" width="36" height="36" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 2v20M4 7l8 5 8-5M4 17l8-5 8 5M2 12h20" opacity="0.6"/>
        <circle cx="12" cy="12" r="3.5" fill="var(--c-action)" fill-opacity="0.3"/>
      </svg>
    </div>
    <span class="brand chromatic-pop">Homelab</span>
    <span class="splash-sub">Cargando servicios…</span>
  </div>
{:else if $authState === "anonymous"}
  <Login />
{:else}
  {#if $route.name === "dashboard"}
    <Dashboard />
  {:else if $route.name === "containers"}
    <Containers />
  {:else if $route.name === "container"}
    {#key $route.id}
      <ContainerDetail id={$route.id} />
    {/key}
  {:else if $route.name === "settings"}
    <Settings />
  {:else if $route.name === "korra"}
    <Korra />
  {:else if $route.name === "dinero"}
    <Dinero />
  {/if}
  <NavTabBar />
{/if}

<style>
  .splash {
    min-height: 100dvh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
  }
  .splash-badge {
    width: 64px;
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: color-mix(in srgb, var(--c-glass) 16%, transparent);
    border: 1px solid var(--border-strong);
    color: var(--c-action);
    box-shadow: 
      inset 0 1px 0 var(--glass-specular),
      0 0 24px color-mix(in srgb, var(--c-action) 35%, transparent);
    margin-bottom: 8px;
    animation: breathe 2s var(--ease-spring) infinite;
  }
  .brand {
    font-family: var(--font-display);
    font-size: 32px;
    font-weight: 800;
    background: linear-gradient(135deg, var(--text) 30%, var(--c-action));
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .splash-sub {
    font-size: 13px;
    color: var(--text-secondary);
    font-weight: 500;
  }
  @keyframes breathe {
    0%,
    100% {
      transform: scale(0.95);
      opacity: 0.7;
    }
    50% {
      transform: scale(1.05);
      opacity: 1;
    }
  }
</style>
