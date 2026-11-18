<script module lang="ts">
  export type Gesto = "neutral" | "feliz" | "pensando" | "preocupada" | "hablando" | "guino";
</script>

<script lang="ts">
  import { onDestroy, onMount } from "svelte";

  // Cashy, estilo caricatura tipo Gorillaz: contorno negro grueso, pelo esponjado y rizado, ojos
  // grandes en blanco con ojeras. Cada gesto es un juego de trazos que entra/sale con opacidad y
  // transform (barato, GPU). En reposo parpadea, se mece y de vez en cuando guiña o sonrie sola
  // (`vivo`). Con "reducir movimiento" se queda quieta.
  let {
    gesto = "neutral",
    size = 40,
    titulo = "Cashy",
    vivo = true,
  }: { gesto?: Gesto; size?: number; titulo?: string; vivo?: boolean } = $props();

  let micro = $state<Gesto | null>(null);
  let timer = 0;
  const quieta = typeof window !== "undefined" && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  // Solo en reposo (neutral/feliz) hace gestos sueltos: guiño o sonrisa, ~1.3 s cada 5-9 s.
  function siguiente() {
    timer = window.setTimeout(() => {
      if (gesto === "neutral" || gesto === "feliz") {
        micro = Math.random() < 0.5 ? "guino" : "feliz";
        timer = window.setTimeout(() => {
          micro = null;
          siguiente();
        }, 1300);
      } else siguiente();
    }, 5000 + Math.random() * 4000);
  }

  onMount(() => {
    if (vivo && !quieta) siguiente();
  });
  onDestroy(() => clearTimeout(timer));

  let actual = $derived(micro && (gesto === "neutral" || gesto === "feliz") ? micro : gesto);
</script>

