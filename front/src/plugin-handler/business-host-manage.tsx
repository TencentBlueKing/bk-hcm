import { withDirectives, Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { Button, Dropdown, Message, bkTooltips } from 'bkui-vue';
import defaultUseColumns from '@/views/resource/resource-manage/hooks/use-columns';
import HostOperations, { OperationActions, operationMap } from '@/views/business/host/children/host-operations';
import useSingleOperation from '@/views/business/host/children/host-operations/use-single-operation';
import defaultUseTableListQuery from '@/hooks/useTableListQuery';
import type { PropsType } from '@/hooks/useTableListQuery';

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
          return (
            <div class={'operation-column'}>
              {[
                withDirectives(
                  <Button
                    text
                    theme={'primary'}
                    class={`mr10 ${
                      getOperationConfig(OperationActions.RECYCLE, data).noPermission ? 'hcm-no-permision-text-btn' : ''
                    }`}
                    onClick={getOperationConfig(OperationActions.RECYCLE, data).clickHandler}
                    disabled={getOperationConfig(OperationActions.RECYCLE, data).disabled}>
                    {operationMap[OperationActions.RECYCLE].label}
                  </Button>,
                  [[bkTooltips, getOperationConfig(OperationActions.RECYCLE, data).tooltips]],
                ),
                <Dropdown
                  trigger='click'
                  popoverOptions={{
                    renderType: 'shown',
                    onAfterShow: () => (currentOperateRowIndex.value = index),
                    onAfterHidden: () => (currentOperateRowIndex.value = -1),
                  }}>
                  {{
                    default: () => (
                      <div
                        class={[`more-action${currentOperateRowIndex.value === index ? ' current-operate-row' : ''}`]}>
                        <i class={'hcm-icon bkhcm-icon-more-fill'}></i>
                      </div>
                    ),
                    content: () => (
                      <DropdownMenu>
                        {operationDropdownList.map(({ label, type }) => {
                          const { disabled, tooltips, noPermission, clickHandler } = getOperationConfig(
                            type as OperationActions,
                            data,
                          );
                          return withDirectives(
                            <DropdownItem
                              key={type}
                              onClick={clickHandler}
                              extCls={`more-action-item${disabled || noPermission ? ' disabled' : ''}`}>
                              {label}
                            </DropdownItem>,
                            [[bkTooltips, tooltips]],
                          );
                        })}
                      </DropdownMenu>
                    ),
                  }}
                </Dropdown>,
              ]}
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
