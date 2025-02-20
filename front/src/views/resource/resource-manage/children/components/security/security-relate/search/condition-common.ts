import { SecurityGroupRelatedResourceName } from '@/store/security-group';
import { RELATED_RES_PROPERTIES_MAP } from '@/constants/security-group';

const conditionFieldIds = new Map<string, string[]>();

const baseFieldIds = ['name'];
const cvmBaseFields = ['private_ipv4_addresses', 'region', ...baseFieldIds];
const cvmBindFields = ['private_ipv4_addresses', 'cloud_vpc_ids', ...baseFieldIds];
// TODO: 业务搜索需要format一下
const cvmUnbindFields = [...cvmBindFields];

conditionFieldIds.set('CVM-base', cvmBaseFields);
conditionFieldIds.set('CVM-bind', cvmBindFields);
conditionFieldIds.set('CVM-unbind', cvmUnbindFields);

export const getConditionFieldIds = (key: string) => {
  return conditionFieldIds.get(key);
};

const getConditionField = (resourceName: SecurityGroupRelatedResourceName, operation: string) => {
  const key = `${resourceName}-${operation}`;
  const fieldIds = getConditionFieldIds(key);
  const fields = fieldIds.map((id) => RELATED_RES_PROPERTIES_MAP[resourceName].find((item) => item.id === id));
  return fields;
};

const factory = {
  getConditionField,
};

export type FactoryType = typeof factory;

export default factory;
