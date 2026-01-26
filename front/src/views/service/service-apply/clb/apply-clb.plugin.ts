import { Message } from 'bkui-vue';
import routerAction from '@/router/utils/action';
import { MENU_SERVICE_TICKET_MANAGEMENT } from '@/constants/menu-symbol';

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export const applyClbSuccessHandler = (isBusinessPage: boolean, goBack: () => void, args?: { bizId: number }) => {
  Message({ theme: 'success', message: '购买成功' });
  if (isBusinessPage) {
    routerAction.redirect({ name: MENU_SERVICE_TICKET_MANAGEMENT, query: { type: 'load_balancer' } });
  } else {
    goBack();
  }
};
