import { inject, provide, reactive, watch } from 'vue';
import { useRoute } from 'vue-router';
import type { IBreadcrumb } from '@/typings';
import { breadcrumbSymbol } from '@/constants/provide-symbols';
import { RouteMetaConfig } from '@/router/meta';

export const provideBreadcrumb = () => {
  const route = useRoute();
  const data = reactive<IBreadcrumb>({
    title: '',
    display: false,
  });

  watch(
    () => route.meta,
    (meta: RouteMetaConfig) => {
      data.title = meta.menu?.i18n;
      data.display = meta.layout?.breadcrumb?.show !== false;
    },
    { immediate: true, deep: true },
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
