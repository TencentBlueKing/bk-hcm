<script setup lang="ts">
import { reactive, ref, useTemplateRef, watch } from 'vue';
import { type LocationQuery, useRoute } from 'vue-router';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { IRollingServerCpuCoreSummary, RollingServerRecordItem, useRollingServerUsageStore } from '@/store';
import { useRollingServerStore } from '@/store/rolling-server';
import { convertDateRangeToObject, getDateRange, transformSimpleCondition } from '@/utils/search';
import usageOrderViewProperties from '@/model/rolling-server/usage-order.view';
import { ISearchCondition, IView } from '../typings';

import Search from './children/search.vue';
import DataList from './children/data-list.vue';
import ReturnedRecordsDialog from './children/returned-records-dialog.vue';
import NotNoticeDialog from './children/not-notice-dialog.vue';
import OrderConfigDialog from './children/order-config-dialog.vue';

const route = useRoute();
const rollingServerStore = useRollingServerStore();
const rollingServerUsageStore = useRollingServerUsageStore();

const { isBusinessPage, getBizsId } = useWhereAmI();
const searchQs = useSearchQs({ key: 'filter', properties: usageOrderViewProperties });
const { pagination, getPageParams } = usePage();

const docList = ref<RollingServerRecordItem[]>();
const summaryInfo = ref<IRollingServerCpuCoreSummary>();
const condition = ref<Record<string, any>>({});

const getList = async (query: LocationQuery) => {
  condition.value = searchQs.get(query, {
    roll_date: isBusinessPage ? getDateRange('last120d', true) : getDateRange('last30d'),
    suborder_id: [],
    bk_biz_id: [],
  });
  const { roll_date, bk_biz_id = [], suborder_id } = condition.value;
  let bk_biz_ids = bk_biz_id.length === 1 && bk_biz_id[0] === -1 ? undefined : bk_biz_id;
  // 业务下需要传入当前业务用于查询
  if (isBusinessPage) bk_biz_ids = [getBizsId()];
  const { start, end } = convertDateRangeToObject(roll_date);

  pagination.current = Number(query.page) || 1;
  pagination.limit = Number(query.limit) || pagination.limit;

  const sort = (query.sort || 'roll_date') as string;
  const order = (query.order || 'DESC') as string;

  // 请求申请单据列表
  const [listRes, summaryRes] = await Promise.all([
    rollingServerUsageStore.getAppliedRecordList({
      filter: transformSimpleCondition(
        { ...condition.value, bk_biz_id: bk_biz_ids, require_type: 6 },
        usageOrderViewProperties,
      ),
      page: getPageParams(pagination, { sort, order }),
    }),
    rollingServerUsageStore.getCpuCoreSummary({ start, end, bk_biz_ids, suborder_ids: suborder_id, require_type: 6 }),
  ]);
  const { list: appliedRecordList, count } = listRes;

  // 只查询非资源池业务的退还记录
  const applied_record_id = appliedRecordList
    .filter((item) => !rollingServerStore.resPollBusinessIds.includes(item.bk_biz_id))
    .map((item) => item.id); // 申请单id与退还记录单applied_record_id一一对应

  // 如果当页appliedRecordList中存有普通业务，则需请求对应的退还记录列表
  if (applied_record_id.length > 0) {
    const returnedRecordList = await rollingServerUsageStore.getReturnedRecordList({
      // 查询正常状态的回收记录
      filter: transformSimpleCondition({ applied_record_id, status: 2 }, usageOrderViewProperties),
    });
    // 设置列表
    docList.value = appliedRecordList.map((appliedRecordItem) => {
      // 资源池业务
      if (rollingServerStore.resPollBusinessIds.includes(appliedRecordItem.bk_biz_id)) {
        return { ...appliedRecordItem, isResPollBusiness: true };
      }
      // 普通业务
      const returned_records = returnedRecordList.filter(
        (returnedRecordItem) => returnedRecordItem.applied_record_id === appliedRecordItem.id,
      );
      // 前端计算字段：returned_core, not_returned_core, exec_rate
      const returned_core = returned_records.reduce((acc, cur) => acc + cur.match_applied_core, 0);
      const not_returned_core = appliedRecordItem.delivered_core - returned_core;
      // exempted_returned_core 直接从申请记录获取（后端接口返回）
      const exempted_returned_core = appliedRecordItem.exempted_returned_core || 0;
      const not_returned_core_after_exempted = not_returned_core - exempted_returned_core;
      const exec_rate = `${Number(((returned_core / (appliedRecordItem.delivered_core || 1)) * 100).toFixed(2))}%`; // 结果保留两位小数，不显示多余0

      return {
        ...appliedRecordItem,
        returned_records,
        returned_core,
        not_returned_core,
        exempted_returned_core,
        not_returned_core_after_exempted,
        exec_rate,
      };
    });
  } else {
    docList.value = appliedRecordList.map((item) => ({ ...item, isResPollBusiness: true }));
  }

  // 设置汇总信息
  summaryInfo.value = summaryRes;

  // 设置页面总条数
  pagination.count = count;
};

