<script lang="ts">
  let {
    value,
    decimals = 0,
    suffix = "",
  }: { value: number; decimals?: number; suffix?: string } = $props();

  // svelte-ignore state_referenced_locally -- el valor inicial real es intencional: la
  // primera pintura muestra el numero de una vez, $effect anima solo los cambios luego.
  let displayed = $state(value);
  const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  let raf = 0;
  $effect(() => {
    const target = value;
    if (reduceMotion) {
      displayed = target;
      return;
    }
    const start = displayed;
    const delta = target - start;
    if (delta === 0) return;

    const duration = 420;
    const t0 = performance.now();
    cancelAnimationFrame(raf);

    function tick(now: number) {
      const p = Math.min(1, (now - t0) / duration);
      // ease-out cubico -- coherente con --ease-out del CSS sin depender de la cascada
      const eased = 1 - Math.pow(1 - p, 3);
      displayed = start + delta * eased;
      if (p < 1) raf = requestAnimationFrame(tick);
      else displayed = target;
    }
    raf = requestAnimationFrame(tick);

    return () => cancelAnimationFrame(raf);
  });
</script>

<span>{displayed.toFixed(decimals)}{suffix}</span>
