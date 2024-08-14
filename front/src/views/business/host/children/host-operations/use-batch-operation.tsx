import { computed, ref, type Ref, watch, reactive, VNode } from 'vue';
import { useRouter } from 'vue-router';
import { Message, Checkbox } from 'bkui-vue';
import { usePreviousState } from '@/hooks/usePreviousState';
import { useBusinessStore } from '@/store';
import { OperationActions, operationMap, HOST_RUNNING_STATUS, HOST_SHUTDOWN_STATUS } from './index';
import cssModule from './index.module.scss';

export type Params = {
  selections: Ref<
    Array<{
      status: string;
      id: string;
      vendor: string;
      private_ipv4_addresses: string[];
      __formSingleOp: boolean;
    }>
  >;
  onFinished: (type: 'confirm' | 'cancel') => void;
};

const useBatchOperation = ({ selections, onFinished }: Params) => {
  const router = useRouter();
  const operationType = ref<OperationActions>(OperationActions.NONE);

  const isConfirmDisabled = ref(true);
  const targetHost = ref([]);
  const unTargetHost = ref([]);
  const isLoading = ref(false);
  const tableData = ref([]);
  const selected = ref('target');
  const isDialogLoading = ref(false);
  const withDiskSet = ref(new Set());
  const withEipSet = ref(new Set());
  const cvmRelResMap = ref(new Map());
  const businessStore = useBusinessStore();

  const previousOperationType = usePreviousState(operationType);

  const operationsDisabled = computed(() => !selections.value.filter((item) => !item.__formSingleOp).length);

  const isDialogShow = computed(() => {
    return operationType.value !== OperationActions.NONE;
  });

  const computedTitle = computed(() => {
    if (operationType.value === OperationActions.NONE) return `批量${operationMap[previousOperationType.value]?.label}`;
    return `批量${operationMap[operationType.value].label}`;
  });

  const getDiskNumByCvmIds = async () => {
    isDialogLoading.value = true;
    try {
      const ids = selections.value.map(({ id }) => id);
      const res = await businessStore.getRelResByCvmIds({ ids });
      for (let i = 0; i < res.data.length; i++) {
        cvmRelResMap.value.set(ids[i], res.data[i]);
      }
    } finally {
      isDialogLoading.value = false;
    }
  };

  const getPrivateIPs = (data: any) => {
    return [...(data.private_ipv4_addresses || []), ...(data.private_ipv6_addresses || [])].join(',') || '--';
  };
  const getPublicIPs = (data: any) => {
    return [...(data.public_ipv4_addresses || []), ...(data.public_ipv6_addresses || [])].join(',') || '--';
  };

  const selectedRowPrivateIPs = computed(() => selections.value.map(getPrivateIPs));
  const selectedRowPublicIPs = computed(() => selections.value.map(getPublicIPs));

  const baseColumns = computed(() => [
    {
      field: '_private_ip',
      label: '内网IP',
      render: ({ data }: any) => <span>{getPrivateIPs(data)}</span>,
    },
    {
      field: '_public_ip',
      label: '外网IP',
      render: ({ data }: any) => <span>{getPublicIPs(data)}</span>,
    },
    {
      field: 'name',
      label: '主机名称',
    },
    {
      field: 'region',
      label: '地域',
    },
    {
      field: 'zone',
      label: '可用区',
    },
    {
      field: 'status',
      label: '状态',
    },
    {
      field: 'machine_type',
      label: '机型',
    },
  ]);

  const recycleColumns = computed(() => [
    {
      field: '_with_disk',
      label: '云硬盘随主机回收',
      render: ({ data }: any) => (
        <Checkbox
          checked={withDiskSet.value.has(data.id)}
          onChange={(isChecked) => {
            if (isChecked) withDiskSet.value.add(data.id);
            else withDiskSet.value.delete(data.id);
          }}>
          {cvmRelResMap.value.get(data.id)?.disk_count - 1} 个数据盘
        </Checkbox>
      ),
    },
    {
      field: '_with_eip',
      label: '弹性 IP 随主机回收',
      render: ({ data }: any) => (
        <Checkbox
          checked={withEipSet.value.has(data.id)}
          onChange={(isChecked) => {
            if (isChecked) withEipSet.value.add(data.id);
            else withEipSet.value.delete(data.id);
          }}>
          {cvmRelResMap.value.get(data.id)?.eip?.join(',') || '--'}
        </Checkbox>
      ),
    },
  ]);

  const searchData = reactive([
    {
      name: '地域',
      id: 'region',
    },
    {
      name: '可用区',
      id: 'zone',
    },
    {
      name: '状态',
      id: 'status',
    },
  ]);

  /**
   * 仅开机状态的主机能：关机、重启、回收
   * 仅关机状态的主机能：开机、回收
   */
  watch(operationType, async (type) => {
    if (type === OperationActions.NONE) return;
    const runningHosts = [];
    const unRunningHosts = [];

    const shutdownHosts = [];
    const unShutdownHosts = [];

    for (const host of selections.value) {
      const { status } = host;
      if (HOST_RUNNING_STATUS.includes(status)) runningHosts.push(host);
      else unRunningHosts.push(host);

      if (HOST_SHUTDOWN_STATUS.includes(status)) shutdownHosts.push(host);
      else unShutdownHosts.push(host);
    }

    switch (type) {
      case OperationActions.START: {
        targetHost.value = shutdownHosts;
        unTargetHost.value = unShutdownHosts;
        break;
      }
      case OperationActions.STOP:
      case OperationActions.REBOOT: {
        targetHost.value = runningHosts;
        unTargetHost.value = unRunningHosts;
        break;
      }
      case OperationActions.RECYCLE: {
        targetHost.value = [...runningHosts, ...shutdownHosts];
        unTargetHost.value = [...unRunningHosts, ...unShutdownHosts];
        const targetIdsSet = new Set([...runningHosts, ...shutdownHosts].map(({ id }) => id));
        const untargetIdSet = new Set();
        unTargetHost.value = [...unRunningHosts, ...unShutdownHosts].reduce((acc, cur) => {
          if (!untargetIdSet.has(cur.id) && !targetIdsSet.has(cur.id)) {
            acc.push(cur);
            untargetIdSet.add(cur.id);
          }
          return acc;
        }, []);
      }
    }

    isConfirmDisabled.value = targetHost.value.length === 0;
    handleSwitch(targetHost.value.length > 0);
  });

  const computedContent = computed(() => {
    const targetOperationName = operationMap[operationType.value].label;
    return (
      <div class={cssModule['host-operations-selection-tip']}>
        已选择 <span class={cssModule['host-operations-selection-tip-all']}>{selections.value.length}</span> 个主机，
        其中可
        {targetOperationName}{' '}
        <span class={cssModule['host-operations-selection-tip-target']}>{targetHost.value.length}</span> 个， 不可
        {targetOperationName}{' '}
        <span class={cssModule['host-operations-selection-tip-untarget']}>{unTargetHost.value.length}</span> 个。
      </div>
    );
  });

  const computedTips = computed(() => {
    const tips: { [key in OperationActions]: VNode | null } = {
      [OperationActions.RECYCLE]: (
        <>
          <p>
            回收的规则：1.主机已分配业务，仅允许处于配置平台待回收模块的主机回收。2.主机未分配业务，可以直接发起回收。3.回收过程中，对主机进行关机操作。4.主机上绑定的硬盘、弹性IP，由用户确认是否随主机回收销毁。
          </p>
          <p>
            回收后销毁：1.默认在回收站保留 48
            小时，保留的时长可配置。修改后的保留时长，仅对下一次的回收任务生效。2.支持对进入回收站的主机进行销毁操作。3.如无任何操作，过期后将自动进行销毁操作。4.一经销毁，无法从云上再找回该资源。
          </p>
          <p>回收后恢复：1.资源在保留时间内，可以在回收站进行恢复。2.主机恢复后，默认是关机的状态。</p>
        </>
      ),
      [OperationActions.NONE]: null,
      [OperationActions.REBOOT]: null,
      [OperationActions.START]: null,
      [OperationActions.STOP]: null,
    };
    return tips[operationType.value];
  });

  const handleConfirm = async () => {
    const isRecycle = operationType.value === OperationActions.RECYCLE;
    try {
      isLoading.value = true;
      Message({
        message: `${computedTitle.value}中, 请不要操作`,
        theme: 'warning',
        delay: 500,
      });
      if (isRecycle) {
        const hostIds = targetHost.value.map((v) => ({
          id: v.id,
          with_disk: withDiskSet.value.has(v.id),
          with_eip: withEipSet.value.has(v.id),
        }));
        await businessStore.recycledCvmsData({ infos: hostIds });
      } else {
        const hostIds = targetHost.value.map((v) => v.id);
        await businessStore.cvmOperate(operationType.value, { ids: hostIds });
      }
      Message({
        message: '操作成功',
        theme: 'success',
      });
      onFinished?.('confirm');
      if (isRecycle) {
        router.push({ name: 'businessRecyclebin' });
      }
    } finally {
      isLoading.value = false;
      operationType.value = OperationActions.NONE;
    }
  };

  const handleSwitch = (isTarget = true) => {
    if (isTarget) {
      tableData.value = targetHost.value;
      selected.value = 'target';
    } else {
      tableData.value = unTargetHost.value;
      selected.value = 'untarget';
    }
  };

  const handleCancelDialog = () => {
    operationType.value = OperationActions.NONE;
    const singleOpIndex = selections.value.findIndex((item) => item.__formSingleOp);
    if (singleOpIndex !== -1) {
      selections.value.splice(singleOpIndex, 1);
    }
  };

  return {
    operationType,
    isDialogShow,
    baseColumns,
    recycleColumns,
    computedTitle,
    computedTips,
    computedContent,
    operationsDisabled,
    isConfirmDisabled,
    isLoading,
    targetHost,
    unTargetHost,
    tableData,
    selected,
    isDialogLoading,
    searchData,
    selectedRowPrivateIPs,
    selectedRowPublicIPs,
    getDiskNumByCvmIds,
    handleSwitch,
    handleConfirm,
    handleCancelDialog,
  };
};

export type UseBatchOperationType = typeof useBatchOperation;

export default useBatchOperation;
