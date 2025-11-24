<script setup lang="ts">
import { computed, Ref, ref, useTemplateRef } from 'vue';
import { Message } from 'bkui-vue';
import { IOrderStatistics } from './typings';
import { actionColumns } from './column';
import http from '@/http';
import dayjs from 'dayjs';
import { SelectColumn, TagInputColumn, DateTimePickerColumn, InputColumn } from '@blueking/ediatable';
import { useBusinessGlobalStore } from '@/store/business-global';

const emit = defineEmits<{
  refresh: [];
}>();

type InnerConfig = IOrderStatistics & { timeRange?: string[] };
interface IFormData {
  month: string;
  configs: InnerConfig[];
}

const showDialog = ref(false);
const actionType = ref('add'); // add 或 edit
const createTemp = (): InnerConfig => {
  return {
    bk_biz_id: '',
    sub_order_ids: [],
    start_at: '',
    end_at: '',
    memo: '',
    timeRange: [],
  };
};
const formData = ref<IFormData>({
  month: '',
  configs: [createTemp()],
});
const title = computed(() => {
  return actionType.value === 'add' ? '新增剔除' : '编辑剔除';
});
const businessGlobalStore = useBusinessGlobalStore();

const formRef = useTemplateRef('formRef');

const setFormData = (data?: IFormData) => {
  if (data) {
    formData.value = {
      month: data.month,
      configs: data?.configs.length
        ? data.configs.map((item) => {
            return {
              ...item,
              timeRange: [item.start_at, item.end_at],
            };
          })
        : [createTemp()],
    };
    return;
  }
  formData.value = {
    month: '',
    configs: [createTemp()],
  };
};
const rowRefs: Ref<Ref[][]> = ref([[ref(), ref(), ref(), ref()]]);
const createRowRefs = (rows: number) => {
  rowRefs.value = Array.from({ length: rows }, () => [ref(), ref(), ref(), ref()]);
};
const add = () => {
  setFormData();
  createRowRefs(1);
  actionType.value = 'add';
  showDialog.value = true;
};
const edit = (data: IFormData) => {
  setFormData(data);
  createRowRefs(data.configs.length);
  actionType.value = 'edit';
  showDialog.value = true;
};

const closeDialog = () => {
  setFormData();
  showDialog.value = false;
};

const selectList = computed(() => {
  return businessGlobalStore.businessFullList;
});
const loading = ref(false);

const validateForm = () => {
  const validArr = rowRefs.value.flat().map((item) => item.value?.getValue());
  return Promise.all([formRef.value?.validate(), ...validArr]);
};
const handleApply = async () => {
  const valid = await validateForm();

  if (!valid) {
    return;
  }
  const params = {
    stat_month: dayjs(formData.value.month).format('YYYY-MM'),
    configs: formData.value.configs.map((item) => {
      return {
        bk_biz_id: item.bk_biz_id,
        sub_order_ids: item.sub_order_ids,
        start_at: item.timeRange[0] ? dayjs(item.timeRange[0]).format('YYYY-MM-DD HH:mm:ss') : '',
        end_at: item.timeRange[1] ? dayjs(item.timeRange[1]).format('YYYY-MM-DD HH:mm:ss') : '',
        memo: item.memo,
      };
    }),
  };
  loading.value = true;
  const promise =
    actionType.value === 'add'
      ? http.post('/api/v1/woa/task/config/create/apply/order/statistics', params)
      : http.put('/api/v1/woa/task/config/update/apply/order/statistics', params);
  promise
    .then(() => {
      Message({ theme: 'success', message: '提交成功' });
      emit('refresh');
      closeDialog();
    })
    .finally(() => {
      loading.value = false;
    });
};

const handelAdd = () => {
  rowRefs.value.push([ref(), ref(), ref(), ref()]);
  formData.value.configs.push(createTemp());
};
const handleRemove = (index: number) => {
  formData.value.configs.splice(index, 1);
};

defineExpose({
  add,
  edit,
});
</script>

<template>
  <bk-dialog
    class="order-config-dialog"
    :title="title"
    v-model:is-show="showDialog"
    width="1100"
    render-directive="if"
    :quick-close="false"
    @closed="closeDialog"
  >
    <bk-form ref="formRef" :model="formData" form-type="vertical">
      <bk-form-item label="月份" property="month" required>
        <hcm-search-datetime
          type="month"
          v-model="formData.month"
          placeholder="请选择月份"
          format="yyyy-MM"
          :disabled="actionType === 'edit'"
        />
      </bk-form-item>
    </bk-form>

    <!-- 配置表格 -->
    <data-list-table
      class="config-list"
      :list="formData.configs"
      :columns="actionColumns"
      :settings="false"
      :border="['row', 'col', 'outer']"
      :show-overflow-tooltip="false"
      max-height="900"
    >
      <!-- 业务 -->
      <template #bk_biz_id="{ row, index }">
        <SelectColumn
          :ref="rowRefs?.[index]?.[0]"
          v-model="row.bk_biz_id"
          id-key="id"
          display-key="name"
          :list="selectList"
          :rules="[
            {
              message: '请选择业务',
              validator: (v) => !!v,
            },
          ]"
          placeholder="请选择"
        />
      </template>

      <!-- 单号 -->
      <template #sub_order_ids="{ row, index }">
        <TagInputColumn
          :ref="rowRefs?.[index]?.[1]"
          v-model="row.sub_order_ids"
          has-delete-icon
          :show-clear-only-hover="true"
          :collapse-tags="true"
          :rules="[
            {
              message: '最多输入100个单号',
              validator: (v) => v.length <= 100,
            },
            {
              message: '单号和时间段必须填写一个',
              validator: (v) => !!row?.sub_order_ids?.length || !!(row?.timeRange?.[0] && row?.timeRange?.[1]),
            },
          ]"
        />
      </template>

      <template #start_at="{ row, index }">
        <DateTimePickerColumn
          :ref="rowRefs?.[index]?.[2]"
          type="daterange"
          v-model="row.timeRange"
          placeholder="请选择"
          format="yyyy-MM-dd HH:mm:ss"
          :rules="[
            {
              message: '单号和时间段必须填写一个',
              validator: (v) => !!row?.sub_order_ids?.length || !!(row?.timeRange?.[0] && row?.timeRange?.[1]),
            },
          ]"
        />
      </template>

      <template #memo="{ row, index }">
        <InputColumn
          :ref="rowRefs?.[index]?.[3]"
          v-model="row.memo"
          :rules="[
            {
              message: '请输入备注',
              validator: (v) => !!v,
            },
          ]"
        />
      </template>

      <template #action="{ index }">
        <div class="action-btn">
          <i class="hcm-icon bkhcm-icon-plus-circle-shape" @click="handelAdd"></i>
          <i class="hcm-icon bkhcm-icon-minus-circle-shape ml18" @click="handleRemove(index)"></i>
        </div>
      </template>
    </data-list-table>

    <template #footer>
      <div>
        <bk-button class="mr10" theme="primary" :loading="loading" @click="handleApply">确定</bk-button>
        <bk-button @click="closeDialog">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.config-list {
  :deep(.bk-table-body table td .cell) {
    padding: 0;
  }
}

.action-btn {
  display: flex;
  align-items: center;
  height: 42px;
  padding: 0 16px;

  i {
    color: #c4c6cc;
    cursor: pointer;
    font-size: 14px;
    transition: all 0.15s;
  }

  & i:hover {
    color: #979ba5;
  }

  & i.disabled {
    color: #dcdee5;
    cursor: not-allowed;
  }
}
</style>
