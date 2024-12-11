import { PropType, defineComponent, reactive, ref, watch, computed, toRefs } from 'vue';
import { Button, Dialog, Loading } from 'bkui-vue';
import cssModule from './index.module.scss';
import { AngleDown } from 'bkui-vue/lib/icon';
import { BkDropdownItem } from 'bkui-vue/lib/dropdown';
import CommonLocalTable from '@/components/LocalTable';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import useBatchOperation from './use-batch-operation';
import HcmDropdown from '@/components/hcm-dropdown/index.vue';

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

export enum OperationActions {
  NONE = 'none',
  START = 'start',
  STOP = 'stop',
  REBOOT = 'reboot',
  RECYCLE = 'recycle',
}

export type OperationActionType = keyof typeof OperationActions;

export type OperationMapItem = {
  label: string;
  disabledStatus?: string[];
  loading?: boolean;
};

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
          private_ipv4_addresses: string[];
          __formSingleOp: boolean;
        }>
      >,
    },
    onFinished: {
      type: Function as PropType<(type: 'confirm' | 'cancel') => void>,
    },
  },
  setup(props) {
    const dialogRef = ref(null);
    const dropdownOperationRef = ref(null);
    const dropdownCopyRef = ref(null);

    const { selections } = toRefs(props);

    const {
      operationType,
      isDialogShow,
      isConfirmDisabled,
      operationsDisabled,
      baseColumns,
      recycleColumns,
      computedTitle,
      computedTips,
      computedContent,
      isLoading,
      selected,
      isDialogLoading,
      tableData,
      selectedRowPrivateIPs,
      selectedRowPublicIPs,
      getDiskNumByCvmIds,
      handleSwitch,
      handleConfirm,
      handleCancelDialog,
    } = useBatchOperation({
      selections,
      onFinished: props.onFinished,
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

    const computedColumns = computed(() => {
      const columns = baseColumns.value.slice();
      if (operationType.value === OperationActions.RECYCLE) {
        columns.push(...recycleColumns.value);
      }
      return columns;
    });

    watch(operationType, async (type) => {
      if (type === OperationActions.RECYCLE) {
        await getDiskNumByCvmIds();
      }
    });

    return () => (
      <>
        <HcmDropdown
          ref={dropdownOperationRef}
          class={cssModule.host_operations_container}
          disabled={operationsDisabled.value}>
          {{
            default: () => (
              <>
                批量操作
                <AngleDown class={cssModule.f26}></AngleDown>
              </>
            ),
            menus: () => (
              <>
                {Object.entries(operationMap)
                  .filter(([opType]) => opType !== OperationActions.NONE)
                  .map(([opType, opData]) => (
                    <BkDropdownItem
                      onClick={() => {
                        operationType.value = opType as OperationActions;
                        dropdownOperationRef.value?.hidePopover();
                      }}>
                      {`批量${opData.label}`}
                    </BkDropdownItem>
                  ))}
              </>
            ),
          }}
        </HcmDropdown>

        <HcmDropdown
          ref={dropdownCopyRef}
          class={cssModule.host_operations_container}
          disabled={operationsDisabled.value}>
          {{
            default: () => (
              <>
                复制
                <AngleDown class={cssModule.f26}></AngleDown>
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
          quickClose={false}
          title={computedTitle.value}
          ref={dialogRef}
          width={1500}
          closeIcon={!isLoading.value}>
          {{
            default: () => (
              <Loading loading={isDialogLoading.value}>
                <div class={cssModule['host-operations-main']}>
                  {computedTips.value && <div class={cssModule['host-operations-tips']}>{computedTips.value}</div>}
                  <CommonLocalTable
                    data={tableData.value}
                    columns={computedColumns.value}
                    changeData={(data) => (tableData.value = data)}
                    searchData={searchData}>
                    <div class={cssModule['host-operations-toolbar']}>
                      <BkButtonGroup>
                        <Button onClick={() => handleSwitch(true)} selected={selected.value === 'target'}>
                          可{operationMap[operationType.value].label}
                        </Button>
                        <Button onClick={() => handleSwitch(false)} selected={selected.value === 'untarget'}>
                          不可{operationMap[operationType.value].label}
                        </Button>
                      </BkButtonGroup>
                      {computedContent.value}
                    </div>
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
                <Button onClick={handleCancelDialog} class='ml10' disabled={isLoading.value}>
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
