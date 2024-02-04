import { VendorEnum } from '@/common/constant';
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';

export interface IDetail {
  vendor: VendorEnum;
  [key: string]: any;
}

export interface IMeta {
  id: string; // IDetail[id] 作为跳转链接参数 id 的值
  type: TypeEnum; // 如果是业务，将跳转到 ${type}BusinessDetail; 如果是资源，跳转到resource/detail/:type; 这里强依赖于 router 配置文件的路由命名规范
  name: string; // IDetail[name] 作为按钮要显示的文字内容
  isExpand?: boolean; // 是否拓展网卡，当存在拓展网卡时，网络、子网、公私IPV4、公私IPV6都是2份，此时 IDetail[id] 和 IDetail[name] 都是一个长度为2的数组, 需要特殊处理
}

export enum TypeEnum {
  VPC = 'vpc',
  SUBNET = 'subnet',
  ACCOUNT = 'account',
  IMAGE = 'image',
}

export const useRouteLinkBtn = (data: IDetail, meta: IMeta) => {
  const router = useRouter();
  const route = useRoute();
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
    const routeInfo = {
      query: {
        ...route.query,
        id: computedId.value,
        type: vendor,
      },
    };
    if (route.path.includes('business')) {
      Object.assign(routeInfo, {
        name: type === TypeEnum.ACCOUNT ? 'accountDetail' : `${type}BusinessDetail`,
      });
    } else {
      Object.assign(routeInfo, {
        name: type === TypeEnum.ACCOUNT ? 'accountDetail' : 'resourceDetail',
        params: {
          type,
        },
      });
    }
    if (id === 'account_id') {
      Object.assign(routeInfo.query, { accountId: computedId.value });
    }
    router.push(routeInfo);
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
