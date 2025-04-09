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
      data.title = meta.title;
      data.display = meta?.layout?.breadcrumbs?.show ?? meta.isShowBreadcrumb;
    },
    { deep: true },
  );

  provide(breadcrumbSymbol, data);
};

export default function useBreadcrumb() {
  const breadcrumb = inject<IBreadcrumb>(breadcrumbSymbol);

  const setTitle = (newTitle: string) => {
    breadcrumb.title = newTitle;
  };

  return {
    breadcrumb,
    setTitle,
  };
}
