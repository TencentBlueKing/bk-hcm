import { Button, Dialog, Dropdown, Message } from 'bkui-vue';
import { PropType, computed, defineComponent, ref, watch } from 'vue';
import './index.scss';
import { usePreviousState } from '@/hooks/usePreviousState';
import { useResourceStore } from '@/store';
import { AngleDown } from 'bkui-vue/lib/icon';
import { BkDropdownItem, BkDropdownMenu } from 'bkui-vue/lib/dropdown';
import { useLocalTable } from '@/hooks/useLocalTable';

export enum Operations {
  None = 'none',
  Open = 'start',
  Close = 'stop',
  Reboot = 'reboot',
  Recycle = 'destroy',
}

export const OperationsMap = {
  [Operations.Open]: '开机',
  [Operations.Close]: '关机',
  [Operations.Reboot]: '重启',
  [Operations.Recycle]: '回收',
};

export const HOST_SHUTDOWN_STATUS = [
  'TERMINATED',
  'PowerState/stopped',
  'SHUTOFF',
  'STOPPED',
  'STOPPING',
  'PowerState/stopping',
  'stopped',
];
export const HOST_RUNNING_STATUS = [
  'STAGING',
  'RUNNING',
  'PowerState/starting',
  'PowerState/running',
  'ACTIVE',
  'running',
];
export const HOST_REBOOT_STATUS = ['REBOOT', 'HARD_REBOOT', 'REBOOTING'];

