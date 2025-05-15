import { Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { Message } from 'bkui-vue';
import { CLOUD_HOST_STATUS } from '@/common/constant';
import { useAccountStore } from '@/store';
import defaultUseColumns from '@/views/resource/resource-manage/hooks/use-columns';
import HostOperations, { OperationActions, operationMap } from '@/views/business/host/children/host-operations';
import useSingleOperation from '@/views/business/host/children/host-operations/use-single-operation';
import defaultUseTableListQuery from '@/hooks/useTableListQuery';
import type { PropsType } from '@/hooks/useTableListQuery';
import { AUTH_BIZ_DELETE_IAAS_RESOURCE } from '@/constants/auth-symbols';

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
  const accountStore = useAccountStore();

  const { handleOperate, isOperateDisabled, currentOperateRowIndex } = useSingleOperation({
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

  const getBkToolTipsOption = (type: OperationActions, data: { status: keyof typeof CLOUD_HOST_STATUS }) => {
    return {
      content: `当前主机处于 ${CLOUD_HOST_STATUS[data.status]} 状态`,
      disabled: !isOperateDisabled(type, data.status),
    };
  };

  const { columns, generateColumnsSettings } = defaultUseColumns(columnType, isSimpleShow, vendor);

  return {
    columns: [
      ...columns,
      {
        label: '操作',
        width: 120,
        showOverflowTooltip: false,
        render: ({ data, index }: { data: any; index: number }) => {
          return (
            <div class={'operation-column'}>
              <hcm-auth sign={{ type: AUTH_BIZ_DELETE_IAAS_RESOURCE, relation: [accountStore.bizs] }}>
                {{
                  default: ({ noPerm }: { noPerm: boolean }) => (
                    <bk-button
                      v-bk-tooltips={getBkToolTipsOption(OperationActions.RECYCLE, data)}
                      text
                      theme='primary'
                      class='mr8'
                      disabled={noPerm || isOperateDisabled(OperationActions.RECYCLE, data)}
                      onClick={() => handleOperate(OperationActions.RECYCLE, data)}>
                      {operationMap[OperationActions.RECYCLE].label}
                    </bk-button>
                  ),
                }}
              </hcm-auth>
              <bk-dropdown
                trigger='click'
                popoverOptions={{
                  renderType: 'shown',
                  onAfterShow: () => (currentOperateRowIndex.value = index),
                  onAfterHidden: () => (currentOperateRowIndex.value = -1),
                }}>
                {{
                  default: () => (
                    <div class={[`more-action${currentOperateRowIndex.value === index ? ' current-operate-row' : ''}`]}>
                      <i class={'hcm-icon bkhcm-icon-more-fill'}></i>
                    </div>
                  ),
                  content: () => (
                    <bk-dropdown-menu>
                      {operationDropdownList.map(({ label, type }) => (
                        <bk-dropdown-item
                          v-bk-tooltips={getBkToolTipsOption(type as OperationActions, data)}
                          key={type}
                          onClick={() => handleOperate(type as OperationActions, data)}
                          extCls={`more-action-item${
                            isOperateDisabled(type as OperationActions, data.status) ? ' disabled' : ''
                          }`}>
                          {label}
                        </bk-dropdown-item>
                      ))}
                    </bk-dropdown-menu>
                  ),
                }}
              </bk-dropdown>
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
