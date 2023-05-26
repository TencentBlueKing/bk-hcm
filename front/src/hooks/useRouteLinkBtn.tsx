import { VendorEnum } from "@/common/constant";
import { computed } from "vue";
import { useRouter, useRoute } from "vue-router";

export interface IDetail {
  vendor: VendorEnum;
  [key: string]: any;
}

export interface IMeta {
  id: string;
  type: TypeEnum;
  name: string;
  isExpand?: boolean; // 是否拓展网卡
}

export enum TypeEnum {
  HOST = 'vpc',
  SUBNET = 'subnet',
  ACCOUNT = 'account',
  IMAGE = 'image'
}

export const useRouteLinkBtn = (
  data: IDetail,
  meta: IMeta
) => {
  const router = useRouter();
  const route = useRoute();
  const { id, name, type, isExpand } = meta;
  const { vendor } = data;
  const computedId = computed(() => Array.isArray(data[id]) ? isExpand? data[name][1] : data[id][0] : data[id]);
  const computedName = computed(() => {
    let txt = Array.isArray(data[name]) ? isExpand? data[name][1] : data[name][0] : data[name];
    if(vendor === VendorEnum.AZURE && type === TypeEnum.HOST) txt = txt.split('/').reverse()[0];
    return txt;
  });

  const handleClick = () => {
    const routeInfo = {
      query: { id: computedId.value, type: vendor }
    };
    if (route.path.includes('business')) {
      Object.assign(
        routeInfo,
        {
          name: type === TypeEnum.ACCOUNT ? 'accountDetail' : `${type}BusinessDetail`,
        },
      );
    } else {
      Object.assign(
        routeInfo,
        {
          name: type === TypeEnum.ACCOUNT ? 'accountDetail' : 'resourceDetail',
          params: {
            type,
          },
        },
      );
    }
    router.push(routeInfo);
  }

  return (
    <bk-button text theme="primary" onClick={handleClick}>
      { computedName.value }
    </bk-button>
  )
}