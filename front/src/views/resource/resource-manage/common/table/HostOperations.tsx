import { Button, Checkbox, Dialog, Loading, Message, bkTooltips } from 'bkui-vue';
import { PropType, Ref, computed, defineComponent, inject, reactive, ref, watch, withDirectives } from 'vue';
import './index.scss';
import { usePreviousState } from '@/hooks/usePreviousState';
import { useResourceStore } from '@/store';
import { VendorEnum } from '@/common/constant';
import { AngleDown } from 'bkui-vue/lib/icon';
import { BkDropdownItem } from 'bkui-vue/lib/dropdown';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import CommonLocalTable from '../commonLocalTable';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import http from '@/http';
import HcmDropdown from '@/components/hcm-dropdown/index.vue';
import HcmAuth from '@/components/auth/auth.vue';
import { AUTH_UPDATE_IAAS_RESOURCE } from '@/constants/auth-symbols';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

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

export enum OperationActions {
  NONE = 'none',
  START = 'start',
  STOP = 'stop',
  REBOOT = 'reboot',
  RECYCLE = 'recycle',
}

export const operationMap = {
  [OperationActions.NONE]: {
    label: 'unknown',
    disabledStatus: [] as string[],
    loading: false,
  },
  [OperationActions.START]: {
    label: '开机',
    disabledStatus: HOST_RUNNING_STATUS,
    loading: false,
  },
  [OperationActions.STOP]: {
    label: '关机',
    disabledStatus: HOST_SHUTDOWN_STATUS,
    loading: false,
  },
  [OperationActions.REBOOT]: {
    label: '重启',
    disabledStatus: HOST_SHUTDOWN_STATUS,
    loading: false,
  },
  [OperationActions.RECYCLE]: {
    label: '回收',
    disabledStatus: HOST_SHUTDOWN_STATUS,
    loading: false,
  },
};

