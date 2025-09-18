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
  });

  watch(
    () => route.meta,
    (meta: RouteMetaConfig, oldMeta: RouteMetaConfig) => {
      // 比较是为了防止push等操作产生路由更新时通过setTitle设置的title被覆盖，这是目前比较经济的做法
      // 视之后的使用情况，如果比较不能满足所有场景可以考虑通过route.name判断或者重新赋值title时优先取当前data.title
      if (!isEqual(meta, oldMeta)) {
        data.title = meta.title;
        data.display = meta?.layout?.breadcrumbs?.show ?? meta.isShowBreadcrumb;
      }
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
