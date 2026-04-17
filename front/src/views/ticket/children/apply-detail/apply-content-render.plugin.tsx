import { Ref, VNode } from 'vue';
import { ACCOUNT_TYPES } from '@/views/ticket/constants';
import CommonApplyDetail from './common-apply-detail/index.vue';
import AccountApplyDetail from './account-apply-detail';

export const applyContentRender: (...args: any) => VNode = (
  currentApplyData: Ref<any>,
  curApplyKey: Ref<string>,
  applyDetailProps: any,
) => {
  if (ACCOUNT_TYPES.includes(currentApplyData.value.operation)) {
    return <AccountApplyDetail detail={currentApplyData.value} />;
  }
  return <CommonApplyDetail details={currentApplyData.value} key={curApplyKey.value} {...applyDetailProps} />;
};