export default defineComponent({
  props: {
    selections: {
      type: Array as PropType<
        Array<{
          status: string;
          id: string;
          vendor: string;
        }>
      >,
    },
    onFinished: {
      type: Function as PropType<(type: 'confirm' | 'cancel') => void>,
    },
  },
  setup(props) {
    const operationType = ref<OperationActions>(OperationActions.NONE);
    const dialogRef = ref(null);
    const dropdownOperationRef = ref(null);
    const dropdownCopyRef = ref(null);
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
    const resourceAccountStore = useResourceAccountStore();

    const isOtherVendor = inject<Ref<boolean>>('isOtherVendor');

    const isDialogShow = computed(() => {
      return operationType.value !== OperationActions.NONE;
    });

    const vendorSet = computed(() => {
      const vendors = props.selections.map((item) => item.vendor);
      return new Set(vendors);
    });

    const isMixOtherVendor = computed(() => {
      return vendorSet.value.size > 1 && vendorSet.value.has(VendorEnum.OTHER);
    });

    const computedTitle = computed(() => {
      if (operationType.value === OperationActions.NONE) {
        return `批量${operationMap[previousOperationType.value]?.label}`;
      }
      return `批量${operationMap[operationType.value]?.label}`;
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

    const getPrivateIPs = (data: any) => {
      return [...(data.private_ipv4_addresses || []), ...(data.private_ipv6_addresses || [])].join(',') || '--';
    };
    const getPublicIPs = (data: any) => {
      return [...(data.public_ipv4_addresses || []), ...(data.public_ipv6_addresses || [])].join(',') || '--';
    };
    const selectedRowPrivateIPs = computed(() => props.selections.map(getPrivateIPs));
    const selectedRowPublicIPs = computed(() => props.selections.map(getPublicIPs));

    const computedColumns = computed(() =>
      [
        {
          field: '_private_ip',
          label: '内网IP',
          render: ({ data }: any) => getPrivateIPs(data),
        },
        {
          field: '_public_ip',
          label: '外网IP',
          render: ({ data }: any) => getPublicIPs(data),
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
        ({ field }) => !['_with_disk', '_with_eip'].includes(field) || operationType.value === OperationActions.RECYCLE,
      ),
    );

    /**
     * 仅开机状态的主机能：关机、重启、回收
     * 仅关机状态的主机能：开机、回收
     */
    watch(
      () => operationType.value,
      async () => {
        if (operationType.value === OperationActions.NONE) return;
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
        handleSwitch(true);
        isConfirmDisabled.value = targetHost.value.length === 0;

        await getDiskNumByCvmIds();
      },
    );

    const computedContent = computed(() => {
      const targetOperationName = operationMap[operationType.value].label;
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
    });

    const getOperationConfig = (type: OperationActions) => {
      // 点击事件（值缺省时，为默认点击事件）
      const clickHandler = () => handleClickMenu(type);

      if (isMixOtherVendor.value) {
        return {
          disabled: true,
          tooltips: { content: '所选择的资源包含内置账号，不允许和其他云厂商同时选择', disabled: false },
          clickHandler,
        };
      }

      return { disabled: false, tooltips: { disabled: true }, clickHandler };
    };

    const handleClickMenu = (type: OperationActions) => {
      if (getOperationConfig(type).disabled) {
        return;
      }
      operationType.value = type;
    };

    const handleConfirm = async () => {
      try {
        isLoading.value = true;
        Message({
          message: `${computedTitle.value}中, 请不要操作`,
          theme: 'warning',
          delay: 1000,
        });
        if (operationType.value === OperationActions.RECYCLE) {
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
        operationType.value = OperationActions.NONE;
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
        <HcmAuth sign={{ type: AUTH_UPDATE_IAAS_RESOURCE, relation: [resourceAccountStore.resourceAccount?.id] }}>
          {{
            default: ({ noPerm }: { noPerm: boolean }) => (
              <HcmDropdown
                ref={dropdownOperationRef}
                disabled={noPerm || isOtherVendor.value || operationsDisabled.value}>
                {{
                  default: () => (
                    <>
                      批量操作
                      <AngleDown class='icon-angle-down'></AngleDown>
                    </>
                  ),
                  menus: () => (
                    <>
                      {Object.entries(operationMap)
                        .filter(([opType]) => opType !== OperationActions.NONE)
                        .map(([opType, opData]) => {
                          const { disabled, tooltips, clickHandler } = getOperationConfig(opType as OperationActions);
                          return withDirectives(
                            <BkDropdownItem
                              onClick={clickHandler}
                              extCls={`more-action-item${disabled ? ' disabled' : ''}`}>
                              批量{opData.label}
                            </BkDropdownItem>,
                            [[bkTooltips, tooltips]],
                          );
                        })}
                    </>
                  ),
                }}
              </HcmDropdown>
            ),
          }}
        </HcmAuth>

        <HcmDropdown ref={dropdownCopyRef} disabled={operationsDisabled.value}>
          {{
            default: () => (
              <>
                复制
                <AngleDown class='icon-angle-down'></AngleDown>
              </>
            ),
            menus: () => (
              <>
                <CopyToClipboard
                  type='dropdown-item'
                  text='内网IP'
                  content={selectedRowPrivateIPs.value?.join?.(',')}
                  onSuccess={() => dropdownCopyRef.value?.hidePopover()}
                />
                <CopyToClipboard
                  type='dropdown-item'
                  text='公网IP'
                  content={selectedRowPublicIPs.value?.join?.(',')}
                  onSuccess={() => dropdownCopyRef.value?.hidePopover()}
                />
              </>
            ),
          }}
        </HcmDropdown>

        <Dialog
          isShow={isDialogShow.value}
          // quick-close={!isLoading.value}
          quickClose={false}
          title={computedTitle.value}
          ref={dialogRef}
          width={1500}
          closeIcon={!isLoading.value}
          onClosed={() => (operationType.value = OperationActions.NONE)}>
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
                        可{operationMap[operationType.value].label}
                      </Button>
                      <Button onClick={() => handleSwitch(false)} selected={selected.value === 'untarget'}>
                        不可{operationMap[operationType.value].label}
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
                  {operationMap[operationType.value].label}
                </Button>
                <Button
                  onClick={() => (operationType.value = OperationActions.NONE)}
                  class='ml10'
                  disabled={isLoading.value}>
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
