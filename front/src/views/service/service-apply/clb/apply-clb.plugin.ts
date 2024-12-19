import { Message } from 'bkui-vue';
import routerAction from '@/router/utils/action';
import { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';

export const applyClbSuccessHandler = (isBusinessPage: boolean, goBack: () => void, _formModel?: ApplyClbModel) => {
  Message({ theme: 'success', message: '购买成功' });
  if (isBusinessPage) {
    // 业务下购买CLB, 跳转至我的单据
    routerAction.redirect({ path: '/service/my-apply' });
  } else goBack();
};
