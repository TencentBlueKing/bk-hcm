<script lang="ts" setup>
import BusinessSelector from '@/components/business-selector/index.vue';
import AccountSelector from '@/components/account-selector/index.vue';
import MemberSelect from '@/components/MemberSelect';
import FilterItemAction from './children/filter-item-action.vue';
import AuditDetail from './detail.vue';
import ErrorPages from '@/views/error-pages/403.tsx';

import { computed, reactive, ref, watch, h } from 'vue';
import dayjs from 'dayjs';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import useList from './use-list';
import { AUDIT_RESOURCE_TYPES } from '@/common/constant';
import { timeFormatter } from '@/common/util';
import { AUDIT_SOURCE_MAP, AUDIT_ACTION_MAP } from './constants';
import { Button } from 'bkui-vue';

import { useVerify } from '@/hooks';

const { t } = useI18n();
const route = useRoute();

const businessSelectorComp = ref(null);

const tabs = [
  {
    type: 'biz',
    label: '业务',
  },
  {
    type: 'resource',
    label: '资源',
  },
];

const todayStart = dayjs(new Date()).format('YYYY-MM-DD 00:00:00');
const todayEnd = dayjs(new Date()).format('YYYY-MM-DD 23:59:59');
const defaultFilter = () => ({
  bk_biz_id: null as number,
  account_id: [] as any,
  res_type: 'account',
  action: '',
  created_at: [todayStart, todayEnd],
  operator: [] as any[],
  res_id: '',
  res_name: '',
  source: '',
});

let filter = reactive(defaultFilter());
const filterOptions = reactive({
  instValue: '',
  instType: 'name',
  instFuzzy: false,
  auditType: route.query.type || tabs[0].type,
});

const details = reactive({
  id: undefined,
  bizId: undefined,
  show: false,
});

const sourceOptions = Object.entries(AUDIT_SOURCE_MAP);

const resourceTypeOptions = AUDIT_RESOURCE_TYPES;

const { query, datas, isLoading, pagination, handlePageChange, handlePageSizeChange, handleSort } = useList({
  filter,
  filterOptions,
});

const isBizType = computed(() => filterOptions.auditType === 'biz');

const getBizName = (id: number) => {
  return businessSelectorComp?.value?.businessList?.find((item) => item.id === id)?.name ?? '--';
};

const columns = computed(() => {
  const values = [
    {
      label: 'ID',
      field: 'id',
      width: 120,
    },
    {
      label: t('云资源 ID'),
      field: 'cloud_res_id',
      showOverflowTooltip: true,
      sort: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: t('名称'),
      field: 'res_name',
      sort: true,
      showOverflowTooltip: true,
    },
    {
      label: t('资源类型'),
      field: 'res_type',
    },
    {
      label: t('动作'),
      field: 'action',
      sort: true,
      render({ cell }: { cell: string }) {
        return h('span', [AUDIT_ACTION_MAP[cell] || '--']);
      },
    },
    {
      label: t('所属业务'),
      field: 'bk_biz_id',
      sort: true,
      visible: isBizType.value,
      render({ cell }: { cell: number }) {
        return h('span', getBizName(cell));
      },
    },
    {
      label: t('云账号'),
      sort: true,
      field: 'account_id',
    },
    {
      label: t('操作者'),
      field: 'operator',
    },
    {
      label: t('来源'),
      field: 'source',
      sort: true,
      render({ cell }: { cell: string }) {
        return h('span', [AUDIT_SOURCE_MAP[cell] || '--']);
      },
    },
    {
      label: t('时间'),
      field: 'created_at',
      sort: true,
      render({ cell }: { cell: string }) {
        return h('span', [timeFormatter(cell)]);
      },
    },
    {
      label: t('操作'),
      render({ data }: any) {
        return h(
          Button,
          {
            theme: 'primary',
            text: true,
            onClick() {
              handleShowDetailSlider(data);
            },
          },
          '详情',
        );
      },
    },
  ];

  return values.filter((item) => item.visible !== false);
});

const handleSearch = () => {
  query();
};
const handleReset = () => {
  filter = Object.assign(filter, defaultFilter());
  query();
};
const handleShowDetailSlider = (row: any) => {
  details.id = row.id;
  details.bizId = isBizType.value ? row.bk_biz_id : null;
  details.show = true;
};

watch(
  () => filter.bk_biz_id,
  (bizId, oldBizId) => {
    console.log(bizId);
    if (oldBizId === null && bizId !== oldBizId) {
      query();
    }
  },
  { immediate: true },
);

watch(isBizType, (isBizType) => {
  if (!isBizType) {
    filter.bk_biz_id = null;
  }
  datas.value = [];
});

const { authVerifyData } = useVerify();
</script>

