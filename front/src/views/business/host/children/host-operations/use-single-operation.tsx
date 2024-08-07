import { ref } from 'vue';
import { Checkbox } from 'bkui-vue';
import { useBusinessStore } from '@/store';
import Confirm, { confirmInstance } from '@/components/confirm';
import { operationMap, OperationActions } from '@/views/business/host/children/host-operations';

const businessStore = useBusinessStore();

const useSingleOperation = ({
  beforeConfirm,
  confirmSuccess,
  confirmComplete,
}: {
  beforeConfirm: Function;
  confirmSuccess: Function;
  confirmComplete: Function;
}) => {
  const currentOperateRowIndex = ref(-1);

  // 回收参数「云硬盘/EIP 随主机回收」
  const isRecycleDiskWithCvm = ref(false);
  const isRecycleEipWithCvm = ref(false);

  // 重置回收参数
  const resetRecycleSingleCvmParams = () => {
    isRecycleDiskWithCvm.value = false;
    isRecycleEipWithCvm.value = false;
  };

  const isOperateDisabled = (type: OperationActions, status: string) =>
    operationMap[type].disabledStatus.includes(status);

  const handleOperate = async (type: OperationActions, data: any) => {
    if (isOperateDisabled(type, data.status)) return;

    const { label } = operationMap[type];

    resetRecycleSingleCvmParams();

    let infoboxContent = <>当前操作主机为：{data.name}</>;

    if (type === OperationActions.RECYCLE) {
      // 请求主机所关联的资源(硬盘, eip)个数
      const {
        data: [target],
      } = await businessStore.getRelResByCvmIds({ ids: [data.id] });
      const { disk_count, eip_count, eip } = target;
      infoboxContent = (
        <div style={{ textAlign: 'justify' }}>
          <div style={{ marginBottom: '10px' }}>
            当前操作主机为：{data.name}
            <br />
            共关联 {disk_count - 1} 个数据盘，{eip_count} 个弹性 IP{eip ? `(${eip.join(',')})` : ''}
          </div>
          <div>
            <Checkbox
              checked={isRecycleDiskWithCvm.value}
              onChange={(checked) => (isRecycleDiskWithCvm.value = checked)}>
              云硬盘随主机回收
            </Checkbox>
            <Checkbox checked={isRecycleEipWithCvm.value} onChange={(checked) => (isRecycleEipWithCvm.value = checked)}>
              弹性 IP 随主机回收
            </Checkbox>
          </div>
        </div>
      );
    }

    Confirm(`确定${label}`, infoboxContent, async () => {
      confirmInstance.hide();
      beforeConfirm();
      try {
        if (type === 'recycle') {
          await businessStore.recycledCvmsData({
            infos: [{ id: data.id, with_disk: isRecycleDiskWithCvm.value, with_eip: isRecycleEipWithCvm.value }],
          });
        } else {
          await businessStore.cvmOperate(type, { ids: [data.id] });
        }
        confirmSuccess(type);
      } finally {
        confirmComplete();
      }
    });
  };

  return {
    currentOperateRowIndex,
    isOperateDisabled,
    handleOperate,
  };
};

export default useSingleOperation;
