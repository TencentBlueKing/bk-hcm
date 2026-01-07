import { defineComponent, computed, ref, PropType, onMounted, reactive } from 'vue';
import { Button, Dropdown } from 'bkui-vue';
import { Plus as PlusIcon } from 'bkui-vue/lib/icon';
import { useTable } from '@/hooks/useResourcePlanTable';
import { useI18n } from 'vue-i18n';
import routerAction from '@/router/utils/action';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import BatchCancellationDialog from '@/components/resource-plan/resource-manage/list/table/components/batch-cancellation-dialog/batch-cancellation-dialog';
import BatchPostponeSideslider from '@/components/resource-plan/resource-manage/list/table/components/batch-postpone-sideslider/index.vue';
import cssModule from './index.module.scss';
import { useRoute, useRouter } from 'vue-router';
import { IListResourcesDemandsItem, IListResourcesDemandsParam, ResourcesDemandsStatus } from '@/typings/resourcePlan';
import { useConfigRequirementStore, type IRequirementObsProject } from '@/store/config/requirement';
import { IPageQuery } from '@/typings';
import { useResourcePlanStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';
import { ITimeRange } from '@/typings/plan';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';

const { DropdownMenu, DropdownItem } = Dropdown;

export enum OperationActions {
  SERVICE_ADJUST = 'service_adjust',
  SERVICE_CANCEL = 'service_cancel',
  BIZ_ADJUST = 'biz_adjust',
  BIZ_CANCEL = 'biz_cancel',
  PURCHASE = 'purchase',
  BIZ_BATCH_POSTPONE = 'biz_batch_postpone',
}

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      default: true,
    },
    expectTimeRange: {
      type: Object as PropType<ITimeRange>,
    },
  },
  setup(props, { expose }) {
    let searchModel: Partial<IListResourcesDemandsParam> = undefined;

    const { t } = useI18n();
    const route = useRoute();
    const router = useRouter();
    const { columns, generateColumnsSettings } = useColumns('resourceForecast');
    const { getResourcesDemandsList, getResourcesDemandsListByOrg } = useResourcePlanStore();
    const { getRequirementObsProject } = useConfigRequirementStore();
    const requirementObsProjectMap = ref<IRequirementObsProject>({});
    const { getBizsId } = useWhereAmI();
    const { cvmChargeTypes } = useCvmChargeType();

    const { authVerifyData, handleAuth } = useVerify();
    const globalPermissionDialog = useGlobalPermissionDialog();
    const operationMap = {
      [OperationActions.SERVICE_ADJUST]: {
        label: t('调整'),
        loading: false,
      },
      [OperationActions.SERVICE_CANCEL]: {
        label: t('取消'),
        loading: false,
      },
      [OperationActions.BIZ_ADJUST]: {
        label: t('调整'),
        loading: false,
      },
      [OperationActions.BIZ_CANCEL]: {
        label: t('取消'),
        loading: false,
      },
      [OperationActions.BIZ_BATCH_POSTPONE]: {
        label: t('部分延期'),
        loading: false,
      },
    };

    const bizActions = [OperationActions.BIZ_ADJUST, OperationActions.BIZ_CANCEL, OperationActions.BIZ_BATCH_POSTPONE];
    const serviceActions = [OperationActions.SERVICE_ADJUST, OperationActions.SERVICE_CANCEL];
    const operationDropdownList = Object.entries(operationMap)
      .filter(([type]) => (props.isBiz ? bizActions : serviceActions).includes(type as OperationActions))
      .map(([type, value]) => ({
        type,
        label: value.label,
      }));

    const tableRef = ref();
    const isShow = ref(false);
    const currentRowsData = ref<IListResourcesDemandsItem[]>([]);

    const selection = computed<IListResourcesDemandsItem[]>(() => tableRef.value?.getSelection?.() || []);
    const tableColumns = computed(() => {
      const newColumns = props.isBiz ? columns.slice(2) : columns.slice(0, -1);
      return [
        {
          type: 'selection',
          width: 30,
          minWidth: 30,
          fixed: 'left',
          onlyShowOnList: true,
        },
        {
          label: t('预测ID'),
          field: 'demand_id',
          minWidth: 90,
          fixed: 'left',
          isDefaultShow: true,
          render: ({ data }: any) => {
            return (
              <Button
                theme='primary'
                text
                onClick={() => {
                  router.push({
                    path: props.isBiz ? '/business/resource-plan/detail' : '/service/resource-plan/detail',
                    query: { ...route.query, demandId: data.demand_id },
                  });
                }}>
                {data.demand_id}
              </Button>
            );
          },
        },
        ...newColumns,
        {
          label: t('操作'),
          field: 'actions',
          fixed: 'right',
          minWidth: 100,
          isDefaultShow: true,
          render: ({ data }: { data: IListResourcesDemandsItem }) => {
            return (
              <div class={cssModule['operation-column']}>
                <Button
                  text
                  theme={'primary'}
                  class={`${
                    !authVerifyData.value?.permissionAction?.biz_iaas_resource_create
                      ? 'hcm-no-permision-text-btn'
                      : undefined
                  }`}
                  disabled={
                    data.status !== ResourcesDemandsStatus.CAN_APPLY ||
                    !['常规项目', '短租项目', '2026春节保障', '2026机房裁撤'].includes(data.obs_project)
                  }
                  onClick={() => handleApply(data)}>
                  一键申领
                </Button>
                <Dropdown trigger='click' disabled={!isRowSelectEnable({ row: data })}>
                  {{
                    default: () => (
                      <div class={cssModule['more-action']}>
                        <i class={'hcm-icon bkhcm-icon-more-fill'}></i>
                      </div>
                    ),
                    content: () => (
                      <DropdownMenu>
                        {operationDropdownList.map(({ label, type }) => (
                          <DropdownItem
                            key={type}
                            onClick={() => handleOperate(type as OperationActions, data)}
                            class={`${
                              !authVerifyData.value?.permissionAction?.biz_resource_plan_operate
                                ? 'hcm-no-permision-text-btn'
                                : undefined
                            }`}>
                            {label}
                          </DropdownItem>
                        ))}
                      </DropdownMenu>
                    ),
                  }}
                </Dropdown>
              </div>
            );
          },
        },
      ];
    });

    onMounted(async () => {
      requirementObsProjectMap.value = await getRequirementObsProject();
    });

    const settings = generateColumnsSettings(tableColumns.value);

    const getData = (page: IPageQuery) => {
      const params = {
        page,
        ...searchModel,
      };
      try {
        return props.isBiz ? getResourcesDemandsList(getBizsId(), params) : getResourcesDemandsListByOrg(params);
      } catch (error) {
        console.error('Error fetching data:', error);
      }
    };

    const {
      tableData,
      overview,
      pagination,
      isLoading,
      handlePageChange,
      handlePageSizeChange,
      handleSort,
      triggerApi,
      resetPagination,
    } = useTable(getData);

    const isRowSelectEnable = ({ row }: { row: IListResourcesDemandsItem }) => {
      return row.status === ResourcesDemandsStatus.CAN_APPLY || row.status === ResourcesDemandsStatus.NOT_READY;
    };

    const handleToAdd = () => {
      // 无权限
      if (!authVerifyData.value.permissionAction.biz_resource_plan_operate) {
        handleAuth('biz_resource_plan_operate');
        globalPermissionDialog.setShow(true);
      } else {
        router.push({
          path: '/business/resource-plan/add',
          query: { ...route.query },
        });
      }
    };

    const handleToAdjust = (data: IListResourcesDemandsItem[]) => {
      if (!authVerifyData.value.permissionAction.biz_resource_plan_operate) {
        // 无权限
        handleAuth('biz_resource_plan_operate');
        globalPermissionDialog.setShow(true);
      } else {
        const planIds = data.map(({ demand_id }) => demand_id).join(',');
        const path = props.isBiz ? '/business/service/resource-plan-mod' : '/service/resource-plan/mod';
        router.push({
          path,
          query: {
            [GLOBAL_BIZS_KEY]: getBizsId(),
            planIds,
            start: props.expectTimeRange.start,
            end: props.expectTimeRange.end,
          },
        });
      }
    };

    const handleToBizPage = (bizId: number) => {
      router.push({ name: 'bizResourcePlanList', query: { [GLOBAL_BIZS_KEY]: bizId, ...route.query } });
    };

    const handleCancel = () => {
      if (!authVerifyData.value.permissionAction.biz_resource_plan_operate) {
        handleAuth('biz_resource_plan_operate');
        globalPermissionDialog.setShow(true);
      } else {
        currentRowsData.value = selection.value;
        isShow.value = true;
      }
    };

    const batchPostponeSidesliderState = reactive({ isHidden: true, isShow: false, data: null });

    const handleOperate = (type: OperationActions, data: IListResourcesDemandsItem) => {
      if (!authVerifyData.value?.permissionAction?.biz_resource_plan_operate) {
        // 无权限
        handleAuth('biz_resource_plan_operate');
        globalPermissionDialog.setShow(true);
        return;
      }
      if (type === OperationActions.BIZ_CANCEL) {
        currentRowsData.value = [data];
        isShow.value = true;
      } else if (type === OperationActions.BIZ_ADJUST) {
        handleToAdjust([data]);
      } else if (type === OperationActions.BIZ_BATCH_POSTPONE) {
        Object.assign(batchPostponeSidesliderState, { isHidden: false, isShow: true, data });
      } else if ([OperationActions.SERVICE_ADJUST, OperationActions.SERVICE_CANCEL].includes(type)) {
        handleToBizPage(data.bk_biz_id);
      }
    };

    const handleApply = (data: IListResourcesDemandsItem) => {
      const { demand_id, bk_biz_id, device_type, region_id: region, zone_id: zone, obs_project, plan_type } = data;

      // 由obs_project得到requireType
      const requireType = Object.entries(requirementObsProjectMap.value).find(
        ([, value]) => value === obs_project,
      )?.[0];

      routerAction.open({
        path: '/business/service/service-apply/cvm',
        query: {
          device_type,
          region,
          zone,
          require_type: requireType,
          id: demand_id,
          charge_type: plan_type === '预测内' ? cvmChargeTypes.PREPAID : cvmChargeTypes.POSTPAID_BY_HOUR,
          from: 'businessResourcePlan',
          [GLOBAL_BIZS_KEY]: bk_biz_id,
        },
      });
    };

    const searchTableData = (data: Partial<IListResourcesDemandsParam>) => {
      searchModel = data;
      resetPagination();
      triggerApi();
    };

    expose({
      searchTableData,
    });

    return () => (
      <div class={cssModule['table-wrapper']}>
        <div class={cssModule['table-header']}>
          <div class={cssModule['toolbar-buttons']}>
            {props.isBiz && (
              <>
                <Button
                  class={`${cssModule.button} ${
                    !authVerifyData.value.permissionAction.biz_resource_plan_operate
                      ? 'hcm-no-permision-btn'
                      : undefined
                  }`}
                  theme='primary'
                  onClick={handleToAdd}>
                  <PlusIcon class={cssModule['plus-icon']} />
                  {t('新增预测')}
                </Button>
                <Button
                  class={`${cssModule.button} ${
                    !authVerifyData.value.permissionAction.biz_resource_plan_operate
                      ? 'hcm-no-permision-btn'
                      : undefined
                  }`}
                  onClick={() => handleToAdjust(selection.value)}
                  disabled={!selection.value.length}>
                  {t('批量调整')}
                </Button>
                <Button
                  class={`${cssModule.button} ${
                    !authVerifyData.value.permissionAction.biz_resource_plan_operate
                      ? 'hcm-no-permision-btn'
                      : undefined
                  }`}
                  onClick={handleCancel}
                  disabled={!selection.value.length}>
                  {t('批量取消')}
                </Button>
              </>
            )}
          </div>
          <div class={cssModule.overview}>
            <span>{`${t('本月即将过期 CPU ')}${overview.value?.expiring_cpu_core ?? '--'}${t('核')}`}</span>
            <div class={cssModule['cpu-stats']}>
              <span>
                {t('CPU 总核数')}：
                <span class={cssModule.num}>
                  {overview.value?.total_cpu_core ?? '--'}/{overview.value?.total_applied_core ?? '--'}
                </span>
              </span>
              <span>
                {t('预测内')}：
                <span class={cssModule.num}>
                  {overview.value?.in_plan_cpu_core ?? '--'}/{overview.value?.in_plan_applied_cpu_core ?? '--'}
                </span>
              </span>
              <span>
                {t('预测外')}：
                <span class={cssModule.num}>
                  {overview.value?.out_plan_cpu_core ?? '--'}/{overview.value?.out_plan_applied_cpu_core ?? '--'}
                </span>
              </span>
            </div>
          </div>
        </div>
        <bk-loading loading={isLoading.value}>
          <bk-table
            ref={tableRef}
            row-hover='auto'
            remote-pagination
            show-overflow-tooltip
            data={tableData.value}
            pagination={pagination.value}
            columns={tableColumns.value}
            settings={settings.value}
            isRowSelectEnable={isRowSelectEnable}
            onPageLimitChange={handlePageSizeChange}
            onPageValueChange={handlePageChange}
            onColumnSort={handleSort}
          />
        </bk-loading>
        <BatchCancellationDialog v-model:isShow={isShow.value} data={currentRowsData.value} onRefresh={triggerApi} />
        {!batchPostponeSidesliderState.isHidden && (
          <BatchPostponeSideslider
            v-model={batchPostponeSidesliderState.isShow}
            data={batchPostponeSidesliderState.data}
            onHidden={() => {
              batchPostponeSidesliderState.isHidden = true;
              batchPostponeSidesliderState.data = null;
            }}
          />
        )}
      </div>
    );
  },
});
