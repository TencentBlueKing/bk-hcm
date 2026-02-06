<template>
  <bk-sideslider v-model:is-show="model" :title="title" quick-close width="640" @hidden="emit('hidden')">
    <template #default>
      <div class="container">
        <bk-alert theme="info" class="mb16">
          <template #title>支持部分资源延期，其余需求时间不变。可调整"实例数量"或"期望到货日期"</template>
        </bk-alert>
        <section class="panel mb24">
          <p class="title">资源信息</p>
          <grid-container :column="2" label-width="100" :content-min-width="100" :content-max-width="180">
            <grid-item v-for="field in fields" :key="field.id" :label="field.name">
              {{ field.format ? field.format(data[field.id]) : data[field.id] }}
            </grid-item>
          </grid-container>
        </section>
        <section class="panel">
          <p class="title">部分延期</p>
          <bk-form form-type="vertical" class="postpone-form">
            <bk-form-item label="延期期望到货日期">
              <hcm-form-datetime
                v-model="formModel.expect_time"
                append-to-body
                format="yyyy-MM-dd"
                class="full-width"
                :disabled-date="(date: number | Date) => dayjs(date).isBefore(dayjs())"
              />
              <div class="desc-info">原期望到货日期 {{ data.expect_time }}</div>
            </bk-form-item>
            <grid-container :column="2" :gap="[0, 24]">
              <bk-form-item label="延期实例数量">
                <hcm-form-number v-model="formModel.delay_os" :min="0" :max="Number(data.remained_os)" precision />
                <div class="desc-info">
                  其余
                  <span class="primary-text">{{ Number(data.remained_os) - formModel.delay_os }}</span>
                  实例数按原期望日期到货
                </div>
              </bk-form-item>
              <bk-form-item label="延期CPU总核数">
                <div class="readonly-value">{{ delayTotalCpuCore }}</div>
                <div class="desc-info">
                  其余
                  <span class="primary-text">{{ data.remained_cpu_core - delayTotalCpuCore }}</span>
                  CPU总核数按原期望日期到货
                </div>
              </bk-form-item>
              <bk-form-item label="延期内存总数">
                <div class="readonly-value">{{ delayTotalMemory }} GB</div>
              </bk-form-item>
              <bk-form-item label="延期云盘总数">
                <div class="readonly-value">{{ delayTotalDisk }} GB</div>
              </bk-form-item>
            </grid-container>
          </bk-form>
        </section>
      </div>
    </template>
    <template #footer>
      <bk-button theme="primary" class="button mr8" @click="handleConfirm">提交延期</bk-button>
      <bk-button class="button" @click="handleClosed">取消</bk-button>
    </template>
  </bk-sideslider>
</template>

<script setup lang="ts">
import { computed, onBeforeMount, reactive, ref, watchEffect } from 'vue';
import { useRouter } from 'vue-router';
import usePlanStore from '@/store/usePlanStore';
import { useCvmDeviceStore } from '@/store/cvm/device';
import { IListResourcesDemandsItem } from '@/typings/resourcePlan';
import { CvmDeviceType } from '@/views/ziyanScr/components/devicetype-selector/types';
import { AdjustType } from '@/typings/plan';
import { QueryRuleOPEnum } from '@/typings';
import { timeFormatter } from '@/common/util';
import dayjs from 'dayjs';

import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import { MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS } from '@/constants/menu-symbol';

interface IProps {
  data: IListResourcesDemandsItem;
}

const model = defineModel<boolean>();
const props = defineProps<IProps>();
const emit = defineEmits(['hidden']);

const router = useRouter();
const planStore = usePlanStore();
const cvmDeviceStore = useCvmDeviceStore();

const title = computed(() => `部分延期配置 - ${props.data.demand_id}`);

const fields: Array<{ id: keyof IListResourcesDemandsItem; name: string; format?: (v: any) => string }> = [
  { id: 'region_name', name: '城市' },
  { id: 'obs_project', name: '项目类型' },
  { id: 'zone_name', name: '可用区' },
  { id: 'remained_os', name: '实例数' },
  { id: 'demand_class', name: '类型' },
  { id: 'remained_cpu_core', name: 'CPU总核数' },
  { id: 'device_type', name: '机型规格' },
  { id: 'remained_memory', name: '内存总量', format: (v: number) => `${v} GB` },
  { id: 'expect_time', name: '期望到货日期' },
  { id: 'remained_disk_size', name: '云盘总量', format: (v: number) => `${v} GB` },
  { id: 'disk_type_name', name: '云盘类型' },
];

const formModel = reactive({
  expect_time: dayjs().add(13, 'week').format('YYYY-MM-DD'),
  delay_os: Number(props.data.remained_os),
});

watchEffect(() => {
  // input组件在针对小数的最大禁用行为有问题，这里手动处理
  if (formModel.delay_os > Number(props.data.remained_os)) {
    formModel.delay_os = Number(props.data.remained_os);
  }
});

const devicetypeInfo = ref<CvmDeviceType>(null);
const getDevicetypeInfo = async () => {
  devicetypeInfo.value = await cvmDeviceStore.getOneDevicetype({
    filter: {
      op: 'and',
      rules: [
        { field: 'zone', op: QueryRuleOPEnum.EQ, value: props.data.zone_id },
        { field: 'device_type', op: QueryRuleOPEnum.EQ, value: props.data.device_type },
      ],
    },
  });
};

const delayTotalCpuCore = computed(() => formModel.delay_os * (devicetypeInfo.value?.cpu_core ?? 0));
const delayTotalMemory = computed(() => formModel.delay_os * (devicetypeInfo.value?.memory ?? 0));
const delayTotalDisk = computed(
  () => formModel.delay_os * (props.data.remained_disk_size / Number(props.data.remained_os)),
);
const handleConfirm = async () => {
  const originDetail = { ...props.data, res_mode: '按机型', demand_source: '指标变化' };
  const updatedDetail = {
    ...originDetail,
    adjustType: AdjustType.time,
    expect_time: timeFormatter(formModel.expect_time, 'YYYY-MM-DD'),
  };
  const info = planStore.convertToAdjust(originDetail, updatedDetail, String(formModel.delay_os));
  const { data } = await planStore.adjust_biz_resource_plan_demand({ adjusts: [info] });
  if (!data.id) return;
  router.push({
    name: MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS,
    query: { id: data.id },
  });
};

const handleClosed = () => {
  model.value = false;
};

onBeforeMount(() => {
  getDevicetypeInfo();
});
</script>

<style scoped lang="scss">
.container {
  padding: 20px 40px;

  .panel {
    .title {
      margin-bottom: 16px;
      font-size: 16px;
      color: #313238;
      font-weight: 700;
    }

    .postpone-form {
      .desc-info {
        font-size: 12px;
        color: #979ba5;
      }

      .readonly-value {
        padding: 0 10px;
        background: #f0f1f5;
        border-radius: 2px;
        color: #4d4f56;
      }

      .primary-text {
        color: #3a84ff;
        font-weight: 700;
      }
    }
  }
}

.button {
  min-width: 88px;
}
</style>
