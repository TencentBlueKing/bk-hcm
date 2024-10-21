import { AccountLevelEnum } from '../constants';

import { firstAccountColumns, secondaryAccountColumns } from './columns';
import { injectBizField } from '@/utils';

export const getColumns = (accountLevel: AccountLevelEnum) => {
  let columns = firstAccountColumns.slice();
  if (accountLevel === AccountLevelEnum.SecondLevel) {
    columns = secondaryAccountColumns.slice();
    injectBizField(columns, 6);
  }

  return columns;
};
