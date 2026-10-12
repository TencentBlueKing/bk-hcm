import { GLOBAL_BIZS_KEY, VendorEnum } from '@/common/constant';
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import {
  MENU_BUSINESS_IMAGE_DETAILS,
  MENU_BUSINESS_SUBNET_DETAILS,
  MENU_BUSINESS_VPC_DETAILS,
  MENU_RESOURCE_ACCOUNT_DETAIL,
  MENU_RESOURCE_DETAIL,
} from '@/constants/menu-symbol';
import routerAction from '@/router/utils/action';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

export interface IDetail {
  vendor: VendorEnum;
  [key: string]: any;
}

export interface IMeta {
  id: string; // IDetail[id] 作为跳转链接参数 id 的值
  type: TypeEnum; // 业务下跳业务详情 /business/:type/detail/:id，资源下跳资源纳管详情 /resource/manage/:resourceType/details/:id
  name: string; // IDetail[name] 作为按钮要显示的文字内容
  isExpand?: boolean; // 是否拓展网卡，当存在拓展网卡时，网络、子网、公私IPV4、公私IPV6都是2份，此时 IDetail[id] 和 IDetail[name] 都是一个长度为2的数组, 需要特殊处理
}

export enum TypeEnum {
  VPC = 'vpc',
  SUBNET = 'subnet',
  ACCOUNT = 'account',
  IMAGE = 'image',
}

const BUSINESS_DETAIL_ROUTE_MAP: Partial<Record<TypeEnum, symbol>> = {
  [TypeEnum.VPC]: MENU_BUSINESS_VPC_DETAILS,
  [TypeEnum.SUBNET]: MENU_BUSINESS_SUBNET_DETAILS,
  [TypeEnum.IMAGE]: MENU_BUSINESS_IMAGE_DETAILS,
};

export const useRouteLinkBtn = (data: IDetail, meta: IMeta) => {
  const router = useRouter();
  const route = useRoute();
  const { whereAmI } = useWhereAmI();
  const { id, name, type, isExpand } = meta;
  const { vendor } = data;
  // eslint-disable-next-line no-nested-ternary
  const computedId = computed(() => (Array.isArray(data[id]) ? (isExpand ? data[name][1] : data[id][0]) : data[id]));
  const computedName = computed(() => {
    // eslint-disable-next-line no-nested-ternary
    let txt = Array.isArray(data[name]) ? (isExpand ? data[name][1] : data[name][0]) : data[name];
    // eslint-disable-next-line prefer-destructuring
    if (vendor === VendorEnum.AZURE && type === TypeEnum.VPC) txt = txt.split('/').reverse()[0];
    return txt;
  });

  const handleClick = () => {
    // 账号详情在资源运营 → 账号管理下，业务、资源两侧同一地址
    if (type === TypeEnum.ACCOUNT) {
      router.push({ name: MENU_RESOURCE_ACCOUNT_DETAIL, params: { accountId: computedId.value } });
      return;
    }
    const routeInfo: any = {};
    if (whereAmI.value === Senarios.business) {
      Object.assign(routeInfo, {
        name: BUSINESS_DETAIL_ROUTE_MAP[type],
        params: { id: computedId.value },
        query: { [GLOBAL_BIZS_KEY]: route.query[GLOBAL_BIZS_KEY], vendor },
      });
    } else {
      Object.assign(routeInfo, {
        name: MENU_RESOURCE_DETAIL,
        params: { resourceType: type, id: computedId.value },
        query: { accountId: route.query.accountId, vendor },
      });
    }
    routerAction.redirect(routeInfo, { history: true });
  };

  const render = () => {
    if (!computedName.value) return '--';
    return (
      <bk-button text theme='primary' onClick={handleClick}>
        {computedName.value}
      </bk-button>
    );
  };

  return render();
};
