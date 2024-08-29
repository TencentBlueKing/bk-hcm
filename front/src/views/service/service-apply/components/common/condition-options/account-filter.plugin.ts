import { type RouteLocationNormalizedLoaded } from 'vue-router';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { IAccountItem } from '@/typings';
import { VendorEnum } from '@/common/constant';

export const accountFilter = (
  list: IAccountItem[],
  { route, whereAmI }: { route: RouteLocationNormalizedLoaded; whereAmI: ReturnType<typeof useWhereAmI> },
) => {
  const { isResourcePage, isBusinessPage } = whereAmI;
  if (
    (isResourcePage && route.query.type === 'certs') ||
    (isBusinessPage && route.path.includes('cert')) ||
    ['lb', 'targetGroup'].includes(route.meta.applyRes as string)
  ) {
    return list.filter((item) => item.vendor === VendorEnum.TCLOUD);
  }
  return list;
};
