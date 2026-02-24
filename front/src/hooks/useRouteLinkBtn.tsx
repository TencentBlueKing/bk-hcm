import { VendorEnum } from '@/common/constant';
import {
  MENU_BUSINESS_VPC_DETAILS,
  MENU_BUSINESS_SUBNET_DETAILS,
  MENU_BUSINESS_IMAGE_DETAILS,
  MENU_RESOURCE_DETAIL,
  MENU_SERVICE_ACCOUNT_DETAIL,
} from '@/constants/menu-symbol';
import routerAction from '@/router/utils/action';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { computed } from 'vue';

export interface IDetail {
  vendor: VendorEnum;
  [key: string]: any;
}

export interface IMeta {
  id: string;
  type: TypeEnum;
  name: string;
  isExpand?: boolean;
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
    if (type === TypeEnum.ACCOUNT) {
      routerAction.redirect(
        { name: MENU_SERVICE_ACCOUNT_DETAIL, params: { accountId: computedId.value } },
        { history: true },
      );
      return;
    }

    const isBusiness = whereAmI.value === Senarios.business;
    const routeInfo: any = { query: { type: vendor } };

    if (isBusiness) {
      Object.assign(routeInfo, {
        name: BUSINESS_DETAIL_ROUTE_MAP[type],
        params: { id: computedId.value },
      });
    } else {
      Object.assign(routeInfo, {
        name: MENU_RESOURCE_DETAIL,
        params: { resourceType: type, id: computedId.value },
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
