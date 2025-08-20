<script setup lang="ts">
import { ref, watch, toRaw, computed } from 'vue';
import get from 'lodash/get';
import {
  type ITaskDetailItem,
  type ITaskItem,
  type ITaskDetailParam,
  type ITaskRerunParams,
  useTaskStore,
} from '@/store';
import { getModel } from '@/model/manager';
import { RerunView } from '@/model/task/rerun.view';
import { ResourceTypeEnum } from '@/common/resource-constant';
import routerAction from '@/router/utils/action';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { Ediatable, HeadColumn, FixedColumn } from '@blueking/ediatable';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import columnFactory from '@/views/task/details/children/action-list/column-factory';

const model = defineModel<boolean>({ default: false });
const props = defineProps<{ resource: ResourceTypeEnum; info: Partial<ITaskItem>; selected: ITaskDetailItem[] }>();
const properties = getModel(RerunView).getProperties();

const fields = ['vendors', 'region_id', 'operations'].map((id) => properties.find((item) => item.id === id));

const { getRerunColumns } = columnFactory(props.info.vendor as any);
const columns = getRerunColumns(props.resource, props.info.operations);

const { getBizsId } = useWhereAmI();
const taskStore = useTaskStore();

const list = ref<ITaskDetailItem[]>([]);
const editState = ref<ITaskDetailItem>();

const editCompRefs = ref();

// 重试参数中所有的region_id
const regionIds = computed(() => [...new Set(list.value.map((item) => item.param.region_id))].filter((item) => item));

const baseInfo = computed<Partial<ITaskItem>>(() => ({ ...props.info, region_id: regionIds.value }));

// 最近一次重试参数
const latestParams = ref<ITaskRerunParams>();
// 参数是否校验通过
const isValidatePassed = ref(null);

// 是否触发过校验参数操作
const hasChecked = ref(false);

watch(
  [() => props.selected, () => model.value],
  ([val, isShow]) => {
    if (isShow) {
      list.value = structuredClone(toRaw(val));

      // 每次打开重置相关数据
      editState.value = null;
      isValidatePassed.value = null;
      hasChecked.value = false;

      // 默认自动触发一次校验
      handleValidate();
    }
  },
  { deep: true, immediate: true },
);

const handelEdit = (row: ITaskDetailItem) => {
  editState.value = { ...row };
  columns.forEach((col) => {
    const val = get(row, col.field.id);
    editState.value[col.field.id] = col.field.type === 'bool' ? String(val) : val;
  });

  isValidatePassed.value = false;
};
const handleRemove = (row: ITaskDetailItem) => {
  const listRowIndex = list.value.findIndex((item) => item.id === row.id);
  if (listRowIndex !== -1) {
    list.value.splice(listRowIndex, 1);
  }

  isValidatePassed.value = false;
};
const handelDone = async (row: ITaskDetailItem) => {
  try {
    // 触发校验
    await Promise.all(editCompRefs.value.map((item: { getValue: () => any }) => item.getValue()));

    // 更新值
    const listRow = list.value.find((item) => item.id === row.id);
    columns.forEach((col) => {
      const val = editState.value[col.field.id];
      // bool类型使用select展示，需要转换值因为select不支持bool值选项，保存时转回bool值
      if (col.field.type === 'bool') {
        editState.value[col.field.id] = typeof val !== 'boolean' ? val === 'true' : val;
      }
      // 保存时转回数组值，当数组型的值使用input时约定为使用,分隔符填写
      if (col.field.type === 'array') {
        editState.value[col.field.id] = Array.isArray(val) ? val : val?.split?.(',');
      }
    });
    Object.assign(listRow, editState.value);

    // 退出编辑状态
    editState.value = null;
  } catch (error) {
    console.error(error);
  }
};

const handleValidate = async () => {
  const params: ITaskRerunParams = {
    bk_biz_id: props.info.bk_biz_id,
    vendor: props.info.vendors?.[0],
    operation_type: props.info.operations?.[0],
    data: {
      account_id: props.info.account_ids?.[0],
      region_ids: regionIds.value,
      details: [],
    },
  };
  list.value.forEach((item) => {
    const newParam: ITaskDetailParam = {};
    for (const [k, v] of Object.entries(item.param)) {
      // 先取编辑后的值，如果没有值取原始值，原始字段可能会多于编辑的字段，多出的字段原样保留
      newParam[k] = Object.hasOwn(item, `param.${k}`) ? item[`param.${k}`] : v;
    }
    params.data.details.push(newParam);
  });

  latestParams.value = params;

  const result = await taskStore.taskRerunValidate(params);
  result?.forEach?.((item: ITaskDetailParam, index: number) => {
    list.value[index].param.status = item.status;
    list.value[index].param.validate_result = item.validate_result;
    // 删除掉编辑时添加的字段，使get方法取值时读取到上面设置的字段
    Reflect.deleteProperty(list.value[index], 'param.status');
    Reflect.deleteProperty(list.value[index], 'param.validate_result');
  });

  if (result?.length) {
    isValidatePassed.value = result.every((item: ITaskDetailParam) => item.status === 'executable');
  }

  hasChecked.value = true;
};

