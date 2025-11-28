<script setup lang="ts">
import { InputColumn } from '@blueking/ediatable';
import { computed, inject, nextTick, ref, watch } from 'vue';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import http from '@/http';
import { useRoute } from 'vue-router';
import { Message } from 'bkui-vue';
import { type ICvmDevicetypeItem, CoreTypeMap } from '@/store/cvm/device';

const model = defineModel<boolean>();
const props = defineProps<{ tableData: any[]; deviceTypeList: ICvmDevicetypeItem[] }>();
const emit = defineEmits(['update:table']);
const route = useRoute();
const cloudColumns = inject<any[]>('cloudColumns', []);
const retainSource = ref(false);

// 列配置
const mapColumns = (columns: any[], filter?: (col: any, idx: number) => any) =>
  columns.map((item, idx) => ({
    ...cloudColumns.find((v: any) => v.field === item.id),
    ...item,
    ...(filter ? filter(item, idx) : {}),
  }));

const filterColumns = [
  { name: '机型', id: 'spec.device_type' },
  { name: '计费模式', id: 'spec.charge_type' },
  { name: '地域', id: 'spec.region' },
  { name: '可用区', id: 'spec.zone' },
  { name: '操作系统', id: 'spec.image_id' },
  { name: '数据盘', id: 'spec.data_disk' },
  { name: 'VPC', id: 'spec.vpc' },
  { name: '子网', id: 'spec.subnet' },
  { name: '原始需求数量', id: 'total_num', sort: true, width: 150, fixed: 'right' },
  { name: '修改后需求数量', id: 'modify_num', width: 150, fixed: 'right' },
];
const resourceFilterColumns = [
  { name: '机型', id: 'spec.device_type' },
  { name: '可用区', id: 'spec.zone' },
  { name: '需求数量', id: 'replicas', type: 'string', width: 100 },
  { name: '地域', id: 'spec.region' },
  { name: '计费模式', id: 'spec.charge_type' },
  { name: '操作系统', id: 'spec.image_id' },
  { name: '数据盘', id: 'spec.data_disk' },
  { name: 'VPC', id: 'spec.vpc' },
  { name: '子网', id: 'spec.subnet' },
];

const columns = mapColumns(filterColumns);
const resourceColumns = mapColumns(resourceFilterColumns, (_, idx) => (idx <= 1 ? { render: null } : {}));

// 数据处理
const modifyTableData = ref([]);
const resourceTableData = ref([]);

// 初始化修改表格数据
const initModifyTableData = () => {
  modifyTableData.value = JSON.parse(JSON.stringify(props.tableData)).map((item: any, index: number) => ({
    ...item,
    source: 'business',
    id: index, // 记录修改前的位置
  }));
};
watch(() => props.tableData, initModifyTableData, { immediate: true });
watch(
  [modifyTableData, () => props.deviceTypeList],
  ([modifyData, deviceTypeList]) => {
    resourceTableData.value = JSON.parse(JSON.stringify(modifyData))
      .filter((item: any) => item.modify_num)
      .map((item: any) => {
        const num = item.replicas || item.total_num;
        const deviceType = deviceTypeList.find((v) => v.device_type === item.spec.device_type);
        const remainder = Number(num) - Number(item.modify_num);
        return {
          ...item,
          replicas: remainder,
          source: 'purchase_to_resource_pool',
          pageData: {
            remainderCPUAmount: deviceType.cpu_amount * remainder,
            deviceType,
          },
        };
      });
  },
  { deep: true },
);

// 校验相关
const replicasValidArr = computed(() => modifyTableData.value.map(() => ref()));
const rules = {
  modifyNum: [{ validator: (v: string) => !!v, message: '不能为空或0' }],
  zones: [{ validator: (v: string[]) => !!v.length, message: '请选择可用区' }],
  deviceType: [{ validator: (v: string) => !!v, message: '请选择机型' }],
  replicas: [{ validator: (v: string) => !!v, message: '需求数量需大于0' }],
};

// 工具方法
const removeUnchangeItem = () => {
  modifyTableData.value = modifyTableData.value.filter((item) => item.modify_num);
};
const handleDevicetypeChange = (device: any, row: any, index: number) => {
  row.replicas = Math.floor(row.pageData.remainderCPUAmount / device.cpu_amount);
  row.pageData.deviceType = device;
  nextTick(() => replicasValidArr.value[index].value?.getValue());
};

// 提交状态校验
const hasInvalidModifyNum = computed(() =>
  modifyTableData.value.some((item) => !item.modify_num || item.modify_num === 0),
);
const hasInvalidReplicas = computed(() =>
  resourceTableData.value.some((item) => !item.replicas || item.replicas === 0),
);
const hasInvalidZones = computed(() => resourceTableData.value.some((item) => !item.spec?.zones?.length));
const submitDisabled = computed(
  () =>
    !modifyTableData.value.length ||
    hasInvalidModifyNum.value ||
    (retainSource.value && (hasInvalidReplicas.value || hasInvalidZones.value)),
);

const isSubmitting = ref(false);

