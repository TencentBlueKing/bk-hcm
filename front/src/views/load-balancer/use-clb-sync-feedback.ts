import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { IClbTaskFeedbackResponse, useClbTaskFeedback } from './use-clb-task-feedback';

export type IClbSyncResponse = IClbTaskFeedbackResponse;

export function useClbSyncFeedback(getBizId?: () => number) {
  const { t } = useI18n();
  const { getTaskId, notifyClbTaskCreated } = useClbTaskFeedback(getBizId);

  const handleClbSyncError = (error: IClbSyncResponse) => {
    const message = error?.code === 2000002 ? t('该账号地域同步任务进行中') : error?.message || t('同步失败');
    Message({ theme: 'error', message });
  };

  const handleClbSyncSuccess = (res: IClbSyncResponse) => {
    const taskId = getTaskId(res);
    if (!taskId) {
      Message({ theme: 'warning', message: t('没有可处理的负载均衡') });
      return;
    }

    notifyClbTaskCreated(taskId, t('同步任务已创建，可在'));
  };

  return { handleClbSyncSuccess, handleClbSyncError };
}
