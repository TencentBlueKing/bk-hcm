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

  const from = computed(() => {
    if (Object.hasOwn(route.query, '_f')) {
      try {
        return HistoryStorage.pop();
      } catch (error) {
        return defaultFrom.value;
      }
    }
    return defaultFrom.value;
  });

  // 引入 fromConfig 是为了解决业务下 defaultFrom 没有业务ID 的问题
  const handleBack = (fromConfig: Partial<RouteLocationRaw> = {}) => {
    routerAction.redirect(merge({}, from.value, fromConfig), { back: true });
  };

  return { from, handleBack };
};
