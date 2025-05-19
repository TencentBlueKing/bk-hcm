import { PropType, defineComponent, reactive, ref, watch, computed, toRefs, withDirectives } from 'vue';
import { Button, Dialog, Loading, bkTooltips } from 'bkui-vue';
import cssModule from './index.module.scss';
import { AngleDown } from 'bkui-vue/lib/icon';
import { BkDropdownItem } from 'bkui-vue/lib/dropdown';
import { VendorEnum } from '@/common/constant';
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
  authId?: string;
  actionName?: string;
};

export const operationMap: Record<OperationActions, OperationMapItem> = {
  [OperationActions.NONE]: {
    label: 'unknown',
    disabledStatus: [] as string[],
    loading: false,
  },
  [OperationActions.START]: {
    label: '开机',
    disabledStatus: HOST_RUNNING_STATUS,
    loading: false,
    // 鉴权参数
    authId: 'biz_iaas_resource_operate',
    actionName: 'biz_iaas_resource_operate',
  },
  [OperationActions.STOP]: {
    label: '关机',
    disabledStatus: HOST_SHUTDOWN_STATUS,
    loading: false,
    // 鉴权参数
    authId: 'biz_iaas_resource_operate',
    actionName: 'biz_iaas_resource_operate',
  },
  [OperationActions.REBOOT]: {
    label: '重启',
    disabledStatus: HOST_SHUTDOWN_STATUS,
    loading: false,
    authId: 'biz_iaas_resource_operate',
    actionName: 'biz_iaas_resource_operate',
  },
  [OperationActions.RECYCLE]: {
    label: '回收',
    disabledStatus: HOST_SHUTDOWN_STATUS,
    loading: false,
    authId: 'biz_iaas_resource_delete',
    actionName: 'biz_iaas_resource_delete',
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

    const vendorSet = computed(() => {
      const vendors = selections.value.map((item) => item.vendor);
      return new Set(vendors);
    });

    const isOtherOnly = computed(() => vendorSet.value.size === 1 && vendorSet.value.has(VendorEnum.OTHER));

    const isMixOtherVendor = computed(() => {
      return vendorSet.value.size > 1 && vendorSet.value.has(VendorEnum.OTHER);
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

      if (isOtherOnly.value) {
        return {
          disabled: true,
          tooltips: { content: '暂不支持', disabled: false },
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
        <HcmDropdown ref={dropdownOperationRef} disabled={operationsDisabled.value}>
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
                      <BkDropdownItem onClick={clickHandler} extCls={`more-action-item${disabled ? ' disabled' : ''}`}>
                        批量{opData.label}
                      </BkDropdownItem>,
                      [[bkTooltips, tooltips]],
                    );
                  })}
              </>
            ),
          }}
        </HcmDropdown>

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