export default defineComponent({
  props: {
    selections: {
      type: Array as PropType<Array<{ status: string }>>,
    },
    onFinished: {
      type: Function as PropType<() => void>,
    },
  },
  setup(props) {
    const operationType = ref<Operations>(Operations.None);
    const dialogRef = ref(null);
    const isConfirmDisabled = ref(true);
    const targetHost = ref([]);
    const isLoading = ref(false);

    const previousOperationType = usePreviousState(operationType);
    const resourceStore = useResourceStore();

    const isDialogShow = computed(() => {
      return operationType.value !== Operations.None;
    });

    const computedTitle = computed(() => {
      if (operationType.value === Operations.None) return `批量${OperationsMap[previousOperationType.value]}`;
      return `批量${OperationsMap[operationType.value]}`;
    });

    // const computedPreviousOperationType = computed(() => {
    //   switch (operationType.value) {
    //     case Operations.None:
    //       return OperationsMap[previousOperationType.value];
    //     case Operations.Reboot:
    //       return OperationsMap[Operations.Close];
    //     case Operations.Recycle:
    //       return OperationsMap[Operations.Reboot];
    //   }
    //   return OperationsMap[operationType.value];
    // });

    /**
     * 仅开机状态的主机能：关机、重启、回收
     * 仅关机状态的主机能：开机、回收
     */
    watch(
      () => operationType.value,
      () => {
        if (operationType.value === Operations.None) return;

        const runningHosts = [];
        const shutdownHosts = [];

        for (const host of props.selections) {
          const { status } = host;
          if (HOST_RUNNING_STATUS.includes(status)) runningHosts.push(host);
          if (HOST_SHUTDOWN_STATUS.includes(status)) shutdownHosts.push(host);
        }

        switch (operationType.value) {
          case Operations.Open: {
            targetHost.value = shutdownHosts;
            break;
          }
          case Operations.Close:
          case Operations.Reboot: {
            targetHost.value = runningHosts;
            break;
          }
          case Operations.Recycle: {
            targetHost.value = [...runningHosts, ...shutdownHosts];
          }
        }
        isConfirmDisabled.value = targetHost.value.length === 0;
      },
    );

    // const computedContent = computed(() => {
    //   const allHostsNum = props.selections.length;
    //   const targetHostsNum = targetHost.value.length;
    //   const targetOperationName = OperationsMap[operationType.value];
    //   let oppositeOperationName = '';
    //   switch (operationType.value) {
    //     case Operations.Open: {
    //       oppositeOperationName = OperationsMap[Operations.Close];
    //       break;
    //     }
    //     case Operations.Close:
    //     case Operations.Reboot: {
    //       oppositeOperationName = OperationsMap[Operations.Open];
    //       break;
    //     }
    //     case Operations.Recycle: {
    //       oppositeOperationName = `${OperationsMap[Operations.Open]}、${OperationsMap[Operations.Close]}`;
    //       break;
    //     }
    //   }
    //   if (targetHostsNum === 0) {
    //     return (
    //       <p>
    //         您已选择了 {allHostsNum} 台主机进行
    //         {targetOperationName}操作, 其中
    //         <span class={'host_operations_blue_txt'}> {allHostsNum} </span>
    //         台是已{computedPreviousOperationType.value}的，不支持对其操作。
    //         <br />
    //         <span class={'host_operations_red_txt'}>
    //           由于所选主机均处于{targetOperationName}
    //           状态,无法进行操作。
    //         </span>
    //       </p>
    //     );
    //   }
    //   if (targetHostsNum === allHostsNum) {
    //     return (
    //       <p>
    //         您已选择了 {allHostsNum} 台主机进行
    //         {targetOperationName}操作,本次操作将对
    //         <span class={'host_operations_blue_txt'}> {allHostsNum} </span>
    //         台处于{oppositeOperationName}
    //         状态的主机进行{targetOperationName}操作。
    //         <br />
    //         <span class={'host_operations_red_txt'}>
    //           请确认您所选择的目标是正确的，该操作将对主机进行
    //           {targetOperationName}操作。
    //         </span>
    //       </p>
    //     );
    //   }
    //   if (allHostsNum > targetHostsNum) {
    //     return (
    //       <p>
    //         您已选择了 {allHostsNum} 台主机进行
    //         {targetOperationName}。本次操作将对
    //         <span class={'host_operations_blue_txt'}> {targetHostsNum} </span>
    //         台处于{oppositeOperationName}状态的主机进行
    //         {targetOperationName}，其余主机的状态不支持{targetOperationName}。
    //         <br />
    //         <span class={'host_operations_red_txt'}>
    //           请确认您所选择的目标是正确的,该操作将对主机进行
    //           {targetOperationName}操作
    //         </span>
    //       </p>
    //     );
    //   }
    //   return '';
    // });

    const handleConfirm = async () => {
      try {
        isLoading.value = true;
        Message({
          message: `${computedTitle.value}中, 请不要操作`,
          theme: 'warning',
        });
        if (operationType.value === Operations.Recycle) {
          const hostIds = targetHost.value.map(v => ({ id: v.id })) as Array<Record<string, string>>;
          await resourceStore.recycledCvmsData({ infos: hostIds });
        } else {
          const hostIds = targetHost.value.map(v => v.id);
          await resourceStore.cvmOperate(operationType.value, { ids: hostIds });
        }
        Message({
          message: '操作成功',
          theme: 'success',
        });
      } catch (err) {
        Message({
          message: '操作失败',
          theme: 'error',
        });
      } finally {
        isLoading.value = false;
        operationType.value = Operations.None;
        props.onFinished();
      }
    };

    const operationsDisabled = computed(() => !props.selections.length);

    const { CommonLocalTable } = useLocalTable({
      data: props.selections,
      columns: [
        {
          field: '_private_ip',
          label: '内网IP',
          render: ({ data }) => ([...data.private_ipv4_addresses, ...data.private_ipv6_addresses].join(',') || '--'),
        },
        {
          field: '_public_ip',
          label: '外网IP',
          render: ({ data }) => ([...data.public_ipv4_addresses, ...data.public_ipv6_addresses].join(',') || '--'),
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
        {
          field: '_with_disk',
          label: '云硬盘随主机回收',
        },
        {
          field: '_with_eip',
          label: '弹性 IP 随主机回收',
        },
      ],
      searchData: [
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
      ],
    });

    return () => (
      <>
        <div class={'host_operations_container'}>
          <Dropdown trigger='click' disabled={operationsDisabled.value}>
            {{
              default: () => (
                <Button disabled={operationsDisabled.value}>
                  批量操作
                  <AngleDown class={'f20'}></AngleDown>
                </Button>
              ),
              content: () => (
                <BkDropdownMenu>
                  {Object.entries(OperationsMap).map(([opType, opName]) => (
                    <BkDropdownItem
                      onClick={
                        () => (operationType.value = opType as Operations)
                      }>
                      {`批量${opName}`}
                    </BkDropdownItem>
                  ))}
                </BkDropdownMenu>
              ),
            }}
          </Dropdown>
        </div>

        <Dialog
          isShow={isDialogShow.value}
          quick-close={!isLoading.value}
          onClosed={() => (operationType.value = Operations.None)}
          onConfirm={handleConfirm}
          title={computedTitle.value}
          ref={dialogRef}
          width={1200}
          closeIcon={!isLoading.value}>
          {{
            default: () => (
             <CommonLocalTable/>
            ),
            footer: (
              <>
                <Button
                  onClick={dialogRef?.value?.handleConfirm}
                  theme='primary'
                  disabled={isConfirmDisabled.value}
                  loading={isLoading.value}>
                  确定
                </Button>
                <Button onClick={dialogRef?.value?.handleClose} class='ml10' disabled={isLoading.value}>
                  取消
                </Button>
              </>
            ),
          }}
        </Dialog>
      </>
    );
  },
});
