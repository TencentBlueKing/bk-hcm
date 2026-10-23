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
      // oldMeta 为 undefined 时是首次调用（immediate），必须初始化；
      // 后续变更通过 isEqual 比较，防止 query 变化时覆盖 setTitle 设置的 title
      if (!oldMeta || !isEqual(meta, oldMeta)) {
        data.title = meta.menu?.i18n;
        data.display = meta.layout?.breadcrumb?.show !== false;
        data.back = meta.layout?.breadcrumb?.back ?? true;
      }
    },
    { deep: true, immediate: true },
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
