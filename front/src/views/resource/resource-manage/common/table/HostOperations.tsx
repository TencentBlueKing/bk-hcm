import { Button, Checkbox, Dialog, Dropdown, Loading, Message } from 'bkui-vue';
import { PropType, computed, defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import { usePreviousState } from '@/hooks/usePreviousState';
import { useResourceStore } from '@/store';
import { AngleDown } from 'bkui-vue/lib/icon';
import { BkDropdownItem, BkDropdownMenu } from 'bkui-vue/lib/dropdown';
import CommonLocalTable from '../commonLocalTable';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import http from '@/http';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

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
      type: Array as PropType<
        Array<{
          status: string;
          id: string;
        }>
      >,
    },
    onFinished: {
      type: Function as PropType<(type: 'confirm' | 'cancel') => void>,
    },
  },
  setup(props) {
    const operationType = ref<Operations>(Operations.None);
    const dialogRef = ref(null);
    const isConfirmDisabled = ref(true);
    const targetHost = ref([]);
    const unTargetHost = ref([]);
    const isLoading = ref(false);
    const tableData = ref([]);
    const selected = ref('target');
    const withDiskSet = ref(new Set());
    const withEipSet = ref(new Set());
    const cvmRelResMap = ref(new Map());
    const isDialogLoading = ref(false);

    const previousOperationType = usePreviousState(operationType);
    const resourceStore = useResourceStore();

    const isDialogShow = computed(() => {
      return operationType.value !== Operations.None;
    });

    const computedTitle = computed(() => {
      if (operationType.value === Operations.None) return `批量${OperationsMap[previousOperationType.value]}`;
      return `批量${OperationsMap[operationType.value]}`;
    });

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

    const computedColumns = computed(() =>
      [
        {
          field: '_private_ip',
          label: '内网IP',
          render: ({ data }: any) => [...data.private_ipv4_addresses, ...data.private_ipv6_addresses].join(',') || '--',
        },
        {
          field: '_public_ip',
          label: '外网IP',
          render: ({ data }: any) => [...data.public_ipv4_addresses, ...data.public_ipv6_addresses].join(',') || '--',
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
          render: ({ data }: any) => (
            <div>
              <Checkbox
                checked={withDiskSet.value.has(data.id)}
                onChange={(isChecked) => {
                  if (isChecked) withDiskSet.value.add(data.id);
                  else withDiskSet.value.delete(data.id);
                }}>
                {cvmRelResMap.value.get(data.id)?.disk_count - 1} 个数据盘
              </Checkbox>
            </div>
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
      ].filter(
        ({ field }) => !['_with_disk', '_with_eip'].includes(field) || operationType.value === Operations.Recycle,
      ),
    );

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
      async () => {
        if (operationType.value === Operations.None) return;
        const runningHosts = [];
        const unRunningHosts = [];

        const shutdownHosts = [];
        const unShutdownHosts = [];

        for (const host of props.selections) {
          const { status } = host;
          if (HOST_RUNNING_STATUS.includes(status)) runningHosts.push(host);
          else unRunningHosts.push(host);

          if (HOST_SHUTDOWN_STATUS.includes(status)) shutdownHosts.push(host);
          else unShutdownHosts.push(host);
        }

        switch (operationType.value) {
          case Operations.Open: {
            targetHost.value = shutdownHosts;
            unTargetHost.value = unShutdownHosts;
            break;
          }
          case Operations.Close:
          case Operations.Reboot: {
            targetHost.value = runningHosts;
            unTargetHost.value = unRunningHosts;
            break;
          }
          case Operations.Recycle: {
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
        handleSwitch(true);
        isConfirmDisabled.value = targetHost.value.length === 0;

        await getDiskNumByCvmIds();
      },
    );

    const computedContent = computed(() => {
      const targetOperationName = OperationsMap[operationType.value];
      // let oppositeOperationName = '';
      // switch (operationType.value) {
      //   case Operations.Open: {
      //     oppositeOperationName = OperationsMap[Operations.Close];
      //     break;
      //   }
      //   case Operations.Close:
      //   case Operations.Reboot: {
      //     oppositeOperationName = OperationsMap[Operations.Open];
      //     break;
      //   }
      //   case Operations.Recycle: {
      //     oppositeOperationName = `${OperationsMap[Operations.Open]}、${OperationsMap[Operations.Close]}`;
      //     break;
      //   }
      // }
      return (
        <p class={'host-operations-selection-tip'}>
          已选择 <span class={'host-operations-selection-tip-all'}>{props.selections.length}</span> 个主机， 其中可
          {targetOperationName} <span class={'host-operations-selection-tip-target'}>{targetHost.value.length}</span>{' '}
          个， 不可{targetOperationName}{' '}
          <span class={'host-operations-selection-tip-untarget'}>{unTargetHost.value.length}</span> 个。
          <Button text theme='primary'>
            查看{targetOperationName}规则
          </Button>
        </p>
      );
      // if (targetHostsNum === 0) {
      //   return (
      //     <p>
      //       您已选择了 {allHostsNum} 台主机进行
      //       {targetOperationName}操作, 其中
      //       <span class={'host_operations_blue_txt'}> {allHostsNum} </span>
      //       台是已{computedPreviousOperationType.value}的，不支持对其操作。
      //       <br />
      //       <span class={'host_operations_red_txt'}>
      //         由于所选主机均处于{targetOperationName}
      //         状态,无法进行操作。
      //       </span>
      //     </p>
      //   );
      // }
      // if (targetHostsNum === allHostsNum) {
      //   return (
      //     <p>
      //       您已选择了 {allHostsNum} 台主机进行
      //       {targetOperationName}操作,本次操作将对
      //       <span class={'host_operations_blue_txt'}> {allHostsNum} </span>
      //       台处于{oppositeOperationName}
      //       状态的主机进行{targetOperationName}操作。
      //       <br />
      //       <span class={'host_operations_red_txt'}>
      //         请确认您所选择的目标是正确的，该操作将对主机进行
      //         {targetOperationName}操作。
      //       </span>
      //     </p>
      //   );
      // }
      // if (allHostsNum > targetHostsNum) {
      //   return (
      //     <p>
      //       您已选择了 {allHostsNum} 台主机进行
      //       {targetOperationName}。本次操作将对
      //       <span class={'host_operations_blue_txt'}> {targetHostsNum} </span>
      //       台处于{oppositeOperationName}状态的主机进行
      //       {targetOperationName}，其余主机的状态不支持{targetOperationName}。
      //       <br />
      //       <span class={'host_operations_red_txt'}>
      //         请确认您所选择的目标是正确的,该操作将对主机进行
      //         {targetOperationName}操作
      //       </span>
      //     </p>
      //   );
      // }
      // return '';
    });

    const handleConfirm = async () => {
      try {
        isLoading.value = true;
        Message({
          message: `${computedTitle.value}中, 请不要操作`,
          theme: 'warning',
          delay: 1000,
        });
        if (operationType.value === Operations.Recycle) {
          const hostIds = targetHost.value.map((v) => ({
            id: v.id,
            with_disk: withDiskSet.value.has(v.id),
            with_eip: withEipSet.value.has(v.id),
          }));
          await resourceStore.recycledCvmsData({ infos: hostIds });
        } else {
          const hostIds = targetHost.value.map((v) => v.id);
          await resourceStore.cvmOperate(operationType.value, { ids: hostIds });
        }
        Message({
          message: '操作成功',
          theme: 'success',
        });
        props.onFinished('confirm');
      } catch (err) {
        // Message({
        //   message: '操作失败',
        //   theme: 'error',
        // });
      } finally {
        isLoading.value = false;
        operationType.value = Operations.None;
      }
    };

    const getDiskNumByCvmIds = async () => {
      isDialogLoading.value = true;
      try {
        const ids = props.selections.map(({ id }) => id);
        const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/cvms/rel_res/batch`, {
          ids,
        });
        for (let i = 0; i < res.data.length; i++) {
          cvmRelResMap.value.set(ids[i], res.data[i]);
        }
      } finally {
        isDialogLoading.value = false;
      }
    };

    const operationsDisabled = computed(() => !props.selections.length);

    const handleSwitch = (isTarget = true) => {
      if (isTarget) {
        tableData.value = targetHost.value;
        selected.value = 'target';
      } else {
        tableData.value = unTargetHost.value;
        selected.value = 'untarget';
      }
    };

    return () => (
      <>
        <div class={'host_operations_container'}>
          <Dropdown disabled={operationsDisabled.value}>
            {{
              default: () => (
                <Button disabled={operationsDisabled.value}>
                  批量操作
                  <AngleDown class={'f26'}></AngleDown>
                </Button>
              ),
              content: () => (
                <BkDropdownMenu>
                  {Object.entries(OperationsMap).map(([opType, opName]) => (
                    <BkDropdownItem
                      onClick={() => {
                        operationType.value = opType as Operations;
                      }}>
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
          // quick-close={!isLoading.value}
          quickClose={false}
          title={computedTitle.value}
          ref={dialogRef}
          width={1500}
          closeIcon={!isLoading.value}
          onClosed={() => (operationType.value = Operations.None)}>
          {{
            default: () => (
              <Loading loading={isDialogLoading.value}>
                <div>
                  <p>{computedContent.value}</p>
                  <CommonLocalTable
                    data={tableData.value}
                    columns={computedColumns.value}
                    changeData={(data) => (tableData.value = data)}
                    searchData={searchData}>
                    <BkButtonGroup>
                      <Button onClick={() => handleSwitch(true)} selected={selected.value === 'target'}>
                        可{OperationsMap[operationType.value]}
                      </Button>
                      <Button onClick={() => handleSwitch(false)} selected={selected.value === 'untarget'}>
                        不可{OperationsMap[operationType.value]}
                      </Button>
                    </BkButtonGroup>
                  </CommonLocalTable>
                </div>
              </Loading>
            ),
            footer: (
              <>
                <Button
                  onClick={handleConfirm}
                  theme='primary'
                  disabled={isConfirmDisabled.value}
                  loading={isLoading.value}>
                  {OperationsMap[operationType.value]}
                </Button>
                <Button onClick={() => (operationType.value = Operations.None)} class='ml10' disabled={isLoading.value}>
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