const handleSubmit = async () => {
  const result = await taskStore.taskRerunSubmit(latestParams.value);

  // 跳转至新任务详情页
  routerAction.redirect(
    {
      name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
      params: { id: result.task_management_id },
      query: { bizs: getBizsId() },
    },
    { reload: true },
  );

  // 关闭侧滑
  model.value = false;
};

const handleCancel = () => {
  model.value = false;
};
</script>

<template>
  <bk-sideslider v-model:is-show="model" title="失败任务重试" width="60vw">
    <template #default>
      <grid-container :column="1" :label-width="110" class="info-content">
        <grid-item v-for="field in fields" :key="field.id" :label="field.name">
          <display-value
            :property="field"
            :value="baseInfo[field.id]"
            :display="{ ...field.meta?.display, on: 'info' }"
            :vendor="baseInfo?.vendors?.[0]"
          />
        </grid-item>
      </grid-container>
      <grid-container layout="vertical" :column="1" class="params-content">
        <grid-item :label="'参数'">
          <Ediatable>
            <template #default>
              <HeadColumn v-for="(col, index) in columns" :key="index" :required="false">
                {{ col.field.name }}
              </HeadColumn>
              <HeadColumn fixed="right" :required="false" :width="100">操作</HeadColumn>
            </template>
            <template #data>
              <tr v-for="row in list" :key="row.id">
                <td v-for="col in columns" :key="col.field.id">
                  <component
                    v-if="editState?.id === row.id && col.setting.editable"
                    ref="editCompRefs"
                    v-model="editState[col.field.id]"
                    :is="`hcm-form-${col.field.type}`"
                    :option="col.field.option"
                    :display="col.setting.display"
                    :filterable="false"
                    :clearable="false"
                    :multiple="false"
                    :rules="col.setting.rules"
                  ></component>
                  <div
                    :class="[
                      'text-cell',
                      col.field.id,
                      { uneditable: !col.setting.editable, 'has-error': row.param.status !== 'executable' },
                    ]"
                    v-else
                  >
                    <display-value
                      :property="col.field"
                      :value="get(row, col.field.id)"
                      :display="{ ...col.field?.meta?.display, ...col.setting.display }"
                    />
                  </div>
                </td>
                <FixedColumn>
                  <div class="text-cell operation-cell">
                    <bk-button
                      text
                      theme="primary"
                      @click.stop="handelEdit(row)"
                      :disabled="row.param.status === 'executable' || !hasChecked"
                      v-bk-tooltips="{
                        content: hasChecked ? '校验通过，不可编辑' : '请先点击校验参数',
                        disabled: row.param.status !== 'executable',
                      }"
                      v-if="editState?.id !== row.id"
                    >
                      编辑
                    </bk-button>
                    <bk-button text theme="primary" @click.stop="handelDone(row)" v-else>完成</bk-button>
                    <bk-button text theme="primary" :disabled="list.length === 1" @click.stop="handleRemove(row)">
                      移除
                    </bk-button>
                  </div>
                </FixedColumn>
              </tr>
            </template>
          </Ediatable>
        </grid-item>
      </grid-container>
    </template>
    <template #footer>
      <div class="contnet-footer">
        <bk-button
          theme="primary"
          :loading="taskStore.taskRerunSubmitLoading"
          @click="handleSubmit"
          v-if="isValidatePassed === true"
        >
          提交
        </bk-button>
        <bk-button theme="primary" :loading="taskStore.taskRerunValidateLoading" @click="handleValidate" v-else>
          校验参数
        </bk-button>
        <bk-button @click="handleCancel">取消</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.info-content {
  margin-top: 12px;
}

.params-content {
  margin-top: 12px;
  padding: 0 32px;
}

.contnet-footer {
  display: flex;
  gap: 8px;

  .bk-button {
    min-width: 86px;
  }
}

.text-cell {
  max-width: 100%;
  padding: 1px 0 1px 16px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;

  &.uneditable {
    background: #f2f2f2;
  }

  &.operation-cell {
    display: flex;
    gap: 8px;
  }

  &.has-error {
    &.param\.validate_result {
      color: #ea3636;
    }
  }
}
</style>