const recordsPoll = useTimeoutPoll(
  () => {
    getList(route.query);
  },
  30000,
  { immediate: true },
);

watch(
  () => route.query,
  () => {
    recordsPoll.pause();
    recordsPoll.resume();
  },
);

const handleSearch = (vals: ISearchCondition) => {
  searchQs.set(vals);
};

const handleReset = () => {
  searchQs.clear();
};

const returnedRecordsDialogRef = useTemplateRef<typeof ReturnedRecordsDialog>('returned-records-dialog');
const handleReturnedRecords = (id: string) => {
  returnedRecordsDialogRef.value.show(docList.value.find((item) => item.id === id));
};

const notNoticeDialogState = reactive({ isHidden: true, isShow: false, details: null });
const handleShowNotNotice = async (details: RollingServerRecordItem) => {
  recordsPoll.pause();
  Object.assign(notNoticeDialogState, {
    isHidden: false,
    isShow: true,
    details: { ...details, roll_date: String(details.roll_date) },
  });
};
const handleNotNoticeSuccess = () => {
  recordsPoll.resume();
};
const handleNotNoticeHidden = () => {
  Object.assign(notNoticeDialogState, { isHidden: true, details: null });
};

// 单据配置对话框
const orderConfigDialogState = reactive({ isHidden: true, isShow: false, details: null });
const handleShowOrderConfig = (details: RollingServerRecordItem) => {
  recordsPoll.pause();
  Object.assign(orderConfigDialogState, {
    isHidden: false,
    isShow: true,
    details,
  });
};
const handleOrderConfigSuccess = () => {
  recordsPoll.resume();
};
const handleOrderConfigHidden = () => {
  Object.assign(orderConfigDialogState, { isHidden: true, details: null });
};
</script>

<template>
  <search :view="IView.ORDER" :condition="condition" @search="handleSearch" @reset="handleReset" />
  <data-list
    v-bkloading="{ loading: rollingServerUsageStore.appliedRecordsListLoading }"
    :view="IView.ORDER"
    :list="docList"
    :pagination="pagination"
    :summary-info="summaryInfo"
    @show-returned-records="handleReturnedRecords"
    @show-not-notice="handleShowNotNotice"
    @show-order-config="handleShowOrderConfig"
  />
  <returned-records-dialog ref="returned-records-dialog" />
  <template v-if="!notNoticeDialogState.isHidden">
    <not-notice-dialog
      v-model="notNoticeDialogState.isShow"
      :details="notNoticeDialogState.details"
      @success="handleNotNoticeSuccess"
      @hidden="handleNotNoticeHidden"
    />
  </template>
  <template v-if="!orderConfigDialogState.isHidden">
    <order-config-dialog
      v-model="orderConfigDialogState.isShow"
      :details="orderConfigDialogState.details"
      @success="handleOrderConfigSuccess"
      @hidden="handleOrderConfigHidden"
    />
  </template>
</template>
