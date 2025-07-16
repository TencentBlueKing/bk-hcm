<script setup lang="ts">
import { computed, ref, watchEffect, h } from 'vue';
import { Message } from 'bkui-vue';
import { ModelPropertyColumn } from '@/model/typings';
import { Info } from 'bkui-vue/lib/icon';
import { useSecurityGroupStore, type ISecurityGroupItem } from '@/store/security-group';
import UsageBizValue from '@/views/resource/resource-manage/children/components/security/usage-biz-value.vue';

const model = defineModel<boolean>();

const props = defineProps<{ selections: ISecurityGroupItem[] }>();

const emit = defineEmits<{
  success: [];
  closed: [];
}>();

const securityGroupStore = useSecurityGroupStore();

const DataView = {
  Assignable: 'assignable',
  NonAssignable: 'nonAssignable',
};

const previewList = ref([]);
const activeView = ref<string>(DataView.Assignable);
const assignableList = ref([]);
const nonAssignableList = ref([]);
const displayList = ref([]);
const displayColumns = ref([]);

const baseColumns: ModelPropertyColumn[] = [
  {
    id: 'cloud_id',
    name: '安全组 ID',
    type: 'string',
  },
  {
    id: 'name',
    name: '安全组名称',
    type: 'string',
  },
  {
    id: 'res_res_types',
    name: '关联的资源类型',
    type: 'array',
  },
  {
    id: 'rel_res_count',
    name: '关联实例数',
    type: 'number',
    align: 'right',
    sort: true,
  },
  {
    id: 'usage_biz_ids',
    name: '使用业务',
    type: 'business',
    showOverflowTooltip: false,
    render: ({ cell }: any) => h(UsageBizValue, { value: cell }),
  },
];

const assignableColumns: ModelPropertyColumn[] = [
  ...baseColumns,
  {
    id: 'mgmt_biz_id',
    name: '管理业务',
    type: 'business',
  },
  {
    id: 'assign_biz_id',
    name: '分配的目标业务',
    type: 'business',
  },
];
const nonAssignableColumns: ModelPropertyColumn[] = [
  ...baseColumns,
  {
    id: 'reason',
    name: '不可分配原因',
    type: 'string',
  },
];

const closeDialog = () => {
  model.value = false;
  emit('closed');
};

const tipsContent = h('dl', { class: 'tips-content' }, [
  h('div', { class: 'tips-content-group' }, [
    h('dt', { class: 'tips-content-title' }, '可分配'),
    h('dd', '安全组有标签，且标签字段完整(一级业务，二级业务，主备负责人) '),
  ]),
  h('div', { class: 'tips-content-group' }, [
    h('dt', { class: 'tips-content-title' }, '不可分配'),
    h('dd', '1.管理类型为「未确认」「平台管理」'),
    h('dd', '2.安全组无标签或标签不完整(一级业务，二级业务，主备负责人)'),
    h('dd', '3.安全组中的规则引用的安全组不是本业务'),
  ]),
]);

watchEffect(async () => {
  // 获取分配预览数据
  const ids = props.selections.map((item) => item.id);
  previewList.value = await securityGroupStore.getAssignPreview(ids);

  props.selections.forEach((item) => {
    const preview = previewList.value.find((previewItem) => item.id === previewItem.id);
    if (preview.assignable) {
      assignableList.value.push({ ...item, assign_biz_id: item.mgmt_biz_id, visible: true });
    } else {
      nonAssignableList.value.push({ ...item, reason: preview.reason });
    }
  });

  setCurrentView(assignableList.value.length > 0 ? DataView.Assignable : DataView.NonAssignable);
});

const confirmButtonDisabled = computed(
  () => activeView.value !== DataView.Assignable || displayList.value.length === 0,
);

const getDisplayList = (view: string) => {
  return view === DataView.Assignable
    ? assignableList.value.slice().filter((item) => item.visible)
    : nonAssignableList.value.slice();
};

const setCurrentView = (view: string) => {
  activeView.value = view;
  displayList.value = getDisplayList(view);
  displayColumns.value = view === DataView.Assignable ? assignableColumns : nonAssignableColumns;
};

const handleChangeView = (view: string) => {
  setCurrentView(view);
};

