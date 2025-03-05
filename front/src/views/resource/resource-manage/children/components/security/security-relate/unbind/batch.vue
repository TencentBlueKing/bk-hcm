<script setup lang="ts">
import { computed, ref, useAttrs, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useVerify } from '@/hooks';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import {
  type SecurityGroupRelResourceByBizItem,
  type SecurityGroupRelatedResourceName,
  useSecurityGroupStore,
} from '@/store/security-group';
import columnFactory from '../data-list/column-factory';
import { RELATED_RES_KEY_MAP } from '@/constants/security-group';
import { ISearchSelectValue } from '@/typings';
import { getLocalFilterFnBySearchSelect } from '@/utils/search';

import { ThemeEnum } from 'bkui-vue/lib/shared';
import search from '../search/index.vue';
import dialogFooter from '@/components/common-dialog/dialog-footer.vue';

const props = withDefaults(
  defineProps<{
    selections: SecurityGroupRelResourceByBizItem[];
    disabled?: boolean;
    tabActive: SecurityGroupRelatedResourceName;
    handleConfirm: (ids: string[]) => Promise<void>;
  }>(),
  { disabled: false },
);

const { t } = useI18n();
const { whereAmI } = useWhereAmI();
const attrs: any = useAttrs();
const securityGroupStore = useSecurityGroupStore();

// 预鉴权
const { handleAuth, authVerifyData } = useVerify();
const authAction = computed(() => {
  return whereAmI.value === Senarios.business ? 'biz_iaas_resource_operate' : 'iaas_resource_operate';
});

const isShow = ref(false);
const handleShow = () => {
  if (!authVerifyData.value?.permissionAction?.[authAction.value]) {
    handleAuth(authAction.value);
    return;
  }
  isShow.value = true;
};
const handleClosed = () => {
  isShow.value = false;
};

const types = [
  { label: t('可解绑'), value: 'target' },
  { label: t('不可解绑'), value: 'unTarget' },
];
const selectedType = ref<'target' | 'unTarget'>('target');

const { getColumns } = columnFactory();
const columns = ref(getColumns(props.tabActive, 'unbind'));
const listMap = ref<Record<'target' | 'unTarget', SecurityGroupRelResourceByBizItem[]>>({
  target: [],
  unTarget: [],
});
const filterFn = ref<(item: any) => boolean>(() => true);
const renderList = computed(() => {
  return listMap.value[selectedType.value].filter(filterFn.value);
});

watch(isShow, async (val) => {
  if (!val) return;
  const res = await securityGroupStore.pullSecurityGroup(RELATED_RES_KEY_MAP[props.tabActive], props.selections);

  const target = res.filter((item) => item.security_groups.length > 1);
  const unTarget = res.filter((item) => item.security_groups.length <= 1);

  listMap.value = { target, unTarget };
  selectedType.value = unTarget.length > 0 ? 'unTarget' : 'target';
});

const handleSearch = (searchValue: ISearchSelectValue) => {
  filterFn.value = getLocalFilterFnBySearchSelect(searchValue);
};

const handleConfirm = async () => {
  await props.handleConfirm(listMap.value.target.map((item) => item.id));
  handleClosed();
};
</script>

<template>
  <bk-button
    :class="{ 'hcm-no-permision-btn': !disabled && !authVerifyData?.permissionAction?.[authAction] }"
    :disabled="disabled"
    @click="handleShow"
    v-bind="attrs"
  >
    {{ t('批量解绑') }}
  </bk-button>
  <bk-dialog v-model:isShow="isShow" :title="t('批量解绑')" :width="1280" @closed="handleClosed">
    <div class="tips">
      {{ t('已选择') }}
      <span class="number primary">{{ selections.length }}</span>
      {{ t('个资源，其中可解绑') }}
      <span class="number success">{{ listMap.target.length }}</span>
      {{ t('个，不可解绑') }}
      <span class="number danger">{{ listMap.unTarget.length }}</span>
      {{ t('个。请确认后再操作。') }}
    </div>

    <div class="tools">
      <bk-radio-group v-model="selectedType">
        <bk-radio-button v-for="{ label, value } in types" :key="value" :label="value">{{ label }}</bk-radio-button>
      </bk-radio-group>
      <span class="tips">
        <i class="hcm-icon bkhcm-icon-info-line"></i>
        <span>{{ t('仅绑定1个安全组的资源不允许进行批量解绑') }}</span>
      </span>
      <!-- TODO：本地搜索 -->
      <search class="search" :resource-name="tabActive" operation="unbind" @search="handleSearch" />
    </div>

    <bk-table
      v-bkloading="{ loading: securityGroupStore.isBatchQuerySecurityGroupByResIdsLoading }"
      ref="tableRef"
      row-hover="auto"
      :data="renderList"
      :max-height="'calc(100vh - 401px)'"
      show-overflow-tooltip
      row-key="id"
    >
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :render="column.render"
      >
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </bk-table-column>
    </bk-table>

    <template #footer>
      <dialog-footer
        :disabled="!listMap.target.length"
        :loading="securityGroupStore.isBatchDisassociateCvmsLoading"
        :confirm-text="t('解绑')"
        :confirm-button-theme="ThemeEnum.DANGER"
        @confirm="handleConfirm"
        @closed="handleClosed"
      />
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.tips {
  color: #313238;

  .number {
    font-weight: 700;

    &.primary {
      color: #3a84ff;
    }
    &.success {
      color: #299e56;
    }
    &.danger {
      color: #ea3636;
    }
  }
}

.tools {
  margin: 16px 0;
  display: flex;
  align-items: center;
  .tips {
    margin-left: 16px;
    font-size: 12px;
    color: #4d4f56;

    .hcm-icon {
      margin-right: 4px;
      font-size: 14px;
    }
  }
  .search {
    margin-left: auto;
    width: 400px;
  }
}
</style>
