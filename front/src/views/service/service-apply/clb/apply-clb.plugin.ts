import { Message } from 'bkui-vue';
import routerAction from '@/router/utils/action';
import { MENU_SERVICE_TICKET_MANAGEMENT } from '@/constants/menu-symbol';

export const applyClbSuccessHandler = () => {
  Message({ theme: 'success', message: '购买成功' });
  routerAction.redirect({ name: MENU_SERVICE_TICKET_MANAGEMENT, query: { type: 'load_balancer' } });
};
