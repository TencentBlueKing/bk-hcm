<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import useBusiness from '../../hooks/use-business';
import useMountedDrive from '../../hooks/use-mounted-drive';
import useUninstallDrive from '../../hooks/use-uninstall-drive';
import useDelete from '../../hooks/use-delete';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';
import useColumns from '../../hooks/use-columns';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

const {
  t,
} = useI18n();

const columns = useColumns('drive');

const {
  isShowDistribution,
  handleDistribution,
  ResourceBusiness,
} = useBusiness();

const {
  isShowMountedDrive,
  handleMountedDrive,
  MountedDrive,
} = useMountedDrive();

const {
  isShowUninstallDrive,
  handleUninstallDrive,
  UninstallDrive,
} = useUninstallDrive();

const {
  selections,
  handleSelectionChange,
} = useSelection();

const {
  handleShowDelete,
  DeleteDialog,
} = useDelete(
  columns,
  selections,
  'disks',
  t('删除硬盘'),
  true,
);

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'disks');
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <section>
      <bk-button
        class="w100"
        theme="primary"
        :disabled="selections.length <= 0"
        @click="handleDistribution"
      >
        {{ t('分配') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleMountedDrive"
      >
        {{ t('挂载') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleUninstallDrive"
      >
        {{ t('卸载') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        :disabled="selections.length <= 0"
        @click="handleShowDelete"
      >
        {{ t('删除') }}
      </bk-button>
    </section>

    <bk-table
      class="mt20"
      row-hover="auto"
      :pagination="pagination"
      :columns="columns"
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      @selection-change="handleSelectionChange"
    />
  </bk-loading>

  <resource-business
    v-model:is-show="isShowDistribution"
    type="disks"
    :title="t('云硬盘分配')"
    :list="selections"
  />

  <mounted-drive
    v-model:is-show="isShowMountedDrive"
  />

  <uninstall-drive
    v-model:is-show="isShowUninstallDrive"
  />

  <delete-dialog>
    {{ t('请注意删除VPC后无法恢复，请谨慎操作') }}
  </delete-dialog>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
.mt20 {
  margin-top: 20px;
}
</style>
