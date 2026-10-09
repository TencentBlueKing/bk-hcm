import business from '@/router/module/business';
import service from '@/router/module/service';
import serviceInside from '@/router/module/service-inside';
import resource from '@/router/module/resource';
import resourceInside from '@/router/module/resource-inside';
import scheme from '@/router/module/scheme';
import { billAccountManage, billAccountManageLegacy, billManage } from '@/views/bill/route-config';

export const businessViews = business.flatMap((group) => group.children);
export const serviceViews = [...service.flatMap((group) => group.children), ...serviceInside];
// TODO(菜单整合二期·组装)：资源接入迁入资源运营后删除
export const resourceAccessViews = [...resource, ...resourceInside];
export const schemeViews = scheme;
export const billViews = [...billAccountManage, ...billManage, ...billAccountManageLegacy];
