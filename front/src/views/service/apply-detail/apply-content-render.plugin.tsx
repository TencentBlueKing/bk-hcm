import { Ref, VNode } from 'vue';
import { ACCOUNT_TYPES } from '../apply-list/constants';

import AccountApplyDetail from './account-apply-detail';
import ApplyDetail from '@/views/service/my-apply/components/apply-detail/index.vue';

export const applyContentRender: (...args: any) => VNode = (
  currentApplyData: Ref<any>,
  curApplyKey: Ref<string>,
  applyDetailProps: any,
) => {
  if (ACCOUNT_TYPES.includes(currentApplyData.value.type)) {
    return <AccountApplyDetail detail={currentApplyData.value} />;
  }
  return <ApplyDetail params={currentApplyData.value} key={curApplyKey.value} {...applyDetailProps} />;
};
