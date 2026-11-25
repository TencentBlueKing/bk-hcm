import { h } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { ResourceTypeEnum } from '@/common/constant';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';
import routerAction from '@/router/utils/action';
import { useWhereAmI } from '@/hooks/useWhereAmI';

export interface IClbTaskFeedbackResponse {
  code?: number;
  message?: string;
  data?: {
    task_management_id?: string;
    [key: string]: unknown;
  };
}

export function useClbTaskFeedback(getBizId?: () => number) {
  const { t } = useI18n();
  const { getBizsId } = useWhereAmI();

  const getTaskId = (res?: IClbTaskFeedbackResponse) => {
    const taskManagementId = res?.data?.task_management_id;
    return typeof taskManagementId === 'string' ? taskManagementId : '';
  };

  const openClbTaskDetail = (taskId: string) => {
    const rawBizId = getBizId?.() ?? getBizsId();
    const bizId = Number(rawBizId);
    const params: Record<string, string | number> = {
      resourceType: ResourceTypeEnum.CLB,
      id: taskId,
    };
    if (Number.isFinite(bizId) && bizId > 0) {
      params.bizId = bizId;
    }

    routerAction.open({
      name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
      params,
    });
  };

  const notifyClbTaskCreated = (taskId: string, prefixText: string) => {
    Message({
      theme: 'success',
      delay: 8000,
      message: h('span', [
        prefixText,
        h(
          'span',
          {
            style: { color: '#3a84ff', cursor: 'pointer' },
            onClick: () => openClbTaskDetail(taskId),
          },
          t('【任务管理-负载均衡】'),
        ),
        t('查看进度'),
      ]),
    });
  };

  return {
    getTaskId,
    openClbTaskDetail,
    notifyClbTaskCreated,
  };
}
