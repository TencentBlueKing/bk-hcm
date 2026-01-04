import { computed, defineComponent, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import cssModule from './index.module.scss';
import scrCssModule from '@/views/resource/resource-manage/hooks/use-scr-columns.module.scss';

import { Select } from 'bkui-vue';
import GridFilterComp from '@/components/grid-filter-comp';
import ExportToExcelBatchButton from '@/components/export-to-excel-batch-button/index.vue';
import FloatInput from '@/components/float-input';
import ScrDatePicker from '@/components/scr/scr-date-picker';

import { useI18n } from 'vue-i18n';
import { useUserStore } from '@/store';
import useScrColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import { getDeviceTypeList, getRecycleStageOpts, getRegionList, getZoneList } from '@/api/host/recycle';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useSaveSearchRules } from '@/views/ticket/utils/useSaveSearchRules';
import useFormModel from '@/hooks/useFormModel';
import { applicationTime } from '@/common/util';
import rollRequest from '@blueking/roll-request';
import http from '@/http';

export default defineComponent({
  setup() {
    const router = useRouter();
    const userStore = useUserStore();
    const { t } = useI18n();
    const { getBusinessApiPath, getBizsId } = useWhereAmI();
    const route = useRoute();

    const defaultDeviceForm = () => ({
      bk_biz_id: [] as number[],
      order_id: [] as any[],
      suborder_id: [] as any[],
      ip: [] as any[],
      device_type: [] as any[],
      bk_zone_name: [] as any[],
      sub_zone: [] as any[],
      stage: [] as any[],
      bk_username: [userStore.username],
      dateRange: applicationTime(),
      bk_asset_id: [] as string[],
    });

    const { formModel, resetForm } = useFormModel(defaultDeviceForm());

    const { selections, handleSelectionChange } = useSelection();

    const deviceTypeList = ref([]);
    const bkZoneNameList = ref([]);
    const subZoneList = ref([]);
    const stageList = ref([]);

    const requestListParams = computed(() => {
      const params = {
        ...formModel,
        start: formModel.dateRange[0],
        end: formModel.dateRange[1],
        bk_biz_id: [getBizsId()],
      };
      params.order_id = params.order_id.length ? params.order_id.map((v) => +v) : [];
      params.dateRange = undefined;
      removeEmptyFields(params);
      return params;
    });

    const { columns } = useScrColumns('hostRecycleDevice');
    columns.splice(1, 0, {
      label: t('子单号'),
      field: 'suborder_id',
      width: 80,
      render: ({ row }: any) => {
        return (
          // 单据详情
          <span class={scrCssModule['sub-order-num']} onClick={() => enterDetail(row)}>
            {row.suborder_id}
          </span>
        );
      },
    });
    const { CommonTable, getListData, pagination, isLoading } = useTable({
      tableOptions: {
        columns,
        extra: {
          onSelect: (selections: any) => {
            handleSelectionChange(selections, () => true, false);
          },
          onSelectAll: (selections: any) => {
            handleSelectionChange(selections, () => true, true);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: { sort: 'ip', order: 'ASC' },
        immediate: false,
      },
      scrConfig: () => {
        return {
          url: `/api/v1/woa/${getBusinessApiPath()}task/findmany/recycle/host`,
          payload: {
            ...requestListParams.value,
          },
        };
      },
    });
    const enterDetail = (row: any) => {
      router.push({
        name: 'HostRecycleDocDetail',
        query: { ...route.query, suborderId: row.suborder_id, bkBizId: getBizsId() },
      });
    };

    const fetchDeviceTypeList = async () => {
      const data = await getDeviceTypeList();
      deviceTypeList.value = data?.info || [];
    };
    const fetchRegionList = async () => {
      const data = await getRegionList();
      bkZoneNameList.value = data?.info || [];
    };
    const fetchZoneList = async () => {
      const data = await getZoneList();
      subZoneList.value = data?.info || [];
    };
    const fetchStageList = async () => {
      const data = await getRecycleStageOpts();
      stageList.value = data?.info || [];
    };

    onMounted(() => {
      fetchDeviceTypeList();
      fetchRegionList();
      fetchZoneList();
      fetchStageList();
    });

    const searchRulesKey = 'host_recycle_device_rules';
    const filterOrders = () => {
      // 回填
      formModel.bk_biz_id = [getBizsId()];
      pagination.start = 0;
      getListData();
    };

    const { saveSearchRules, clearSearchRules } = useSaveSearchRules(searchRulesKey, filterOrders, formModel);

    const handleSearch = () => {
      // update query
      saveSearchRules();
    };

    const handleReset = () => {
      resetForm(defaultDeviceForm());
      formModel.bk_biz_id = [getBizsId()];
      // update query
      clearSearchRules();
    };

    watch(
      () => userStore.username,
      (username) => {
        if (route.query[searchRulesKey]) return;
        // 无搜索记录，设置申请人默认值
        formModel.bk_username = [username];
      },
    );

    // 导出全部的请求函数
    const exportAllRequest = async (signal: AbortSignal) => {
      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'enable_count',
      }).rollReqUseTotalCount(
        `/api/v1/woa/${getBusinessApiPath()}task/findmany/recycle/host`,
        {
          ...requestListParams.value,
        },
        {
          limit: 5000,
          total: pagination.count,
          listGetter: (res: { data: { info: any[] } }) => res.data.info,
          countGetter: (res: { data: { count: number } }) => res.data.count,
        },
        {
          signal,
        },
      );

      return list;
    };

    return () => (
      <>
        <GridFilterComp
          onSearch={handleSearch}
          onReset={handleReset}
          loading={isLoading.value}
          col={5}
          rules={[
            {
              title: t('单号'),
              content: <FloatInput v-model={formModel.order_id} placeholder={t('请输入单号，多个换行分割')} />,
            },
            {
              title: t('子单号'),
              content: <FloatInput v-model={formModel.suborder_id} placeholder={t('请输入子单号，多个换行分割')} />,
            },
            {
              title: t('机型'),
              content: (
                <Select v-model={formModel.device_type} multiple clearable placeholder={t('请选择机型')}>
                  {deviceTypeList.value.map((item) => {
                    return <Select.Option key={item} name={item} id={item} />;
                  })}
                </Select>
              ),
            },
            {
              title: t('地域'),
              content: (
                <Select v-model={formModel.bk_zone_name} multiple clearable placeholder={t('请选择地域')}>
                  {bkZoneNameList.value.map((item) => {
                    return <Select.Option key={item} name={item} id={item} />;
                  })}
                </Select>
              ),
            },
            {
              title: t('园区'),
              content: (
                <Select v-model={formModel.sub_zone} multiple clearable placeholder={t('请选择园区')}>
                  {subZoneList.value.map((item) => {
                    return <Select.Option key={item} name={item} id={item} />;
                  })}
                </Select>
              ),
            },
            {
              title: t('状态'),
              content: (
                <Select v-model={formModel.stage} multiple clearable placeholder={t('请选择状态')}>
                  {stageList.value.map(({ stage, description }) => {
                    return <Select.Option key={stage} name={description} id={stage} />;
                  })}
                </Select>
              ),
            },
            {
              title: t('回收IP'),
              content: <FloatInput v-model={formModel.ip} placeholder={t('请输入IP，多个换行分割')} />,
            },
            {
              title: t('回收人'),
              content: <hcm-form-user v-model={formModel.bk_username} />,
            },
            {
              title: t('完成时间'),
              content: <ScrDatePicker class='full-width' v-model={formModel.dateRange} />,
            },
            {
              title: t('固资号'),
              content: <FloatInput v-model={formModel.bk_asset_id} placeholder={t('请输入固资号，多个换行分割')} />,
            },
          ]}
        />
        <section class={cssModule['table-wrapper']}>
          <div class={[cssModule.buttons, cssModule.mb16]}>
            <ExportToExcelBatchButton
              class={cssModule.button}
              showConfirmDialog
              data={selections.value}
              columns={columns}
              text={t('导出勾选')}
              filename={t('回收设备列表')}
              name='回收主机'
              disabled={selections.value.length === 0}
            />
            <ExportToExcelBatchButton
              showConfirmDialog
              request={exportAllRequest}
              columns={columns}
              filename='回收设备列表'
              text='导出全部'
              name='回收主机'
              pickNum={pagination.count}
              disabled={pagination.count === 0}
            />
          </div>
          <CommonTable style={{ height: 'calc(100% - 48px)' }} />
        </section>
      </>
    );
  },
});
