<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Message, SearchSelect } from 'bkui-vue';
import moment from 'moment';
import http from '@/http';
import { useAccountStore, useResourceStore } from '@/store';
import { useAccountSelectorStore } from '@/store/account-selector';
import { useRegionsStore } from '@/store/useRegionsStore';
import { QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { timeFormatter } from '@/common/util';
import { AUTH_MANAGE_RECYCLE_BIN } from '@/constants/auth-symbols';
import {
  MENU_BUSINESS_DRIVE_DETAILS,
  MENU_BUSINESS_HOST_DETAILS,
  MENU_BUSINESS_RECYCLEBIN,
  MENU_RESOURCE_DETAIL,
  MENU_RESOURCE_RECYCLEBIN,
} from '@/constants/menu-symbol';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import routerAction from '@/router/utils/action';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';

type RecycleType = 'cvm' | 'disk';

const { t } = useI18n();
const route = useRoute();
const resourceStore = useResourceStore();
const accountStore = useAccountStore();
const accountSelectorStore = useAccountSelectorStore();
const { getRegionName } = useRegionsStore();
const { whereAmI } = useWhereAmI();

// 业务回收站与资源回收管理共用本页：业务侧不按账号筛选、不配置保留时间、不提供单行操作
const isBusiness = computed(() => whereAmI.value === Senarios.business);
const listRouteName = computed(() => (isBusiness.value ? MENU_BUSINESS_RECYCLEBIN : MENU_RESOURCE_RECYCLEBIN));

const recycleTypeData: { name: string; value: RecycleType }[] = [
  { name: t('主机回收'), value: 'cvm' },
  { name: t('硬盘回收'), value: 'disk' },
];
const RESERVE_TIME_SET = new Array(8)
  .fill(0)
  .map((_val, idx) => idx)
  .map((num) => ({ label: `${num}${num > 0 ? '天' : ''}`, value: num * 24 }));

const recycleManageSign = { type: AUTH_MANAGE_RECYCLE_BIN };

const filter = reactive({
  op: 'and',
  rules: [
    { field: 'res_type', op: 'eq', value: 'cvm' },
    { field: 'status', op: 'eq', value: 'wait_recycle' },
  ] as { field: string; op: string; value: any }[],
});
const fetchUrl = ref('recycle_records/list');
const selectedType = ref<RecycleType>('cvm');
const searchVal = ref([]);

const { datas, isLoading, pagination, handlePageSizeChange, handlePageChange, getList, handleSort } =
  useQueryCommonList({ filter }, fetchUrl);
const { selections, handleSelectionChange, resetSelections } = useSelection();

const searchData = computed(() => [
  { name: `${selectedType.value === 'cvm' ? '主机' : '硬盘'}ID`, id: 'cloud_res_id' },
]);

// 账号以地址为准：选择器只改 query，列表条件与回收保留时间都跟随 accountId
const accountId = computed(() => {
  const id = route.query.accountId;
  return !isBusiness.value && typeof id === 'string' && id ? id : '';
});
const accountOptions = computed(() => accountSelectorStore.authorizedResourceAccountList);

const setAccountId = (val?: string) => {
  routerAction.redirect(
    { name: listRouteName.value, params: route.params, query: { ...route.query, accountId: val || undefined } },
    { replace: true },
  );
};

const setRule = (field: string, rule?: { field: string; op: string; value: any }) => {
  const idx = filter.rules.findIndex((item) => item.field === field);
  if (!rule) {
    if (idx > -1) filter.rules.splice(idx, 1);
    return;
  }
  if (idx === -1) filter.rules.push(rule);
  else filter.rules[idx] = rule;
};

// 回收保留时间（小时），0 也是合法值
const savedReserveTime = ref<number>(null);
const reserveTimeLabel = computed(
  () => RESERVE_TIME_SET.find((item) => item.value === savedReserveTime.value)?.label ?? '--',
);
const isSettingDialogShow = ref(false);
const isSettingDialogLoading = ref(false);
const recycleReserveTime = ref(48);

watch(
  accountId,
  async (id) => {
    setRule('account_id', id ? { field: 'account_id', op: QueryRuleOPEnum.EQ, value: id } : undefined);
    savedReserveTime.value = null;
    if (!id) return;
    try {
      const res = await accountStore.getAccountDetail(id);
      if (id !== accountId.value) return;
      savedReserveTime.value = res?.data?.recycle_reserve_time ?? null;
    } catch {
      // 账号不存在或无权限：回到全部账号
      if (id === accountId.value) setAccountId();
    }
  },
  { immediate: true },
);

const handleSettingShow = () => {
  recycleReserveTime.value = savedReserveTime.value ?? 48;
  isSettingDialogShow.value = true;
};

const handleSettingConfirm = async () => {
  isSettingDialogLoading.value = true;
  try {
    await http.patch(`/api/v1/cloud/accounts/${accountId.value}`, {
      account_id: accountId.value,
      recycle_reserve_time: recycleReserveTime.value,
    });
    savedReserveTime.value = recycleReserveTime.value;
    Message({ theme: 'success', message: '配置成功' });
    isSettingDialogShow.value = false;
  } finally {
    isSettingDialogLoading.value = false;
  }
};

// 随主机回收的硬盘不可单独操作；资源侧另外不可操作已分配到业务的资源
const isCurRowSelectEnable = (row: any) =>
  !(row?.recycle_type === 'related' || (!isBusiness.value && row?.bk_biz_id !== -1));
const isRowSelectEnable = ({ row, isCheckAll }: any) => isCheckAll || isCurRowSelectEnable(row);

const generateTooltipsOptions = (data: any) => {
  if (data?.recycle_type === 'related') {
    return { content: '该硬盘随主机回收，不可单独操作', disabled: false };
  }
  if (data?.bk_biz_id !== -1) {
    return { content: '该硬盘仅可在业务下操作', disabled: false };
  }
  return { disabled: true };
};

watch(
  () => route.query.type,
  (type) => {
    selectedType.value = type === 'disk' ? 'disk' : 'cvm';
  },
  { immediate: true },
);

watch(
  selectedType,
  (type, oldType) => {
    setRule('res_type', { field: 'res_type', op: 'eq', value: type });
    resetSelections();
    if (route.query.type === type) return;
    // 手动切页签：带上类型，清掉所属主机跳转带来的 cvm 条件
    if (oldType) searchVal.value = [];
    routerAction.redirect(
      { name: listRouteName.value, params: route.params, query: { ...route.query, type, cvm: undefined } },
      { replace: true },
    );
  },
  { immediate: true },
);

watch(
  () => route.query.cvm,
  (cvm) => {
    if (Array.isArray(cvm)) return;
    searchVal.value = cvm ? [{ id: 'cloud_res_id', name: 'ID', values: [{ id: cvm, name: cvm }] }] : [];
  },
  { immediate: true },
);

watch(
  searchVal,
  (vals) => {
    filter.rules = filter.rules
      .filter(({ field }) => field !== 'cloud_res_id')
      .concat(vals.map((val: any) => ({ field: val.id, op: QueryRuleOPEnum.EQ, value: val.values[0].id })));
  },
  { immediate: true },
);

// 硬盘的所属主机：切到主机页签并按主机 ID 搜索
const handleLink = (cvmId: string) => {
  routerAction.redirect(
    { name: listRouteName.value, params: route.params, query: { ...route.query, type: 'cvm', cvm: cvmId } },
    { replace: true },
  );
};

const handleClick = (vendor: VendorEnum, id: string, type: 'drive' | 'host') => {
  const routeInfo = isBusiness.value
    ? {
        name: type === 'host' ? MENU_BUSINESS_HOST_DETAILS : MENU_BUSINESS_DRIVE_DETAILS,
        params: { id },
        query: { vendor },
      }
    : {
        name: MENU_RESOURCE_DETAIL,
        params: { resourceType: type, id },
        query: { accountId: route.query.accountId, vendor },
      };
  routerAction.redirect(routeInfo, { history: true });
};

const operateState = reactive({
  isShow: false,
  title: '',
  type: '' as 'destroy' | 'recover' | '',
  ids: [] as string[],
});

const handleOperate = (type: 'destroy' | 'recover', ids?: string[]) => {
  operateState.ids = ids ?? selections.value.map((item: any) => item.id);
  operateState.type = type;
  operateState.title = `确认要 ${type === 'destroy' ? t('销毁') : t('恢复')}`;
  operateState.isShow = true;
};

const handleOperateConfirm = async () => {
  const params = { record_ids: operateState.ids };
  try {
    if (operateState.type === 'destroy') {
      await resourceStore.deleteRecycledData(`${selectedType.value}s`, params);
    } else {
      await resourceStore.recoverRecycledData(`${selectedType.value}s`, params);
    }
    const operate = operateState.type === 'destroy' ? t('销毁') : t('恢复');
    const resourceName = selectedType.value === 'cvm' ? t('主机') : t('硬盘');
    Message({ message: `${operate}${resourceName}成功`, theme: 'success' });
    getList();
  } finally {
    operateState.isShow = false;
  }
};

onMounted(() => {
  if (!isBusiness.value) {
    accountSelectorStore.getAuthorizedResourceAccountList({
      filter: { op: 'and', rules: [{ field: 'type', op: QueryRuleOPEnum.EQ, value: 'resource' }] },
    });
  }
  moment.updateLocale('zh-cn', {
    relativeTime: {
      future: '余%s',
      past: '%s前',
      s: '%d秒',
      ss: '%d秒',
      m: '1分钟',
      mm: '%d分钟',
      h: '1小时',
      hh: '%d小时',
      d: '1天',
      dd: '%d天',
      w: '1周',
      ww: '%d周',
      M: '1月',
      MM: '%d月',
      y: '1年',
      yy: '%d年',
    },
  });
});
</script>

<template>
  <div class="page-container recycle-manager-page">
    <bk-tab type="card-grid" v-model:active="selectedType" :label-height="42">
      <template v-if="!isBusiness" #setting>
        <div class="setting-container">
          <bk-select
            class="account-select"
            :model-value="accountId"
            filterable
            clearable
            placeholder="全部账号"
            :loading="accountSelectorStore.authorizedResourceAccountLoading"
            @change="setAccountId"
          >
            <bk-option v-for="item in accountOptions" :key="item.id" :value="item.id" :label="item.name" />
          </bk-select>
          <template v-if="accountId">
            <span class="reserve-time">回收保留时间：{{ reserveTimeLabel }}</span>
            <div class="setting-icon-container" @click="handleSettingShow">
              <i class="hcm-icon bkhcm-icon-shezhi"></i>
            </div>
          </template>
        </div>
      </template>
      <bk-tab-panel v-for="item in recycleTypeData" :key="item.value" :name="item.value" :label="item.name">
        <section class="header-container">
          <hcm-auth :sign="recycleManageSign" v-slot="{ noPerm }">
            <span
              v-bk-tooltips="{
                content: `请勾选${selectedType === 'cvm' ? '主机' : '硬盘'}信息`,
                disabled: selections.length,
              }"
            >
              <bk-button :disabled="!selections.length || noPerm" @click="handleOperate('destroy')">
                {{ t('立即销毁') }}
              </bk-button>
            </span>
            <span
              v-bk-tooltips="{
                content: `请勾选${selectedType === 'cvm' ? '主机' : '硬盘'}信息`,
                disabled: selections.length,
              }"
            >
              <bk-button class="ml8" :disabled="!selections.length || noPerm" @click="handleOperate('recover')">
                {{ t('立即恢复') }}
              </bk-button>
            </span>
          </hcm-auth>
          <SearchSelect
            class="w500 common-search-selector"
            v-model="searchVal"
            :data="searchData"
            value-behavior="need-key"
          />
        </section>
        <bk-loading :loading="isLoading" opacity="1">
          <bk-table
            :key="selectedType"
            class="table-layout"
            :data="datas"
            remote-pagination
            :pagination="pagination"
            @page-value-change="handlePageChange"
            @page-limit-change="handlePageSizeChange"
            @selection-change="(rows: any) => handleSelectionChange(rows, isCurRowSelectEnable)"
            @select-all="(rows: any) => handleSelectionChange(rows, isCurRowSelectEnable, true)"
            row-hover="auto"
            row-key="id"
            :is-row-select-enable="isRowSelectEnable"
            show-overflow-tooltip
            @column-sort="handleSort"
          >
            <bk-table-column width="30" min-width="30" type="selection" />
            <bk-table-column :label="`${selectedType === 'cvm' ? '主机' : '硬盘'}ID`" prop="cloud_res_id" sort>
              <template #default="{ data }">
                <bk-button
                  theme="primary"
                  text
                  @click="handleClick(data?.vendor, data?.res_id, selectedType === 'cvm' ? 'host' : 'drive')"
                >
                  {{ data?.cloud_res_id }}
                </bk-button>
              </template>
            </bk-table-column>
            <bk-table-column label="名称" prop="res_name">
              <template #default="{ cell }">
                {{ cell || '--' }}
              </template>
            </bk-table-column>
            <bk-table-column label="所属主机" prop="detail_cvm_id" v-if="selectedType === 'disk'">
              <template #default="{ data }">
                {{ data?.detail?.cvm_id || '--' }}
                <i
                  class="hcm-icon bkhcm-icon-link related-cvm-link"
                  v-if="data?.detail?.cvm_id"
                  @click="handleLink(data.detail.cvm_id)"
                />
              </template>
            </bk-table-column>
            <bk-table-column :label="t('地域')" prop="region">
              <template #default="{ data }">
                {{ getRegionName(data?.vendor, data?.region) }}
              </template>
            </bk-table-column>
            <bk-table-column label="回收人" prop="reviser"></bk-table-column>
            <bk-table-column label="进入回收站时间" prop="created_at" :sort="true">
              <template #default="{ cell }">
                {{ timeFormatter(cell) }}
              </template>
            </bk-table-column>
            <bk-table-column label="过期时间" prop="recycled_at" :sort="true">
              <template #default="{ cell }">
                <bk-tag theme="danger">
                  {{ moment(cell).fromNow() }}
                </bk-tag>
                {{ timeFormatter(cell) }}
              </template>
            </bk-table-column>
            <bk-table-column v-if="!isBusiness" :label="t('操作')" :min-width="150">
              <template #default="{ data }">
                <hcm-auth :sign="recycleManageSign" tag="span" v-slot="{ noPerm }">
                  <bk-button
                    text
                    class="mr10"
                    theme="primary"
                    @click="handleOperate('destroy', [data.id])"
                    v-bk-tooltips="generateTooltipsOptions(data)"
                    :disabled="noPerm || !isCurRowSelectEnable(data)"
                  >
                    销毁
                  </bk-button>
                  <bk-button
                    text
                    theme="primary"
                    @click="handleOperate('recover', [data.id])"
                    v-bk-tooltips="generateTooltipsOptions(data)"
                    :disabled="noPerm || !isCurRowSelectEnable(data)"
                  >
                    恢复
                  </bk-button>
                </hcm-auth>
              </template>
            </bk-table-column>
          </bk-table>
        </bk-loading>
      </bk-tab-panel>
    </bk-tab>

    <bk-dialog
      :is-show="operateState.isShow"
      :title="operateState.title"
      theme="primary"
      :quick-close="false"
      @closed="operateState.isShow = false"
      @confirm="handleOperateConfirm"
    >
      <div v-if="operateState.type === 'destroy'">
        {{ selectedType === 'cvm' ? '销毁之后无法恢复主机信息' : '销毁之后无法从云上恢复硬盘' }}
      </div>
      <div v-else>{{ t(`将恢复${selectedType === 'cvm' ? '主机' : '硬盘'}信息`) }}</div>
    </bk-dialog>

    <bk-dialog
      :is-show="isSettingDialogShow"
      title="回收站配置"
      theme="primary"
      :is-loading="isSettingDialogLoading"
      @closed="isSettingDialogShow = false"
      @confirm="handleSettingConfirm"
    >
      保留时长
      <bk-select v-model="recycleReserveTime" class="mt6" :clearable="false">
        <bk-option v-for="item in RESERVE_TIME_SET" :key="item.value" :value="item.value" :label="item.label" />
      </bk-select>
    </bk-dialog>
  </div>
</template>

<style lang="scss" scoped>
.recycle-manager-page {
  :deep(.bk-tab) {
    height: 100%;

    .bk-tab-header-item {
      height: 42px;
    }

    .bk-tab-content {
      padding: 16px 24px;
      height: calc(100% - 42px);

      .bk-nested-loading {
        height: calc(100% - 48px);

        .bk-table {
          max-height: 100%;
        }
      }
    }
  }
}

.setting-container {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 100%;
  padding-right: 12px;

  .account-select {
    width: 240px;
  }

  .reserve-time {
    font-size: 12px;
    color: #63656e;
    white-space: nowrap;
  }
}

.setting-icon-container {
  width: 32px;
  height: 32px;
  background: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;
  border-radius: 2px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
}

.mt6 {
  margin-top: 6px;
}

.related-cvm-link {
  margin-left: 4px;
  cursor: pointer;
  color: #3a84ff;
}

.header-container {
  display: flex;
  justify-content: space-between;
}
</style>
