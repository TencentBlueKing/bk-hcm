<script setup lang="ts">
import http from '@/http';
import { Select } from 'bkui-vue';
import { computed, ref, watch } from 'vue';

export interface ICvmVpc {
  id: number;
  region: string;
  vpc_id: string;
  vpc_name: string;
}
type ICvmVpcList = Array<ICvmVpc>;

defineOptions({ name: 'CvmVpcSelector' });

const model = defineModel<string>();

const props = defineProps<{ region: string; disabled: boolean }>();

const emit = defineEmits<(e: 'change', val: ICvmVpc) => void>();

const { Option } = Select;

const optionList = ref<ICvmVpcList>([]);

const selectedId = computed({
  get() {
    // 接口入参vpc_id, select值为id, 此处做转换
    const selectedItem = optionList.value.find((item) => item.vpc_id === model.value);
    return selectedItem?.id || undefined;
  },
  set(val) {
    const selectedItem = optionList.value.find((item) => item.id === val);
    emit('change', selectedItem);
    // 接口入参vpc_id, select值为id, 此处做转换
    model.value = selectedItem?.vpc_id;
  },
});

const displayOptionList = computed(() => {
  const priorityVpcName = [
    'VPC-IEG-SZ',
    'VPC-IEG-SH',
    'VPC-IEG-GZ',
    'VPC-IEG-JN-EC',
    'VPC-IEG-HZ-EC',
    'VPC-IEG-NJ',
    'VPC-IEG-FZ',
    'VPC-IEG-HF-EC',
    'VPC-IEG-BJ',
    'VPC-IEG-TJ',
    'VPC-IEG-SJZ-EC',
    'VPC-IEG-WH-EC',
    'VPC-IEG-CS-EC',
    'VPC-IEF-CS-EC',
    'VPC-IEG-ZZ-EC',
    'VPC-IEG-CD',
    'VPC-IEG-CQ',
    'VPC-IEG-XA-EC',
    'VPC-IEG-SY-EC',
    'VPC-IEG-HK',
    'VPC-IEG-SR',
    'VPC-IEG-DJ',
    'VPC-IEG-XJP',
    'VPC-IEG-MG',
    'VPC-IEG-GG',
    'VPC-IEG-FLKF',
    'VPC-IEG-MM',
    'VPC-IEG-FJNY',
    'VPC-IEG-SBL',
  ];
  // 按 priorityVpcName 排序（完整匹配），让这些优先出现在前面，其他按原顺序排在后面
  return [...optionList.value].sort((a, b) => {
    const aIndex = priorityVpcName.findIndex((name) => a.vpc_name === name);
    const bIndex = priorityVpcName.findIndex((name) => b.vpc_name === name);
    if (aIndex === -1 && bIndex === -1) {
      return 0;
    }
    if (aIndex === -1) {
      return 1;
    }
    if (bIndex === -1) {
      return -1;
    }
    return aIndex - bIndex;
  });
});

const findCvmVpcByVpcId = (vpc_id: string) => {
  return optionList.value.find((item) => item.vpc_id === vpc_id);
};

const getOptionList = async (region: string) => {
  const res = await http.post('/api/v1/woa/config/findmany/config/cvm/vpc', { region });
  optionList.value = res.data.info;
};

watch(
  () => props.region,
  (val) => {
    val && getOptionList(val);
  },
  { immediate: true },
);

defineExpose({ findCvmVpcByVpcId });
</script>

<template>
  <Select class="w600" v-model="selectedId" :disabled="props.disabled">
    <Option
      v-for="{ id, vpc_id: vpcId, vpc_name: vpcName } in displayOptionList"
      :key="id"
      :id="id"
      :name="`${vpcId} | ${vpcName}`"
    />
  </Select>
</template>
