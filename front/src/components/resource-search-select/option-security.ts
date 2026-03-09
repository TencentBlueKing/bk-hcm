/**
 * 安全组 / GCP 防火墙 / 参数模板 的搜索选项配置
 */
import type { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { VENDORS } from '@/common/constant';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useRegionStore } from '@/store/region';
import { SecurityGroupManageType, MGMT_TYPE_MAP } from '@/constants/security-group';
import { getAccountList } from './option-common';

export type SecuritySubType = 'group' | 'gcp' | 'template';

const cloudIdLabelMap: Record<SecuritySubType, string> = {
  group: '安全组ID',
  gcp: '防火墙ID',
  template: '模板ID',
};

const getSecurityOptions = (subType: SecuritySubType = 'group'): ISearchItem[] => {
  const businessGlobalStore = useBusinessGlobalStore();
  const businessChildren = () => businessGlobalStore.businessFullList.map(({ id, name }) => ({ id: String(id), name }));

  const cloudId: ISearchItem = { name: cloudIdLabelMap[subType] || '资源ID', id: 'cloud_id' };
  const name: ISearchItem = { name: '名称', id: 'name' };
  const vendor: ISearchItem = { name: '云厂商', id: 'vendor', multiple: true, children: VENDORS };
  const accountId: ISearchItem = { name: '云账号ID', id: 'account_id', async: true, multiple: true, children: [] };

  // group 额外字段
  const groupExtra: ISearchItem[] = [
    {
      name: '使用业务',
      id: 'usage_biz_id',
      multiple: true,
      children: businessChildren(),
    },
    {
      name: '管理类型',
      id: 'mgmt_type',
      multiple: true,
      children: [
        { id: SecurityGroupManageType.BIZ, name: MGMT_TYPE_MAP[SecurityGroupManageType.BIZ] },
        { id: SecurityGroupManageType.PLATFORM, name: MGMT_TYPE_MAP[SecurityGroupManageType.PLATFORM] },
        { id: SecurityGroupManageType.UNKNOWN, name: MGMT_TYPE_MAP[SecurityGroupManageType.UNKNOWN] },
      ],
    },
    {
      name: '管理业务',
      id: 'mgmt_biz_id',
      multiple: true,
      children: businessChildren(),
    },
    {
      name: '地域',
      id: 'region',
      async: true,
      placeholder: '请输入地域名',
    },
  ];

  switch (subType) {
    case 'group':
      return [cloudId, name, vendor, accountId, ...groupExtra];
    case 'gcp':
      // GCP 防火墙不需要 vendor（固定为 GCP）
      return [cloudId, name, accountId];
    case 'template':
      return [cloudId, name, vendor, accountId];
    default:
      return [cloudId, name, vendor, accountId];
  }
};

const getSecurityMenuList = async (item: ISearchItem, keyword: string): Promise<any[]> => {
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
  getOptionData: getSecurityOptions,
  getOptionMenu: getSecurityMenuList,
};
