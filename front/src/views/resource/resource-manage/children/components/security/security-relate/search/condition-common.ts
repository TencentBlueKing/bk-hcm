import { RELATED_RES_PROPERTIES_MAP, SecurityGroupRelatedResourceName } from '@/constants/security-group';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { CLB_STATUS_MAP, LB_NETWORK_TYPE_MAP } from '@/constants';
import { VendorEnum, VendorMap } from '@/common/constant';

const conditionFieldIds = new Map<string, string[]>();

// 默认配置
const cvmConditionConfig: Record<string, Partial<ISearchItem>> = {};
const clbConditionConfig: Record<string, Partial<ISearchItem>> = {
  lb_type: {
    children: Object.keys(LB_NETWORK_TYPE_MAP).map((lbType) => ({
      id: lbType,
      name: LB_NETWORK_TYPE_MAP[lbType],
    })),
  },
  ip_version: {
    children: [
      { id: 'ipv4', name: 'IPv4' },
      { id: 'ipv6', name: 'IPv6' },
      { id: 'ipv6_dual_stack', name: 'IPv6DualStack' },
      { id: 'ipv6_nat64', name: 'IPv6Nat64' },
    ],
  },
  vendor: {
    children: [{ id: VendorEnum.TCLOUD, name: VendorMap[VendorEnum.TCLOUD] }],
  },
  status: {
    children: Object.keys(CLB_STATUS_MAP).map((key) => ({ id: key, name: CLB_STATUS_MAP[key] })),
  },
};
const configMap = {
  [SecurityGroupRelatedResourceName.CVM]: cvmConditionConfig,
  [SecurityGroupRelatedResourceName.CLB]: clbConditionConfig,
};

const cvmBaseFields = ['private_ip', 'name', 'bk_biz_id'];
const cvmBindFields = ['private_ip', 'cloud_vpc_ids', 'name'];
const cvmUnbindFields = [...cvmBindFields];
const clbBaseFields = [
  'name',
  'domain',
  'lb_vip',
  'lb_type',
  'ip_version',
  'vendor',
  'zones',
  'status',
  'cloud_vpc_id',
  'bk_biz_id',
];

conditionFieldIds.set('CVM-base', cvmBaseFields);
conditionFieldIds.set('CVM-bind', cvmBindFields);
conditionFieldIds.set('CVM-unbind', cvmUnbindFields);
conditionFieldIds.set('CLB-base', clbBaseFields);

export const getConditionFieldIds = (key: string) => {
  return conditionFieldIds.get(key);
};

const getConditionField = (resourceName: SecurityGroupRelatedResourceName, operation: string) => {
  const key = `${resourceName}-${operation}`;
  const fieldIds = getConditionFieldIds(key);

  const fields = fieldIds.map((id) => ({
    ...RELATED_RES_PROPERTIES_MAP[resourceName].find((item) => item.id === id),
    ...configMap[resourceName][id],
  }));
  return fields;
};

const factory = {
  getConditionField,
};

export type FactoryType = typeof factory;

export default factory;
