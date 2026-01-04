import { computed, defineComponent, watch } from 'vue';
import cssModule from './index.module.scss';

import { Button, Input, TagInput } from 'bkui-vue';
import ScrCreateFilterSelector from '@/views/ziyanScr/resource-manage/create/ScrCreateFilterSelector';
import ExportToExcelBatchButton from '@/components/export-to-excel-batch-button/index.vue';
import GridFilterComp from '@/components/grid-filter-comp';
import ScrDatePicker from '@/components/scr/scr-date-picker';

import { useI18n } from 'vue-i18n';
import { useUserStore, useZiyanScrStore } from '@/store';
import useFormModel from '@/hooks/useFormModel';
import { useTable } from '@/hooks/useTable/useTable';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { transferSimpleConditions } from '@/utils/scr/simple-query-builder';
import rollRequest from '@blueking/roll-request';
import http from '@/http';
import { applicationTime } from '@/common/util';
import { useRoute } from 'vue-router';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useSaveSearchRules } from '@/views/ticket/utils/useSaveSearchRules';
import useScrColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';

export default defineComponent({
  setup() {
    const { getBusinessApiPath, getBizsId } = useWhereAmI();
    const { t } = useI18n();
    const userStore = useUserStore();
    const scrStore = useZiyanScrStore();
    const route = useRoute();

    const { formModel, resetForm } = useFormModel({
      orderId: '',
      bkBizId: [],
      bkUsername: [userStore.username],
      ip: [],
      requireType: '',
      suborderId: '',
      dateRange: applicationTime(),
      assetId: [],
    });

    const { selections, handleSelectionChange } = useSelection();

    const clipHostIp = computed(() => {
      return selections.value.map((item) => item.ip).join('\n');
    });
    const clipHostAssetId = computed(() => {
      return selections.value.map((item) => item.asset_id).join('\n');
    });

    const { columns } = useScrColumns('hostApplyDevice');

    const { CommonTable, getListData, isLoading, pagination } = useTable({
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
        sortOption: {
          sort: 'create_at',
          order: 'DESC',
        },
        immediate: false,
      },
      scrConfig: () => {
        return {
          url: `/api/v1/woa/${getBusinessApiPath()}task/findmany/apply/device`,
          payload: {
            filter: transferSimpleConditions([
              'AND',
              ['bk_biz_id', 'in', [getBizsId()]],
              ['require_type', '=', formModel.requireType],
              ['order_id', '=', formModel.orderId],
              ['suborder_id', '=', formModel.suborderId],
              ['bk_username', 'in', formModel.bkUsername],
              ['ip', 'in', formModel.ip],
              ['update_at', 'd>=', formModel.dateRange[0]],
              ['update_at', 'd<=', formModel.dateRange[1]],
              ['asset_id', 'in', formModel.assetId],
            ]),
            page: { start: 0, limit: 10 },
          },
        };
      },
    });

    const searchRulesKey = 'host_apply_device_rules';
    const filterOrders = () => {
      pagination.start = 0;
      getListData();
    };
    const { saveSearchRules, clearSearchRules } = useSaveSearchRules(searchRulesKey, filterOrders, formModel);

    // 构建查询条件的函数
    const buildFilterPayload = () => ({
      filter: transferSimpleConditions([
        'AND',
        ['bk_biz_id', 'in', [getBizsId()]],
        ['require_type', '=', formModel.requireType],
        ['order_id', '=', formModel.orderId],
        ['suborder_id', '=', formModel.suborderId],
        ['bk_username', 'in', formModel.bkUsername],
        ['ip', 'in', formModel.ip],
        ['update_at', 'd>=', formModel.dateRange[0]],
        ['update_at', 'd<=', formModel.dateRange[1]],
        ['asset_id', 'in', formModel.assetId],
      ]),
    });

    // 导出全部的请求函数
    const exportAllRequest = async (signal: AbortSignal) => {
      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'enable_count',
      }).rollReqUseTotalCount(
        `/api/v1/woa/${getBusinessApiPath()}task/findmany/apply/device`,
        {
          ...buildFilterPayload(),
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

    const handleSearch = () => {
      // update query
      saveSearchRules();
    };

    const handleReset = () => {
      resetForm({ bkUsername: [userStore.username] });
      // update query
      clearSearchRules();
    };

    watch(
      () => userStore.username,
      (username) => {
        if (route.query[searchRulesKey]) return;
        // 无搜索记录，设置申请人默认值
        formModel.bkUsername = [username];
      },
    );

    return () => (
      <>
        <GridFilterComp
          onSearch={handleSearch}
          onReset={handleReset}
          loading={isLoading.value}
          rules={[
            {
              title: t('需求类型'),
              content: (
                <ScrCreateFilterSelector
                  v-model={formModel.requireType}
                  api={scrStore.getRequirementList}
                  multiple={false}
                  optionIdPath='require_type'
                  optionNamePath='require_name'
                />
              ),
            },
            {
              title: t('单号'),
              content: <Input v-model={formModel.orderId} clearable type='number' placeholder='请输入单号' />,
            },
            {
              title: t('申请人'),
              content: <hcm-form-user v-model={formModel.bkUsername} />,
            },
            {
              title: t('交付时间'),
              content: <ScrDatePicker class='full-width' v-model={formModel.dateRange} clearable={false} />,
            },
            {
              title: t('内网IP'),
              content: (
                <TagInput
                  v-model={formModel.ip}
                  allow-create
                  collapse-tags
                  allow-auto-match
                  pasteFn={(v) => v.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag }))}
                  createTagValidator={(ip) =>
                    /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/.test(ip)
                  }
                  placeholder='输入合法的 IP 地址'
                />
              ),
            },
            {
              title: t('固资号'),
              content: (
                <TagInput
                  v-model={formModel.assetId}
                  allow-create
                  collapse-tags
                  allow-auto-match
                  pasteFn={(v) => v.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag }))}
                  placeholder='请输入固资号'
                />
              ),
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
              filename='设备列表'
              text='导出勾选'
              name='申领主机'
              disabled={selections.value.length === 0}
            />
            <ExportToExcelBatchButton
              class={cssModule.button}
              showConfirmDialog
              request={exportAllRequest}
              columns={columns}
              filename='设备列表'
              text='导出全部'
              name='申领主机'
              pickNum={pagination.count}
              disabled={pagination.count === 0}
            />
            <Button class={cssModule.button} v-clipboard={clipHostIp.value} disabled={selections.value.length === 0}>
              复制IP
            </Button>
            <Button
              class={cssModule.button}
              v-clipboard={clipHostAssetId.value}
              disabled={selections.value.length === 0}>
              复制固单号
            </Button>
          </div>
          <CommonTable style={{ height: 'calc(100% - 48px)' }} />
        </section>
      </>
    );
  },
});
