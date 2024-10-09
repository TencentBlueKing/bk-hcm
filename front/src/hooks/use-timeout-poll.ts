import { getCurrentScope, onScopeDispose, ref } from 'vue';
import type { Awaitable } from '@/typings';

export default function useTimeoutPoll(fn: () => Awaitable<void>, interval: number, options?: { immediate?: boolean }) {
  const { immediate = false } = options || {};

  const isActive = ref(false);

  let timer: ReturnType<typeof setTimeout> | null = null;

  function clear() {
    if (timer) {
      clearTimeout(timer);
      timer = null;
    }
  }

  function start() {
    clear();
    timer = setTimeout(() => {
      timer = null;

      loop();
    }, interval ?? 5000);
  }

  async function loop() {
    if (!isActive.value) {
      return;
    }

    await fn();
    start();
  }

  function resume() {
    if (!isActive.value) {
      isActive.value = true;
      loop();
    }
  }

  function pause() {
    isActive.value = false;
  }

  if (immediate) {
    resume();
  }

  if (getCurrentScope()) {
    onScopeDispose(pause);
  }

  return {
    isActive,
    pause,
    resume,
  };
}
