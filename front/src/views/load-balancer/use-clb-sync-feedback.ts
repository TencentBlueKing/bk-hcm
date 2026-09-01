import { h } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { ResourceTypeEnum } from '@/common/constant';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';
import routerAction from '@/router/utils/action';
import { useWhereAmI } from '@/hooks/useWhereAmI';

export interface IClbSyncResponse {
  code?: number;
  message?: string;
  data?: Record<string, unknown>;
}

export function useClbSyncFeedback(getBizId?: () => number) {
  const { t } = useI18n();
  const { getBizsId } = useWhereAmI();

  const handleClbSyncError = (error: IClbSyncResponse) => {
    const message = error?.code === 2000002 ? t('该账号地域同步任务进行中') : error?.message || t('同步失败');
    Message({ theme: 'error', message });
  };

  const handleClbSyncSuccess = (res: IClbSyncResponse) => {
    const taskManagementId = res.data?.task_management_id;
    const taskId = typeof taskManagementId === 'string' ? taskManagementId : '';
    if (!taskId) {
      Message({ theme: 'warning', message: t('没有可处理的负载均衡') });
      return;
    }

    const openTaskDetail = () => {
      routerAction.open({
        name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
        query: { bizs: getBizId?.() ?? getBizsId() },
        params: { resourceType: ResourceTypeEnum.CLB, id: taskId },
      });
    };

    Message({
      theme: 'success',
      delay: 8000,
      message: h('span', [
        t('同步任务已创建，可在'),
        h(
          'span',
          {
            style: { color: '#3a84ff', cursor: 'pointer' },
            onClick: openTaskDetail,
          },
          t('【任务管理-负载均衡】'),
        ),
        t('查看进度'),
      ]),
    });
  };

  return { handleClbSyncSuccess, handleClbSyncError };
}
