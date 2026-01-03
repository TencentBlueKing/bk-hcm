import { Message } from 'bkui-vue';
import routerAction from '@/router/utils/action';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { MENU_BUSINESS_TICKET_MANAGEMENT } from '@/constants/menu-symbol';

export const applyClbSuccessHandler = (isBusinessPage: boolean, goBack: () => void, args?: { bizId: number }) => {
  Message({ theme: 'success', message: '购买成功' });
  if (isBusinessPage) {
    // 业务下购买CLB, 跳转至单据管理-负载均衡
    routerAction.redirect({
      name: MENU_BUSINESS_TICKET_MANAGEMENT,
      query: {
        [GLOBAL_BIZS_KEY]: args?.bizId,
        type: 'load_balancer',
      },
    });
  } else {
    goBack();
  }
};