const handleDialogConfirm = async () => {
  if (confirmButtonDisabled.value) {
    return;
  }

  // “确定”操作限制为只在可分配视图下进行，当前展示的数据即为最终分配数据，否则需要注意分配数据的正确性
  const ids = assignableList.value.filter((item) => item.visible).map((item) => item.id);
  await securityGroupStore.batchAssignToBiz(ids);

  emit('success');
  Message({ theme: 'success', message: '分配成功' });
  closeDialog();
};

const handleSearch = (keyword: string) => {
  if (!keyword) {
    displayList.value = getDisplayList(activeView.value);
    return;
  }
  displayList.value = getDisplayList(activeView.value).filter(
    (item) => item.name.includes(keyword) || item.cloud_id.includes(keyword),
  );
};

// 只有可分配视图可进行移除操作
const handleRemove = (id: ISecurityGroupItem['id']) => {
  const index = assignableList.value.findIndex((item) => item.id === id);
  if (index > -1) {
    assignableList.value[index].visible = false;
    displayList.value = getDisplayList(activeView.value);
  }
};
</script>

<template>
  <bk-dialog :title="'批量分配安全组'" :width="1280" :quick-close="false" :is-show="model" @closed="closeDialog">
    <div class="stat-tips">
      已选择
      <em class="count total-count">{{ selections.length }}</em>
      个安全组， 其中可分配
      <em class="count valid-count">{{ assignableList.length }}</em>
      个， 不可分配
      <em class="count invalid-count">{{ nonAssignableList.length }}</em>
      个。不可分配将忽略操作
    </div>
    <div class="toolbar">
      <bk-button-group>
        <bk-button :selected="activeView === DataView.Assignable" @click="handleChangeView(DataView.Assignable)">
          可分配
        </bk-button>
        <bk-button :selected="activeView === DataView.NonAssignable" @click="handleChangeView(DataView.NonAssignable)">
          不可分配
        </bk-button>
      </bk-button-group>
      <info
        class="preview-tips-icon"
        v-bk-tooltips="{
          theme: 'light',
          placement: 'right',
          extCls: 'security-group-assign-tooltips-popover',
          content: tipsContent,
        }"
      />
      <div class="searchbar">
        <bk-input type="search" placeholder="请输入 安全组ID/安全组名称 搜索" @enter="handleSearch" />
      </div>
    </div>
    <bk-table
      :data="displayList"
      :max-height="500"
      :min-height="190"
      row-hover="auto"
      show-overflow-tooltip
      v-bkloading="{ loading: securityGroupStore.isAssignPreviewLoading }"
    >
      <bk-table-column
        v-for="(column, index) in displayColumns"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :sort="column.sort"
        :align="column.align"
        :width="column.width"
        :show-overflow-tooltip="column.showOverflowTooltip"
        :render="column.render"
      >
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </bk-table-column>
      <bk-table-column :label="'操作'" v-if="activeView === DataView.Assignable">
        <template #default="{ row }">
          <bk-button text @click="handleRemove(row.id)">
            <i class="hcm-icon bkhcm-icon-minus-circle-shape"></i>
          </bk-button>
        </template>
      </bk-table-column>
    </bk-table>
    <template #footer>
      <div class="dialog-custom-footer">
        <bk-button
          theme="primary"
          :disabled="confirmButtonDisabled"
          :loading="securityGroupStore.isBatchAssignToBizLoading"
          @click="handleDialogConfirm"
        >
          确定
        </bk-button>
        <bk-button @click="closeDialog">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.stat-tips {
  margin-bottom: 16px;

  .count {
    font-style: normal;
  }

  .total-count {
    color: #3a84ff;
    font-style: normal;
  }

  .valid-count {
    color: #299e56;
    font-style: normal;
  }

  .invalid-count {
    color: #ea3636;
    font-style: normal;
  }
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;

  .preview-tips-icon {
    color: #c4c6cc;

    &:hover {
      color: #3a84ff;
      cursor: pointer;
    }
  }

  .searchbar {
    width: 400px;
    margin-left: auto;
  }
}

.dialog-custom-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
<style lang="scss">
.security-group-assign-tooltips-popover {
  .tips-content {
    .tips-content-title {
      display: flex;
      align-items: center;
      gap: 4px;
      font-weight: 700;
      color: #4d4f56;

      &::before {
        content: '';
        width: 2px;
        height: 12px;
        background: #3a84ff;
      }
    }

    .tips-content-group {
      & + .tips-content-group {
        margin-top: 12px;
      }
    }
  }
}
</style>
