<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  ISecurityGroupDetail,
  type SecurityGroupRelResourceByBizItem,
  useSecurityGroupStore,
} from '@/store/security-group';
import { useBusinessGlobalStore } from '@/store/business-global';
import columnFactory from '../data-list/column-factory';
import { RELATED_RES_KEY_MAP, SecurityGroupRelatedResourceName } from '@/constants/security-group';
import { ISearchSelectValue } from '@/typings';

import { Message } from 'bkui-vue';
import { ThemeEnum } from 'bkui-vue/lib/shared';
import search from '../search/index.vue';
import dialogFooter from '@/components/common-dialog/dialog-footer.vue';

const props = defineProps<{
  selections: SecurityGroupRelResourceByBizItem[];
  tabActive: SecurityGroupRelatedResourceName;
  detail: ISecurityGroupDetail;
}>();
const emit = defineEmits(['success']);
const model = defineModel<boolean>();

const { t } = useI18n();
const securityGroupStore = useSecurityGroupStore();
const { getBusinessIds } = useBusinessGlobalStore();

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

watch(
  model,
  async (val) => {
    if (!val) return;
    const res = await securityGroupStore.pullSecurityGroup(RELATED_RES_KEY_MAP[props.tabActive], props.selections);

    const target = res.filter((item) => item.security_groups.length > 1);
    const unTarget = res.filter((item) => item.security_groups.length <= 1);

    listMap.value = { target, unTarget };
    selectedType.value = unTarget.length > 0 ? 'unTarget' : 'target';
  },
  { immediate: true },
);

const formatterOptions = [{ field: 'bk_biz_id', formatter: (name: string) => getBusinessIds(name) }];
const handleSearch = (_: ISearchSelectValue, fn: (item: any) => boolean) => {
  filterFn.value = fn;
};

const handleConfirm = async () => {
  await securityGroupStore.batchDisassociateCvms({
    security_group_id: props.detail.id,
    cvm_ids: listMap.value.target.map((item) => item.id),
  });
  Message({ theme: 'success', message: t('解绑成功') });
  handleClosed();
  emit('success');
};

const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog v-model:is-show="model" :title="t('批量解绑')" :width="1280" @closed="handleClosed">
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
      <search
        class="search"
        :resource-name="tabActive"
        operation="unbind"
        local-search
        :options="formatterOptions"
        @search="handleSearch"
      />
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
