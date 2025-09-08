import { Message } from 'bkui-vue';
import routerAction from '@/router/utils/action';
import { MENU_SERVICE_TICKET_DETAILS } from '@/constants/menu-symbol';

export const applyClbSuccessHandler = (isBusinessPage: boolean, goBack: () => void, args: any) => {
  Message({ theme: 'success', message: '购买成功' });
  const { id } = args || {};
  if (isBusinessPage && id) {
    // 业务下购买CLB, 跳转至单据详情
    routerAction.redirect({ name: MENU_SERVICE_TICKET_DETAILS, query: { id } });
  } else {
    goBack();
  }
};
