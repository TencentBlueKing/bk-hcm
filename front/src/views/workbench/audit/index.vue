<script lang="ts" setup>
import BusinessSelector from '@/components/business-selector/index.vue';
import AccountSelector from '@/components/account-selector/index.vue';
import MemberSelect from '@/components/MemberSelect';
import FilterItemAction from './children/filter-item-action.vue';
import AuditDetail from './detail.vue';

import { computed, reactive, ref, watch } from 'vue';
import dayjs from 'dayjs';
import { useRoute } from 'vue-router';
import {
  useI18n,
} from 'vue-i18n';
import useList from './use-list';
import { AUDIT_RESOURCE_TYPES } from '@/common/constant';
import { timeFormatter } from '@/common/util';
import { AUDIT_SOURCE_MAP, AUDIT_ACTION_MAP } from './constants';

const {
  t,
} = useI18n();
const route = useRoute();

const businessSelectorComp = ref(null);

const tabs = [
  {
    type: 'biz',
    label: '业务',
  }, {
    type: 'resource',
    label: '资源',
  },
];

const todayStart = dayjs(new Date()).format('YYYY-MM-DD 00:00:00');
const todayEnd = dayjs(new Date()).format('YYYY-MM-DD 23:59:59');
const defaultFilter = () => ({
  bk_biz_id: null as number,
  account_id: '',
  res_type: 'account',
  action: '',
  created_at: [todayStart, todayEnd],
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

const {
  query,
  datas,
  isLoading,
  pagination,
  handlePageChange,
  handlePageSizeChange,
} = useList({ filter, filterOptions });

const isBizType = computed(() => filterOptions.auditType === 'biz');

const getBizName = (id: number) => {
  return businessSelectorComp?.value?.businessList?.find(item => item.id === id)?.name ?? '--';
};

const handleSearch = () => {
  query();
};
const handleReset = () => {
  filter = Object.assign(filter, defaultFilter());
  query();
};
const handleShowDetailSlider = (row: any) => {
  details.id = row.id;
  if (isBizType.value) {
    details.bizId = row.bk_biz_id;
  }
  details.show = true;
};

watch(() => filter.bk_biz_id, (bizId, oldBizId) => {
  if (oldBizId === null && bizId !== oldBizId) {
    query();
  }
});

watch(isBizType, (isBizType) => {
  if (!isBizType) {
    filter.bk_biz_id = null;
  }
});
</script>

<template>
  <bk-tab
    v-model:active="filterOptions.auditType"
    type="unborder-card"
    class="resource-main g-scroller"
  >
    <bk-tab-panel
      v-for="item in tabs"
      :key="item.type"
      :name="item.type"
      :label="item.label"
      render-directive="if"
    >
    </bk-tab-panel>
  </bk-tab>

  <div class="audit-filter">
    <div class="filter-item" v-if="isBizType">
      <div class="filter-item-label">业务</div>
      <div class="filter-item-content">
        <business-selector
          v-model="filter.bk_biz_id"
          :authed="isBizType"
          :auto-select="true"
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
          filterable
          multiple
          allow-create />
      </div>
    </div>
    <div class="filter-item">
      <div class="filter-item-label">资源类型</div>
      <div class="filter-item-content">
        <bk-select
          v-model="filter.res_type"
          filterable
          :multiple="false"
        >
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
        <bk-date-picker
          class="audit-date-picker"
          v-model="filter.created_at"
          clearable
          type="daterange"
        />
      </div>
    </div>
    <div class="filter-item">
      <div class="filter-item-label">操作者</div>
      <div class="filter-item-content">
        <member-select></member-select>
      </div>
    </div>
    <div class="filter-item">
      <div class="filter-item-label">实例</div>
      <div class="filter-item-content">
        <bk-input v-model="filterOptions.instValue" placeholder="请输入">
          <template #prefix>
            <bk-select v-model="filterOptions.instType" class="input-prefix-select">
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
          <bk-option
            v-for="(item, index) in sourceOptions"
            :key="index"
            :value="item[0]"
            :label="item[1]"
          />
        </bk-select>
      </div>
    </div>
    <div class="filter-item actions">
      <bk-button theme="primary" class="action-button" @click="handleSearch">查询</bk-button>
      <bk-button class="action-button" @click="handleReset">清空</bk-button>
    </div>
  </div>

  <bk-loading
    :loading="isLoading"
  >
    <bk-table
      class="audit-list-table"
      row-hover="auto"
      remote-pagination
      :border="['outer']"
      :data="datas"
      :pagination="pagination"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
    >
      <bk-table-column label="ID" prop="id" :width="120" />
      <bk-table-column :label="t('云资源 ID')" prop="cloud_res_id" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row?.cloud_res_id || '--' }}
        </template>
      </bk-table-column>
      <bk-table-column :label="t('名称')" prop="res_name" show-overflow-tooltip />
      <bk-table-column :label="t('资源类型')" prop="res_type" />
      <bk-table-column :label="t('动作')" prop="action">
        <template #default="{ row }">
          {{ AUDIT_ACTION_MAP[row?.action] }}
        </template>
      </bk-table-column>
      <bk-table-column :label="t('所属业务')" prop="bk_biz_id">
        <template #default="{ row }">
          {{ getBizName(row?.bk_biz_id) }}
        </template>
      </bk-table-column>
      <bk-table-column :label="t('云账号')" prop="account_id" />
      <bk-table-column :label="t('操作者')" prop="operator" />
      <bk-table-column :label="t('来源')" prop="source">
        <template #default="{ row }">
          {{ AUDIT_SOURCE_MAP[row?.source] }}
        </template>
      </bk-table-column>
      <bk-table-column :label="t('时间')" :width="160" prop="created_at">
        <template #default="{ row }">
          {{ timeFormatter(row?.created_at) }}
        </template>
      </bk-table-column>
      <bk-table-column :label="t('操作')">
        <template #default="{ row }">
          <bk-button theme="primary" @click="handleShowDetailSlider(row)">详情</bk-button>
        </template>
      </bk-table-column>
    </bk-table>
  </bk-loading>

  <bk-sideslider
    v-model:isShow="details.show"
    title="审计详情"
    width="670"
    quick-close
  >
    <template #default>
      <audit-detail :id="details.id" :biz-id="details.bizId"></audit-detail>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.audit-container {
  background: #fff;
}
.audit-filter {
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
