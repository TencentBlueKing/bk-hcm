import { inject, provide, reactive, watch } from 'vue';
import { useRoute } from 'vue-router';
import { isEqual } from 'lodash';
import type { IBreadcrumb } from '@/typings';
import { breadcrumbSymbol } from '@/constants/provide-symbols';
import { RouteMetaConfig } from '@/router/meta';

export const provideBreadcrumb = () => {
  const route = useRoute();
  const data = reactive<IBreadcrumb>({
    title: '',
    display: false,
    back: true,
  });

  watch(
    () => route.meta,
    (meta: RouteMetaConfig, oldMeta: RouteMetaConfig) => {
      // 比较是为了防止 query 等变化产生路由更新时通过 setTitle 设置的 title 被覆盖
      if (!isEqual(meta, oldMeta)) {
        data.title = meta.menu?.i18n;
        data.display = meta.layout?.breadcrumb?.show !== false;
        data.back = meta.layout?.breadcrumb?.back ?? true;
      }
    },
    { deep: true },
  );

  provide(breadcrumbSymbol, data);
};

export default function useBreadcrumb() {
  const data = inject<IBreadcrumb>(breadcrumbSymbol);

  const setTitle = (newTitle: string) => {
    data.title = newTitle;
  };

  return {
    data,
    setTitle,
  };
}