<svg class="cara g-{actual}" viewBox="0 0 120 120" width={size} height={size} role="img" aria-label={titulo}
  stroke="#141317" stroke-linecap="round" stroke-linejoin="round">
  <g class="mece">
    <!-- pelo de atras: nube de rizos -->
    <g class="pelo" stroke-width="2.4">
      <circle cx="34" cy="40" r="13" /><circle cx="27" cy="56" r="12" /><circle cx="29" cy="73" r="11" />
      <circle cx="35" cy="87" r="9" /><circle cx="47" cy="27" r="13" /><circle cx="61" cy="22" r="14" />
      <circle cx="75" cy="27" r="13" /><circle cx="87" cy="40" r="13" /><circle cx="93" cy="56" r="12" />
      <circle cx="91" cy="73" r="11" /><circle cx="85" cy="87" r="9" />
    </g>
    <!-- brillo de los rizos -->
    <g fill="none" stroke="#4b4a8f" stroke-width="1.8">
      <path d="M24 52 q4 -5 9 -1 M23 70 q4 -4 8 0 M42 22 q5 -5 10 -1 M58 16 q5 -4 10 0 M80 33 q5 -4 9 1 M90 51 q4 -4 8 1 M88 69 q4 -3 7 1" />
    </g>

    <!-- orejas -->
    <ellipse cx="35.5" cy="63" rx="4.5" ry="6.5" fill="#efbfa0" stroke-width="2.2" />
    <ellipse cx="84.5" cy="63" rx="4.5" ry="6.5" fill="#efbfa0" stroke-width="2.2" />

    <!-- cara (barbilla un poco en punta, como el estilo de Hewlett) -->
    <path d="M37 52 C36 76 47 94 60 96 C73 94 84 76 83 52 C83 38 73 31 60 31 C47 31 37 38 37 52 Z"
      fill="#f4c9ad" stroke-width="2.8" />
    <path d="M47 89 Q60 99 73 89" fill="none" stroke="#d99b84" stroke-width="2" opacity="0.7" />

    <!-- fleco rizado -->
    <g class="pelo" stroke-width="2.2">
      <circle cx="45" cy="38" r="8.5" /><circle cx="55" cy="33" r="9" /><circle cx="66" cy="33" r="9" />
      <circle cx="76" cy="38" r="8.5" /><circle cx="39" cy="46" r="6" /><circle cx="82" cy="46" r="6" />
    </g>
    <!-- broche -->
    <g stroke-width="1.8">
      <circle cx="84" cy="33" r="4.2" fill="#ff4f9a" />
      <circle cx="84" cy="33" r="1.4" fill="#fff" stroke="none" />
    </g>

    <!-- cachetes -->
    <g class="cachetes" stroke="none">
      <ellipse cx="42" cy="77" rx="5" ry="3" fill="#ff7aa8" opacity="0.55" />
      <ellipse cx="78" cy="77" rx="5" ry="3" fill="#ff7aa8" opacity="0.55" />
    </g>

    <!-- ojeras moradas -->
    <g fill="none" stroke="#8a5d78" stroke-width="2" opacity="0.65">
      <path d="M39 68 Q48 74 57 68" /><path d="M63 68 Q72 74 81 68" />
    </g>

    <!-- cejas -->
    <g class="cejas normal" fill="none" stroke-width="3.2">
      <path d="M40 48 Q47 45 55 47.5" /><path d="M65 47.5 Q73 45 80 48" />
    </g>
    <g class="cejas tristes" fill="none" stroke-width="3.2">
      <path d="M40 50 Q47 49 55 45" /><path d="M65 45 Q73 49 80 50" />
    </g>

    <!-- ojos: grandes y blancos (parpadean) -->
    <g class="ojos">
      <g class="ojo-izq">
        <ellipse class="abierto" cx="48" cy="60" rx="8.6" ry="8" fill="#fff" stroke-width="2.6" />
        <path class="cerrado" d="M40 61 Q48 55 56 61" fill="none" stroke-width="2.8" />
      </g>
      <g class="ojo-der">
        <ellipse class="abierto" cx="72" cy="60" rx="8.6" ry="8" fill="#fff" stroke-width="2.6" />
        <path class="cerrado" d="M64 61 Q72 55 80 61" fill="none" stroke-width="2.8" />
      </g>
      <g class="brillo" fill="#141317" stroke="none">
        <circle cx="49.5" cy="58.5" r="1.3" /><circle cx="73.5" cy="58.5" r="1.3" />
      </g>
    </g>

    <!-- nariz chiquita con orificios (como 2D) -->
    <path d="M56.5 72 Q60 74.5 63.5 72" fill="none" stroke-width="2" />
    <circle cx="58" cy="72.6" r="0.9" fill="#141317" stroke="none" />
    <circle cx="62" cy="72.6" r="0.9" fill="#141317" stroke="none" />

    <!-- bocas -->
    <g class="boca b-neutral">
      <path d="M50 81 Q60 86 71 80" fill="none" stroke-width="2.6" />
      <path d="M62 84 l0 2.6 l3 -0.4 l-0.3 -2.6" fill="#fff" stroke-width="1.4" />
    </g>
    <g class="boca b-feliz">
      <path d="M48 79 Q60 93 72 79 Q60 83 48 79 Z" fill="#5a1d2b" stroke-width="2.4" />
      <path d="M50.5 80.3 Q60 84 69.5 80.3 L68.6 82.4 Q60 85 51.4 82.4 Z" fill="#fff" stroke="none" />
    </g>
    <g class="boca b-guino">
      <path d="M50 80 Q61 88 71 78" fill="none" stroke-width="2.6" />
      <path d="M58 83.5 l0.6 2.8 l3 -0.6 l-0.6 -2.6" fill="#fff" stroke-width="1.4" />
    </g>
    <ellipse class="boca b-pensando" cx="63" cy="82" rx="2.6" ry="3" fill="#5a1d2b" stroke-width="2.2" />
    <path class="boca b-preocupada" d="M51 85 Q60 79 69 85" fill="none" stroke-width="2.6" />
    <ellipse class="boca b-hablando" cx="60" cy="82" rx="5" ry="4" fill="#5a1d2b" stroke-width="2.4" />

    <!-- burbujas de "pensando" -->
    <g class="burbuja" fill="#fff" stroke-width="1.8">
      <circle cx="99" cy="58" r="2.4" /><circle cx="105" cy="48" r="3.6" />
    </g>
  </g>
