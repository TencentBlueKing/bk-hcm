/**
 * 负载均衡 (CLB) 的搜索选项配置
 */
import type { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { VendorEnum, VendorMap } from '@/common/constant';
import { LB_NETWORK_TYPE_MAP, CLB_STATUS_MAP } from '@/constants/clb';
import { useRegionStore } from '@/store/region';
import { getAccountList } from './option-common';

const getClbOptions = (): ISearchItem[] => {
  return [
    { id: 'name', name: '负载均衡名称' },
    { id: 'cloud_id', name: '负载均衡ID' },
    { id: 'domain', name: '负载均衡域名' },
    { id: 'lb_vip', name: '负载均衡VIP' },
    {
      id: 'lb_type',
      name: '网络类型',
      children: Object.keys(LB_NETWORK_TYPE_MAP).map((key) => ({
        id: key,
        name: LB_NETWORK_TYPE_MAP[key],
      })),
    },
    {
      id: 'ip_version',
      name: 'IP版本',
      children: [
        { id: 'ipv4', name: 'IPv4' },
        { id: 'ipv6', name: 'IPv6' },
        { id: 'ipv6_dual_stack', name: 'IPv6DualStack' },
        { id: 'ipv6_nat64', name: 'IPv6Nat64' },
      ],
    },
    {
      id: 'vendor',
      name: '云厂商',
      children: [{ id: VendorEnum.TCLOUD, name: VendorMap[VendorEnum.TCLOUD] }],
    },
    { id: 'zones', name: '可用区域' },
    {
      id: 'status',
      name: '状态',
      children: Object.keys(CLB_STATUS_MAP).map((key) => ({ id: key, name: CLB_STATUS_MAP[key] })),
    },
    { id: 'cloud_vpc_id', name: '所属VPC' },
    {
      id: 'region',
      name: '地域',
      async: true,
      placeholder: '请输入地域名',
    },
    {
      id: 'account_id',
      name: '云账号ID',
      async: true,
      multiple: true,
      children: [],
    },
  ];
};

const getClbMenuList = async (item: ISearchItem, keyword: string): Promise<any[]> => {
  const { id, async: isAsync, children = [] } = item;
  if (!isAsync) return children;

  if (id === 'account_id') {
    return getAccountList(keyword);
  }

  if (id === 'region') {
    const { getAllVendorRegion } = useRegionStore();
    return getAllVendorRegion(keyword);
  }

  return children;
};

export default {
  getOptionData: getClbOptions,
  getOptionMenu: getClbMenuList,
};
