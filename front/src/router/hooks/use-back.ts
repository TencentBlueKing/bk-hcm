import { computed } from 'vue';
import { RouteLocationRaw, useRoute } from 'vue-router';
import { RouteMetaConfig } from '../meta';
import { HistoryStorage } from '../utils/history-storage';
import routerAction from '../utils/action';
import { merge } from 'lodash';

export const useBack = () => {
  const route = useRoute();

  const defaultFrom = computed(() => {
    const routeMeta = route.meta as RouteMetaConfig;
    const menu = routeMeta.menu || {};
    if (menu.relative) {
      return { name: Array.isArray(menu.relative) ? menu.relative[0] : menu.relative };
    }
    return null;
  });

  // 只 peek：渲染面包屑时不能 pop，否则取消按钮再 back 时栈已空、跳不回去
  const from = computed(() => {
    if (Object.hasOwn(route.query, '_f')) {
      return HistoryStorage.peek() ?? defaultFrom.value;
    }
    return defaultFrom.value;
  });

  // fromConfig：补业务 ID；无 history / relative 时作为真正目的地（如提交后进详情）
  const handleBack = (fromConfig: Partial<RouteLocationRaw> = {}) => {
    const hasFromConfigTarget = Boolean(fromConfig?.name || fromConfig?.path);
    const target = from.value ?? (hasFromConfigTarget ? fromConfig : null);
    if (!target) return;
    if (Object.hasOwn(route.query, '_f')) {
      try {
        HistoryStorage.pop();
      } catch {
        // 栈空时仍按 peek/defaultFrom 跳，避免返回按钮完全失效
      }
    }
    routerAction.redirect(merge({}, target, fromConfig), { back: true });
  };

  return { from, handleBack };
};
