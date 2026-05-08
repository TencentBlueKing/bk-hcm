import type { IPermissionTemplateItem } from '@/store/cloud-account-manage/permission-template';
import { TagThemeEnum } from 'bkui-vue/lib/shared';

export const getTypeData = (row: IPermissionTemplateItem) => {
  const cloudType = row.extension?.cloud_type;
  const data = {
    label: '--',
    theme: TagThemeEnum.UNKNOWN,
    isCloudCustom: false,
    isCloudPreset: false,
    isPolicySync: false,
  };
  if (cloudType === 1) {
    if (row.policy_library_id) {
      data.label = '与策略库同步';
      data.theme = TagThemeEnum.SUCCESS;
      data.isPolicySync = true;
    } else {
      data.label = '云自定义';
      data.theme = TagThemeEnum.WARNING;
      data.isCloudCustom = true;
    }
  } else if (cloudType === 2) {
    data.label = '云系统预设';
    data.theme = TagThemeEnum.UNKNOWN;
    data.isCloudPreset = true;
  } else {
    data.label = '--';
    data.theme = TagThemeEnum.UNKNOWN;
  }
  return data;
};
