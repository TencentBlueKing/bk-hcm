import { computed, inject, Ref } from 'vue';

/**
 * @param top 没有通知栏时，fixed元素距离视口顶部的距离
 */
export const useCalcTopWithNotice = (top: number): [Ref<string>, Ref<boolean>] => {
  const isNoticeAlert = inject<Ref<boolean>>('isNoticeAlert');

  const calcTop = computed(() => (isNoticeAlert.value ? `${top + 40}px` : `${top}px`));

  return [calcTop, isNoticeAlert];
};