</svg>

<style>
  .cara {
    display: block;
    overflow: visible;
  }

  .pelo {
    fill: #23214a;
  }

  .mece {
    transform-origin: 60px 96px;
    animation: mece 5.5s ease-in-out infinite;
  }

  .ojos {
    transform-origin: 60px 60px;
    animation: parpadeo 5.2s infinite;
  }

  .brillo {
    transition: transform 240ms ease, opacity 180ms ease;
  }

  .boca,
  .cejas,
  .cachetes,
  .abierto,
  .cerrado,
  .burbuja {
    transition: opacity 180ms ease;
  }

  /* por defecto (neutral) */
  .boca,
  .cejas.tristes,
  .cerrado,
  .burbuja {
    opacity: 0;
  }
  .b-neutral {
    opacity: 1;
  }
  .cachetes {
    opacity: 0.6;
  }

  /* feliz: sonrisa grande, ojos en arco, cachetes */
  .g-feliz .b-neutral,
  .g-feliz .abierto,
  .g-feliz .brillo {
    opacity: 0;
  }
  .g-feliz .b-feliz,
  .g-feliz .cerrado,
  .g-feliz .cachetes {
    opacity: 1;
  }

  /* guiño: cierra el ojo derecho, sonrisa chueca */
  .g-guino .b-neutral,
  .g-guino .ojo-der .abierto,
  .g-guino .brillo circle:last-child {
    opacity: 0;
  }
  .g-guino .b-guino,
  .g-guino .ojo-der .cerrado,
  .g-guino .cachetes {
    opacity: 1;
  }

  /* pensando: mira hacia arriba, boquita, burbujas */
  .g-pensando .b-neutral {
    opacity: 0;
  }
  .g-pensando .b-pensando,
  .g-pensando .burbuja {
    opacity: 1;
  }
  .g-pensando .brillo {
    transform: translate(1.5px, -3.5px);
  }
  .g-pensando .burbuja {
    animation: flota 1.6s ease-in-out infinite;
  }

  /* preocupada: cejas hacia arriba por dentro, boca hacia abajo */
  .g-preocupada .b-neutral,
  .g-preocupada .cejas.normal {
    opacity: 0;
  }
  .g-preocupada .b-preocupada,
  .g-preocupada .cejas.tristes {
    opacity: 1;
  }
  .g-preocupada .cachetes {
    opacity: 0.25;
  }

  /* hablando: la boca se abre y cierra */
  .g-hablando .b-neutral {
    opacity: 0;
  }
  .g-hablando .b-hablando {
    opacity: 1;
    transform-origin: 60px 82px;
    animation: habla 0.3s ease-in-out infinite alternate;
  }

  @keyframes parpadeo {
    0%,
    93%,
    100% {
      transform: scaleY(1);
    }
    96% {
      transform: scaleY(0.12);
    }
  }

  @keyframes mece {
    0%,
    100% {
      transform: rotate(-3deg) translateY(0);
    }
    50% {
      transform: rotate(3deg) translateY(-1.5px);
    }
  }

  @keyframes habla {
    from {
      transform: scaleY(0.35);
    }
    to {
      transform: scaleY(1.1);
    }
  }

  @keyframes flota {
    0%,
    100% {
      transform: translateY(0);
    }
    50% {
      transform: translateY(-2.5px);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .mece,
    .ojos,
    .g-hablando .b-hablando,
    .g-pensando .burbuja {
      animation: none;
    }
  }
</style>
