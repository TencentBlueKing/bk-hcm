<script setup lang="ts">
import { ref, computed, h, reactive } from 'vue';
import { Message } from 'bkui-vue';
import { ModelPropertyColumn } from '@/model/typings';
import { useSecurityGroupStore, type ISecurityGroupItem } from '@/store/security-group';
import HcmFormUser from '@/components/form/user.vue';
import HcmFormBusiness from '@/components/form/business.vue';
import { useAccountBusiness } from '@/views/resource/resource-manage/hooks/use-account-business';
import UsageBizValue from '@/views/resource/resource-manage/children/components/security/usage-biz-value.vue';

const props = defineProps<{ selections: ISecurityGroupItem[] }>();

const emit = defineEmits<{
  success: [];
  closed: [];
}>();

const securityGroupStore = useSecurityGroupStore();

const model = defineModel<boolean>();

const updateValues = reactive<
  Record<
    ISecurityGroupItem['cloud_id'],
    Pick<ISecurityGroupItem, 'manager' | 'bak_manager' | 'mgmt_biz_id'> & { [key: string]: any }
  >
>({});
const updateDefaultValue = () => ({
  manager: '',
  bak_manager: '',
  mgmt_biz_id: undefined as number,
});

const keyword = ref('');
const list = ref(props.selections.slice());
const displayList = computed(() =>
  list.value.filter((item) => item.name.includes(keyword.value) || item.cloud_id.includes(keyword.value)),
);

const accountId = computed(() => list.value?.[0].account_id);

const { accountBizList, isAccountDetailLoading } = useAccountBusiness(accountId.value);

const isUpdateValueValid = computed(() => {
  return list.value.every((item) => {
    const checkFields = ['manager', 'bak_manager', 'mgmt_biz_id'];
    if (!checkFields.every((field) => updateValues[item.id]?.[field])) {
      return false;
    }
    return true;
  });
});

const confirmButtonDisabled = computed(() => list.value.length === 0 || !isUpdateValueValid.value);

const columns: ModelPropertyColumn[] = [
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
    width: 120,
  },
  {
    id: 'rel_res_count',
    name: '关联实例数',
    type: 'number',
    align: 'right',
  },
  {
    id: 'usage_biz_ids',
    name: '使用业务',
    type: 'business',
    showOverflowTooltip: false,
    render: ({ cell }: any) => h(UsageBizValue, { value: cell }),
  },
  {
    id: 'manager',
    name: '主负责人',
    type: 'user',
    width: 200,
    render: ({ row }: { row?: ISecurityGroupItem }) =>
      h(HcmFormUser, {
        multiple: false,
        modelValue: updateValues[row.id]?.manager,
        'onUpdate:modelValue': (val: string | string[]) => {
          if (!updateValues[row.id]) {
            updateValues[row.id] = updateDefaultValue();
          }
          updateValues[row.id].manager = val as string;
        },
        style: { lineHeight: 'normal' },
      }),
  },
  {
    id: 'bak_manager',
    name: '备份主负责人',
    type: 'user',
    width: 200,
    render: ({ row }: { row?: ISecurityGroupItem }) =>
      h(HcmFormUser, {
        multiple: false,
        modelValue: updateValues[row.id]?.bak_manager,
        'onUpdate:modelValue': (val: string | string[]) => {
          if (!updateValues[row.id]) {
            updateValues[row.id] = updateDefaultValue();
          }
          updateValues[row.id].bak_manager = val as string;
        },
        style: { lineHeight: 'normal' },
      }),
  },
  {
    id: 'mgmt_biz_id',
    name: '管理业务',
    type: 'business',
    width: 200,
    render: ({ row }: { row?: ISecurityGroupItem }) =>
      h(HcmFormBusiness, {
        multiple: false,
        ...{ data: accountBizList.value },
        modelValue: updateValues[row.id]?.mgmt_biz_id,
        'onUpdate:modelValue': (val: number | number[]) => {
          if (!updateValues[row.id]) {
            updateValues[row.id] = updateDefaultValue();
          }
          updateValues[row.id].mgmt_biz_id = val as number;
        },
        style: { lineHeight: 'normal' },
      }),
  },
];

const closeDialog = () => {
  model.value = false;
  emit('closed');
};

const handleDialogConfirm = async () => {
  if (confirmButtonDisabled.value) {
    return;
  }

  const data = Object.entries(updateValues).map(([cloudId, value]) => ({ id: cloudId, ...value }));
  await securityGroupStore.batchUpdateMgmtAttr(data);

  Message({ theme: 'success', message: '添加成功' });
  closeDialog();
  emit('success');
};

const handleSearch = (v: string) => {
  keyword.value = v.trim();
};

const handleRemove = (id: ISecurityGroupItem['id']) => {
  const index = list.value.findIndex((item) => item.id === id);
  if (index > -1) {
    list.value.splice(index, 1);
  }
};
</script>

<template>
  <bk-dialog :title="'批量添加资产归属'" :width="1280" :quick-close="false" :is-show="model" @closed="closeDialog">
    <div class="toolbar">
      <bk-input type="search" placeholder="请输入 安全组ID/安全组名称 搜索" @enter="handleSearch" />
    </div>
    <bk-table
      :data="displayList"
      :max-height="500"
      :min-height="190"
      row-hover="auto"
      show-overflow-tooltip
      v-bkloading="{ loading: isAccountDetailLoading }"
    >
      <bk-table-column
        v-for="(column, index) in columns"
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
      <bk-table-column :label="'操作'">
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
          :loading="securityGroupStore.isBatchUpdateMgmtAttrLoading"
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
.toolbar {
  margin-bottom: 16px;
}

.dialog-custom-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