<template>
  <div class="audit-container">
    <bk-tab v-model:active="filterOptions.auditType" type="card" class="resource-main g-scroller">
      <bk-tab-panel
        v-for="item in tabs"
        :key="item.type"
        :name="item.type"
        :label="item.label"
        render-directive="if"
      ></bk-tab-panel>
    </bk-tab>
    <div v-if="authVerifyData?.permissionAction?.resource_audit_find">
      <div class="audit-filter">
        <div class="filter-item" v-if="isBizType">
          <div class="filter-item-label">业务</div>
          <div class="filter-item-content">
            <business-selector
              v-model="filter.bk_biz_id"
              :authed="isBizType"
              :auto-select="true"
              :is-audit="true"
              :clearable="false"
              ref="businessSelectorComp"
            />
          </div>
        </div>
        <div class="filter-item">
          <div class="filter-item-label">云账号</div>
          <div class="filter-item-content">
            <account-selector
              v-model="filter.account_id"
              :biz-id="filter.bk_biz_id"
              multiple-mode="tag"
              :type="'resource'"
              filterable
              multiple
              allow-create
            />
          </div>
        </div>
        <div class="filter-item">
          <div class="filter-item-label">资源类型</div>
          <div class="filter-item-content">
            <bk-select v-model="filter.res_type" filterable :multiple="false" @change="filter.action = ''">
              <bk-option
                v-for="(item, index) in resourceTypeOptions"
                :key="index"
                :value="item.type"
                :label="item.name"
              />
            </bk-select>
          </div>
        </div>
        <div class="filter-item">
          <div class="filter-item-label">动作</div>
          <div class="filter-item-content">
            <filter-item-action :type="filter.res_type" v-model="filter.action"></filter-item-action>
          </div>
        </div>
        <div class="filter-item">
          <div class="filter-item-label">时间</div>
          <div class="filter-item-content">
            <bk-date-picker class="audit-date-picker" v-model="filter.created_at" :clearable="false" type="daterange" />
          </div>
        </div>
        <div class="filter-item">
          <div class="filter-item-label">操作者</div>
          <div class="filter-item-content">
            <member-select v-model="filter.operator" :allow-create="true" />
          </div>
        </div>
        <div class="filter-item">
          <div class="filter-item-label">实例</div>
          <div class="filter-item-content">
            <bk-input v-model="filterOptions.instValue" placeholder="请输入">
              <template #prefix>
                <bk-select v-model="filterOptions.instType" :clearable="false" class="input-prefix-select">
                  <bk-option value="name" label="名称" />
                  <bk-option value="id" label="ID" />
                </bk-select>
              </template>
              <template #suffix>
                <bk-checkbox v-model="filterOptions.instFuzzy" size="small" class="input-suffix-checkbox">
                  模糊
                </bk-checkbox>
              </template>
            </bk-input>
          </div>
        </div>
        <div class="filter-item">
          <div class="filter-item-label">来源</div>
          <div class="filter-item-content">
            <bk-select v-model="filter.source">
              <bk-option v-for="(item, index) in sourceOptions" :key="index" :value="item[0]" :label="item[1]" />
            </bk-select>
          </div>
        </div>
        <div class="filter-item actions">
          <bk-button theme="primary" class="action-button" @click="handleSearch">查询</bk-button>
          <bk-button class="action-button" @click="handleReset">清空</bk-button>
        </div>
      </div>

      <bk-loading :loading="isLoading">
        <bk-table
          class="audit-list-table"
          row-hover="auto"
          remote-pagination
          :border="['outer']"
          :columns="columns"
          :data="datas"
          :pagination="pagination"
          show-overflow-tooltip
          @page-limit-change="handlePageSizeChange"
          @page-value-change="handlePageChange"
          @column-sort="handleSort"
        />
      </bk-loading>

      <bk-sideslider v-model:isShow="details.show" title="审计详情" width="670" quick-close>
        <template #default>
          <audit-detail :id="details.id" :biz-id="details.bizId" :type="filterOptions.auditType"></audit-detail>
        </template>
      </bk-sideslider>
    </div>
    <ErrorPages v-else url-key-id="resource_audit_find"></ErrorPages>
  </div>
</template>

<style lang="scss" scoped>
.audit-container {
  background: #fff;
}
.audit-filter {
  padding: 0 20px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: 12px 48px;

  .filter-item {
    display: flex;
    align-items: center;
    .filter-item-label {
      width: 64px;
      text-align: left;
    }
    .filter-item-content {
      flex: 1;

      .audit-date-picker {
        width: 100%;
      }

      .input-prefix-select {
        width: 90px;
        :deep(.bk-input) {
          border-top: none;
          border-bottom: none;
          border-left: none;
          height: 30px;
        }
      }
      .input-suffix-checkbox {
        padding: 0 4px;
      }
    }

    &.actions {
      .action-button {
        min-width: 86px;
        & + .action-button {
          margin-left: 8px;
        }
      }
    }
  }
}
.audit-list-table {
  margin-top: 20px;

  :deep(.bk-table-footer) {
    padding: 0 15px;
  }
}
</style>
