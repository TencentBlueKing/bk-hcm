import { type RouteLocationNormalizedLoaded } from 'vue-router';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { IAccountItem } from '@/typings';
import { VendorEnum } from '@/common/constant';

export const accountFilter = (
  list: IAccountItem[],
  { route, whereAmI }: { route: RouteLocationNormalizedLoaded; whereAmI: ReturnType<typeof useWhereAmI> },
) => {
  const { isResourcePage } = whereAmI;
  // 负载均衡、目标组、证书托管、参数模板这四个暂时只腾讯云支持
  if (
    (isResourcePage && route.query.type === 'certs') ||
    route.query.scene === 'template' ||
    (route.meta.isFilterAccount as boolean)
  ) {
    return list.filter((item) => item.vendor === VendorEnum.TCLOUD);
  }
  return list;
};
