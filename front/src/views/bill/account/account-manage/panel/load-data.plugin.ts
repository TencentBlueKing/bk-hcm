import { useI18n } from 'vue-i18n';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { AccountLevelEnum } from '../constants';

import { firstAccountColumns, secondaryAccountColumns } from './columns';

export const getColumns = (accountLevel: AccountLevelEnum) => {
  const { t } = useI18n();

  let columns = firstAccountColumns.slice();
  if (accountLevel === AccountLevelEnum.SecondLevel) {
    const businessMapStore = useBusinessMapStore();

    columns = secondaryAccountColumns.slice();
    columns.splice(6, 0, {
      label: t('业务名称'),
      field: 'bk_biz_id',
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
    });
  }

  return columns;
};
