import { Ref, ref, watch } from 'vue';

/**
 * 获取当前状态的前一个取值
 */
export function usePreviousState<T>(state: Ref<T> | (() => T)) {
  const previous = ref<T>();
  watch(state, (_, oldVal) => {
    previous.value = oldVal;
  });
  return previous;
}