// 提交处理
const handleSubmit = async () => {
  const orderId = Number(route.params.id);
  const mData = props.tableData.map((item, idx) => {
    const { modify_num, id, ...rest } = modifyTableData.value.find((v) => v.id === idx) ?? {};
    const deviceType = props.deviceTypeList.find((v) => v.device_type === rest?.spec?.device_type);
    return modify_num
      ? {
          ...rest,
          replicas: Number(modify_num),
          applied_core: deviceType?.cpu_amount * Number(modify_num),
        }
      : { ...item, source: 'business' };
  });
  const rData = retainSource.value
    ? resourceTableData.value.map(({ pageData, ...rest }) => ({
        ...rest,
        applied_core: pageData.deviceType.cpu_amount * rest.replicas,
      }))
    : [];

  isSubmitting.value = true;
  try {
    await http.post('/api/v1/woa/task/apply/ticket/demand/update', {
      ticket_id: orderId,
      suborders: [...mData, ...rData],
    });
    Message({ theme: 'success', message: '修改成功' });
    handleCancel();
    emit('update:table');
  } catch {
    Message({ theme: 'error', message: '修改失败' });
  } finally {
    isSubmitting.value = false;
  }
};

const handleCancel = () => {
  initModifyTableData();
  retainSource.value = false;
  model.value = false;
};
</script>

<template>
  <!-- eslint-disable vue/valid-v-slot -->
  <bk-sideslider
    v-model:is-show="model"
    title="修改单据"
    width="1200"
    class="edit-cvm-sideslider"
    @closed="handleCancel"
  >
    <div class="container">
      <div class="title">
        <h4>修改原始需求配置</h4>
        <div>
          <span class="mr12 text">采购到资源池配置</span>
          <bk-switcher v-model="retainSource" theme="primary" />
        </div>
      </div>

      <bk-button class="mb16" @click="removeUnchangeItem">移除未修改项</bk-button>

      <!-- 修改原始需求配置 -->
      <data-list-table :settings="false" :columns="columns" :list="modifyTableData">
        <template #modify_num="{ row }">
          <div class="modify-column">
            <InputColumn
              v-if="(row.replicas || row.total_num) > 1"
              :class="{ changed: !!row.modify_num }"
              type="number"
              :min="1"
              :max="(row.replicas || row.total_num) - 1"
              :rules="rules.modifyNum"
              clearable
              v-model="row.modify_num"
            />
            <span v-else>--</span>
          </div>
        </template>
      </data-list-table>

      <!-- 资源池资源分配 -->
      <template v-if="retainSource">
        <div class="title mt24">
          <h4>采购到资源池</h4>
        </div>
        <data-list-table :settings="false" :columns="resourceColumns" :list="resourceTableData">
          <template #spec.device_type="{ row, index }">
            <div class="modify-column" v-if="row?.spec">
              <devicetype-selector
                v-model="row.spec.device_type"
                resource-type="cvm"
                :clearable="false"
                :params="{
                  require_type: row.require_type,
                  region: row.spec.region,
                  device_size: CoreTypeMap[row.pageData.deviceType.core_type as keyof typeof CoreTypeMap],
                  technical_class: row.pageData.deviceType.technical_class,
                }"
                :editable="true"
                :rules="rules.deviceType"
                @change="handleDevicetypeChange($event, row, index)"
              />
            </div>
          </template>
          <template #spec.zone="{ row }">
            <div class="modify-column" v-if="row?.spec">
              <zone-selector
                v-model="row.spec.zones"
                :clearable="false"
                multiple
                :separate-campus="false"
                :params="{ resourceType: 'QCLOUDCVM', region: row?.spec?.region }"
                :editable="true"
                :rules="rules.zones"
              />
            </div>
          </template>
          <template #replicas="{ row, index }">
            <div class="modify-column">
              <InputColumn
                :ref="replicasValidArr[index]"
                v-model="row.replicas"
                :rules="rules.replicas"
                readonly
                :clearable="false"
              />
            </div>
          </template>
        </data-list-table>
      </template>
    </div>

    <template #footer>
      <bk-pop-confirm
        width="260"
        :popover-options="{
          hideIgnoreReference: true,
        }"
        content="请确认配置参数是否正确"
        title="确认提交？"
        trigger="click"
        @confirm="handleSubmit"
      >
        <bk-button
          style="width: 88px"
          theme="primary"
          :disabled="submitDisabled"
          :loading="isSubmitting"
          v-bk-tooltips="{ content: '请移除未修改项后提交', disabled: !hasInvalidModifyNum }"
        >
          提交
        </bk-button>
        <bk-button style="margin-left: 8px; width: 88px" @click="handleCancel">取消</bk-button>
      </bk-pop-confirm>
    </template>
  </bk-sideslider>
</template>

<style scoped lang="scss">
.container {
  padding: 28px 40px 0;

  // 使用footer吸底
  min-height: calc(100vh - 100px);
}

.title {
  display: flex;
  align-items: center;
  margin-bottom: 16px;

  h4 {
    color: #313238;
    margin-right: 32px;
  }

  .text {
    color: #4d4f56;
  }

  & > div {
    display: flex;
    align-items: center;
  }
}

.modify-column {
  margin: 0 -16px;

  .changed:deep(input) {
    background-color: #fdeed8;
    font-weight: 700;
    color: #f59500;
  }
}
</style>
