import { h } from 'vue';
import { InfoBox, Message } from 'bkui-vue';
import { useGpuDemandStore } from '@/store/resource-plan/gpu-demand';

/**
 * GPU 需求终止确认弹窗（基于 InfoBox）
 * 统一维护主单终止 & 子单终止两种场景
 */
export function useTerminateConfirm() {
  const gpuDemandStore = useGpuDemandStore();

  /**
   * 主单终止确认
   * @param orderId 主单 ID
   * @param onSuccess 终止成功后的回调
   */
  const confirmTerminateOrder = (orderId: string, onSuccess?: () => void) => {
    InfoBox({
      title: '是否终止该需求？',
      subTitle: () => h('span', { style: { fontSize: '14px' } }, `需求：${orderId}`),
      theme: 'danger',
      confirmText: '终止',
      cancelText: '取消',
      headerAlign: 'center',
      contentAlign: 'center',
      footerAlign: 'center',
      quickClose: false,
      async onConfirm() {
        try {
          await gpuDemandStore.batchTerminateOrders({ order_ids: [orderId] });
          Message({ theme: 'success', message: '终止成功' });
          onSuccess?.();
        } catch {
          Message({ theme: 'error', message: '终止失败' });
        }
      },
    });
  };

  /**
   * 子单终止确认
   * @param subOrderId 子单 ID
   * @param onSuccess 终止成功后的回调
   */
  const confirmTerminateSubOrder = (subOrderId: string, onSuccess?: () => void) => {
    InfoBox({
      title: '是否终止该需求？',
      subTitle: () => h('span', { style: { fontSize: '14px' } }, subOrderId),
      theme: 'danger',
      confirmText: '终止',
      cancelText: '取消',
      headerAlign: 'center',
      contentAlign: 'center',
      footerAlign: 'center',
      quickClose: false,
      async onConfirm() {
        try {
          await gpuDemandStore.batchTerminateSubOrders({ suborder_ids: [subOrderId] });
          Message({ theme: 'success', message: '终止成功' });
          onSuccess?.();
        } catch {
          Message({ theme: 'error', message: '终止失败' });
        }
      },
    });
  };

  return {
    confirmTerminateOrder,
    confirmTerminateSubOrder,
  };
}
