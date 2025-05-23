import { h } from 'vue';
import { PropertyColumnConfig } from '@/model/typings';
import { getInstVip, getPrivateIPs, getPublicIPs } from '@/utils';
import { IResourceBoundSecurityGroupItem, SecurityGroupRelatedResourceName } from '@/store/security-group';
import { RELATED_RES_PROPERTIES_MAP } from '@/constants/security-group';

import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

const columnIds = new Map<string, string[]>();

const columnConfig: Record<string, PropertyColumnConfig> = {
  private_ip: {
    render: ({ data }: any) => {
      const content = getPrivateIPs(data);
      return h('div', { class: 'flex-row align-items-center' }, [
        content,
        content !== '--' ? h(CopyToClipboard, { content, class: 'ml4' }) : null,
      ]);
    },
  },
  public_ip: {
    render: ({ data }: any) => {
      const content = getPublicIPs(data);
      return h('div', { class: 'flex-row align-items-center' }, [
        content,
        content !== '--' ? h(CopyToClipboard, { content, class: 'ml4' }) : null,
      ]);
    },
  },
  lb_vip: {
    render: ({ data }: any) => getInstVip(data),
  },
  security_group_names: {
    render: ({ data }: any) =>
      data?.security_groups
        ?.flatMap((item: IResourceBoundSecurityGroupItem['security_groups'][1]) => item.name)
        ?.join(',') || '--',
  },
};

const relCvmFields = ['private_ip', 'public_ip', 'region', 'zone', 'name', 'status', 'cloud_vpc_ids', 'bk_biz_id'];
const relClbFields = [
  'name',
  'domain',
  'lb_vip',
  'lb_type',
  'ip_version',
  'region',
  'zones',
  'status',
  'cloud_vpc_id',
  'bk_biz_id',
];
const bindCvmFields = ['private_ip', 'public_ip', 'name', 'cloud_vpc_ids', 'status', 'security_group_names'];
const unbindCvmFields = [
  'private_ip',
  'public_ip',
  'name',
  'cloud_vpc_ids',
  'status',
  'bk_biz_id',
  'security_group_names',
];

columnIds.set('CVM-base', relCvmFields);
columnIds.set('CLB-base', relClbFields);

columnIds.set('CVM-bind', bindCvmFields);
columnIds.set('CVM-unbind', unbindCvmFields);

export const getColumnIds = (key: string) => {
  return columnIds.get(key);
};

const getColumns = (resourceName: SecurityGroupRelatedResourceName, operation: string) => {
  const key = `${resourceName}-${operation}`;
  const columnIds = getColumnIds(key);
  return columnIds.map((id) => ({
    ...RELATED_RES_PROPERTIES_MAP[resourceName].find((item) => item.id === id),
    ...columnConfig[id],
  }));
};

const factory = {
  getColumns,
};

export type FactoryType = typeof factory;

export default factory;
