/**
 * 各资源类型的 search 属性定义，用于 useSearchQs.get 与 transformSimpleCondition
 */
import type { ModelPropertyGeneric, ModelPropertyType } from '@/model/typings';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { VENDORS } from '@/common/constant';
import { buildIPFilterRules, buildVIPFilterRules, buildMultipleValueRulesItem } from '@/utils/search';
import { QueryRuleOPEnum } from '@/typings';
import { MGMT_TYPE_MAP } from '@/constants/security-group';
import { LB_NETWORK_TYPE_MAP, CLB_STATUS_MAP } from '@/constants/clb';

const vendorOption = VENDORS.reduce((acc: Record<string, string>, cur) => {
  acc[cur.id] = cur.name;
  return acc;
}, {});

export const createProperty = (
  id: string,
  name: string,
  opts?: {
    type?: ModelPropertyType;
    op?: QueryRuleOPEnum;
    filterRules?: (value: any) => any;
    option?: Record<string, string>;
  },
): ModelPropertyGeneric => ({
  id,
  name,
  type: opts?.type || 'string',
  option: opts?.option,
  meta:
    opts?.filterRules || opts?.op
      ? { search: { ...(opts.filterRules && { filterRules: opts.filterRules }), ...(opts.op && { op: opts.op }) } }
      : undefined,
});

const baseProperties: ModelPropertyGeneric[] = [
  createProperty('name', '名称', { op: QueryRuleOPEnum.CS }),
  createProperty('vendor', '云厂商', { option: vendorOption, op: QueryRuleOPEnum.IN }),
  createProperty('account_id', '云账号ID', { type: 'string' }),
  createProperty('cloud_id', '资源ID', { type: 'string' }),
];

const cvmProperties: ModelPropertyGeneric[] = [
  createProperty('private_ip', '内网IP', {
    filterRules: (value) => buildIPFilterRules(value, 'private'),
  }),
  createProperty('public_ip', '公网IP', {
    filterRules: (value) => buildIPFilterRules(value, 'public'),
  }),
  createProperty('cloud_id', '主机ID'),
  createProperty('bk_asset_id', '固资号', { type: 'string' }),
  ...baseProperties,
  createProperty('bk_cloud_id', '管控区域', { type: 'number' }),
  createProperty('os_name', '操作系统', { op: QueryRuleOPEnum.CS }),
  createProperty('cloud_vpc_ids', '所属VPC', {
    filterRules: (value) => ({ field: 'cloud_vpc_ids', op: QueryRuleOPEnum.JSON_CONTAINS, value }),
  }),
];

const imageProperties: ModelPropertyGeneric[] = [
  createProperty('cloud_id', '镜像ID'),
  createProperty('name', '名称', { op: QueryRuleOPEnum.CS }),
  createProperty('vendor', '云厂商', { option: vendorOption, op: QueryRuleOPEnum.IN }),
];

const subnetProperties: ModelPropertyGeneric[] = [...baseProperties, createProperty('cloud_vpc_id', '所属VPC ID')];

const networkInterfaceProperties: ModelPropertyGeneric[] = [
  ...baseProperties,
  createProperty('public_ipv4', '公网ipv4'),
  createProperty('private_ipv4', '内网ipv4'),
];

/**
 * CLB 属性定义
 * lb_vip 需要 buildVIPFilterRules；lb_type / ip_version / status 用 EQ（枚举精确匹配）
 */
const clbProperties: ModelPropertyGeneric[] = [
  createProperty('name', '负载均衡名称', { op: QueryRuleOPEnum.CS }),
  createProperty('cloud_id', '负载均衡ID'),
  createProperty('domain', '负载均衡域名', { op: QueryRuleOPEnum.CS }),
  createProperty('lb_vip', '负载均衡VIP', {
    filterRules: (value) => buildVIPFilterRules(value),
  }),
  createProperty('lb_type', '网络类型', {
    option: LB_NETWORK_TYPE_MAP,
    filterRules: (value) => ({ field: 'lb_type', op: QueryRuleOPEnum.EQ, value }),
  }),
  createProperty('ip_version', 'IP版本', {
    option: { ipv4: 'IPv4', ipv6: 'IPv6', ipv6_dual_stack: 'IPv6DualStack', ipv6_nat64: 'IPv6Nat64' },
    filterRules: (value) => ({ field: 'ip_version', op: QueryRuleOPEnum.EQ, value }),
  }),
  createProperty('vendor', '云厂商', { option: vendorOption, op: QueryRuleOPEnum.IN }),
  createProperty('zones', '可用区域', { op: QueryRuleOPEnum.CS }),
  createProperty('status', '状态', {
    option: CLB_STATUS_MAP,
    filterRules: (value) => ({ field: 'status', op: QueryRuleOPEnum.EQ, value }),
  }),
  createProperty('cloud_vpc_id', '所属VPC'),
  createProperty('region', '地域'),
  createProperty('account_id', '云账号ID', { type: 'string' }),
];

/**
 * 安全组属性定义（group 子类型）
 * usage_biz_id / mgmt_biz_id / mgmt_type 使用 IN/EQ；cloud_id 使用 buildMultipleValueRulesItem
 */
const securityGroupProperties: ModelPropertyGeneric[] = [
  createProperty('cloud_id', '安全组ID', {
    filterRules: (value) => buildMultipleValueRulesItem('cloud_id', value),
  }),
  createProperty('name', '名称', { op: QueryRuleOPEnum.CS }),
  createProperty('vendor', '云厂商', { option: vendorOption, op: QueryRuleOPEnum.IN }),
  createProperty('account_id', '云账号ID', { type: 'string' }),
  createProperty('usage_biz_id', '使用业务', {
    filterRules: (value) => ({
      field: 'bk_biz_id',
      op: Array.isArray(value) ? QueryRuleOPEnum.IN : QueryRuleOPEnum.EQ,
      value,
    }),
  }),
  createProperty('mgmt_type', '管理类型', {
    option: MGMT_TYPE_MAP,
    filterRules: (value) => ({
      field: 'mgmt_type',
      op: Array.isArray(value) ? QueryRuleOPEnum.IN : QueryRuleOPEnum.EQ,
      value,
    }),
  }),
  createProperty('mgmt_biz_id', '管理业务', {
    filterRules: (value) => ({
      field: 'mgmt_biz_id',
      op: Array.isArray(value) ? QueryRuleOPEnum.IN : QueryRuleOPEnum.EQ,
      value,
    }),
  }),
  createProperty('region', '地域'),
];

const resourcePropertiesMap: Partial<Record<ResourceTypeEnum, ModelPropertyGeneric[]>> = {
  [ResourceTypeEnum.CVM]: cvmProperties,
  [ResourceTypeEnum.VPC]: baseProperties,
  [ResourceTypeEnum.SUBNET]: subnetProperties,
  [ResourceTypeEnum.DISK]: baseProperties,
  [ResourceTypeEnum.EIP]: [
    createProperty('public_ip', '公网IP', {
      filterRules: (value) => buildIPFilterRules(value, 'public'),
    }),
    ...baseProperties,
  ],
  [ResourceTypeEnum.IMAGE]: imageProperties,
  [ResourceTypeEnum.NETWORK_INTERFACE]: networkInterfaceProperties,
  [ResourceTypeEnum.ROUTING]: baseProperties,
  [ResourceTypeEnum.CLB]: clbProperties,
  [ResourceTypeEnum.SECURITY_GROUP]: securityGroupProperties,
};

export const getSearchProperties = (resourceType: ResourceTypeEnum): ModelPropertyGeneric[] => {
  return resourcePropertiesMap[resourceType] || baseProperties;
};
