import { withDirectives, Ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { Button, Dropdown, Message, bkTooltips } from 'bkui-vue';
import defaultUseColumns from '@/views/resource/resource-manage/hooks/use-columns';
import HostOperations, { OperationActions, operationMap } from '@/views/business/host/children/host-operations';
import useSingleOperation from '@/views/business/host/children/host-operations/use-single-operation';
import defaultUseTableListQuery from '@/hooks/useTableListQuery';
import type { PropsType } from '@/hooks/useTableListQuery';
import HcmAuth from '@/components/auth/auth.vue';
import {
  AUTH_UPDATE_IAAS_RESOURCE,
  AUTH_DELETE_IAAS_RESOURCE,
  AUTH_BIZ_UPDATE_IAAS_RESOURCE,
  AUTH_BIZ_DELETE_IAAS_RESOURCE,
} from '@/constants/auth-symbols';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { getAuthSignByBusinessId } from '@/utils';

const { DropdownMenu, DropdownItem } = Dropdown;

type UseColumnsParams = {
  columnType?: string;
  isSimpleShow?: boolean;
  vendor?: string;
  extra?: {
    isLoading: Ref<boolean>;
    triggerApi: () => void;
    getHostOperationRef: () => any;
    getTableRef: () => any;
  };
};

const useColumns = ({ columnType = 'cvms', isSimpleShow = false, vendor, extra }: UseColumnsParams) => {
  const { t } = useI18n();
  const router = useRouter();
  const { getBizsId } = useWhereAmI();

  const bizId = computed(() => getBizsId());
  const operateAuthSign = computed(() =>
    getAuthSignByBusinessId(bizId.value, AUTH_UPDATE_IAAS_RESOURCE, AUTH_BIZ_UPDATE_IAAS_RESOURCE),
  );
  const deleteAuthSign = computed(() =>
    getAuthSignByBusinessId(bizId.value, AUTH_DELETE_IAAS_RESOURCE, AUTH_BIZ_DELETE_IAAS_RESOURCE),
  );

  const { getOperationConfig, currentOperateRowIndex } = useSingleOperation({
    beforeConfirm() {
      extra.isLoading.value = true;
    },
    confirmSuccess(type: string) {
      Message({ message: t('操作成功'), theme: 'success' });
      if (type === OperationActions.RECYCLE) {
        router.push({ name: 'businessRecyclebin' });
      } else {
        extra.triggerApi();
      }
    },
    confirmComplete() {
      extra.isLoading.value = false;
    },
  });

  const operationDropdownList = Object.entries(operationMap)
    .filter(([type]) => ![OperationActions.RECYCLE, OperationActions.NONE].includes(type as OperationActions))
    .map(([type, value]) => ({
      type,
      label: value.label,
    }));

  const { columns, generateColumnsSettings } = defaultUseColumns(columnType, isSimpleShow, vendor);

  return {
    columns: [
      ...columns,
      {
        label: '操作',
        width: 120,
        showOverflowTooltip: false,
        render: ({ data, index }: { data: any; index: number }) => {
          const recycleConfig = getOperationConfig(OperationActions.RECYCLE, data);
          return (
            <div class={'operation-column'}>
              <HcmAuth sign={deleteAuthSign.value} tag='span'>
                {{
                  default: ({ noPerm }: { noPerm: boolean }) =>
                    withDirectives(
                      <Button
                        text
                        theme={'primary'}
                        class={'mr10'}
                        onClick={recycleConfig.clickHandler}
                        disabled={noPerm || recycleConfig.disabled}>
                        {operationMap[OperationActions.RECYCLE].label}
                      </Button>,
                      [[bkTooltips, recycleConfig.tooltips]],
                    ),
                }}
              </HcmAuth>
              <HcmAuth sign={operateAuthSign.value} tag='span'>
                {{
                  default: ({ noPerm }: { noPerm: boolean }) => (
                    <Dropdown
                      trigger='click'
                      popoverOptions={{
                        renderType: 'shown',
                        clickContentAutoHide: true,
                        onAfterShow: () => (currentOperateRowIndex.value = index),
                        onAfterHidden: () => (currentOperateRowIndex.value = -1),
                      }}>
                      {{
                        default: () => (
                          <div
                            class={[
                              'more-action',
                              {
                                'current-operate-row': currentOperateRowIndex.value === index,
                                disabled: noPerm,
                              },
                            ]}>
                            <i class={'hcm-icon bkhcm-icon-more-fill'}></i>
                          </div>
                        ),
                        content: () => (
                          <DropdownMenu>
                            {operationDropdownList.map(({ label, type }) => {
                              const { disabled, tooltips, clickHandler } = getOperationConfig(
                                type as OperationActions,
                                data,
                              );
                              return withDirectives(
                                <DropdownItem
                                  key={type}
                                  onClick={clickHandler}
                                  extCls={`more-action-item${disabled ? ' disabled' : ''}`}>
                                  {label}
                                </DropdownItem>,
                                [[bkTooltips, tooltips]],
                              );
                            })}
                          </DropdownMenu>
                        ),
                      }}
                    </Dropdown>
                  ),
                }}
              </HcmAuth>
            </div>
          );
        },
      },
    ],
    generateColumnsSettings,
  };
};

const useTableListQuery = (
  props: PropsType,
  type = 'cvms',
  completeCallback: () => void,
  apiMethod?: Function,
  apiName = 'list',
  args: any = {},
  extraResolveData?: (...args: any) => Promise<any>,
) => {
  return defaultUseTableListQuery(props, type, completeCallback, apiMethod, apiName, args, extraResolveData);
};

const pluginHandler = {
  useColumns,
  useTableListQuery,
  HostOperations,
};

export default pluginHandler;

export type PluginHandlerType = typeof pluginHandler;
