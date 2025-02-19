import { PropertyColumnConfig } from '@/model/typings';
import { getInstVip, getPrivateIPs, getPublicIPs } from '@/utils';
import { IResourceBoundSecurityGroupItem } from '@/store/security-group';
import securityGroupRelatedResourcesViewProperties from '@/model/security-group/related-resources.view';

const columnIds = new Map<string, string[]>();

const columnConfig: Record<string, PropertyColumnConfig> = {
  private_ip: {
    render: ({ data }: any) => getPrivateIPs(data),
  },
  public_ip: {
    render: ({ data }: any) => getPublicIPs(data),
  },
  vip: {
    render: ({ data }: any) => getInstVip(data),
  },
  security_group_names: {
    render: ({ data }: any) =>
      data?.security_groups
        ?.flatMap((item: IResourceBoundSecurityGroupItem['security_groups'][1]) => item.name)
        ?.join(',') || '--',
  },
};

const relCvmFields = ['private_ip', 'public_ip', 'region', 'zone', 'name', 'status', 'cloud_vpc_ids'];
const relClbFields = ['name', 'domain', 'vip', 'lb_type', 'ip_version', 'region', 'zones', 'status', 'cloud_vpc_id'];
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

const getColumns = (key: string) => {
  const columnIds = getColumnIds(key);
  return columnIds.map((id) => ({
    ...securityGroupRelatedResourcesViewProperties.find((item) => item.id === id),
    ...columnConfig[id],
  }));
};

const factory = {
  getColumns,
};

export type FactoryType = typeof factory;

export default factory;
