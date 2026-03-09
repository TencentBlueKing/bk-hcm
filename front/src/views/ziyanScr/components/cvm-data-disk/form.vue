<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { ICvmDataDisk, ICvmDataDiskOption } from './typings';
import { CVM_DATA_DISK_INFO, CvmDataDiskType } from './constants';
import { ICloudInstanceConfigItem } from '@/typings/ziyanScr';
import http from '@/http';
import { useFormItem } from 'bkui-vue/lib/form';

import { Plus } from 'bkui-vue/lib/icon';

const model = defineModel<ICvmDataDisk[]>();
const props = defineProps<{
  isItDeviceType: boolean;
  currentCloudInstanceConfig: ICloudInstanceConfigItem;
}>();

const formItem = useFormItem();

// 本地盘类型列表
const localDiskTypes = [
  CvmDataDiskType.LOCAL_BASIC,
  CvmDataDiskType.LOCAL_SSD,
  CvmDataDiskType.LOCAL_NVME,
  CvmDataDiskType.LOCAL_NVME_BASIC,
  CvmDataDiskType.LOCAL_PRO,
];
const isLocalDiskType = (type: CvmDataDiskType) => localDiskTypes.includes(type);

// IT机型，并且local_disk_type_list有值，则可以选本地盘（与系统盘组件逻辑一致）
const hasLocalDisk = computed(
  () => props.isItDeviceType && props.currentCloudInstanceConfig?.local_disk_type_list?.length,
);

// 监听机型变化，当机型不支持本地盘时，清空已选择的本地盘数据盘
watch(hasLocalDisk, (val) => {
  if (!val && model.value?.length) {
    model.value = model.value.filter((disk) => !isLocalDiskType(disk.disk_type));
  }
});

const loading = ref(false);
const dataDiskOptions = ref<ICvmDataDiskOption[]>([]);
onMounted(async () => {
  loading.value = true;
  try {
    const res = await http.get('/api/v1/woa/config/find/config/cvm/disktype');
    dataDiskOptions.value = res.data?.info ?? [];
  } catch (error) {
  } finally {
    loading.value = false;
  }
});

const storageBlockAttr = computed(() => props.currentCloudInstanceConfig?.externals?.storage_block_attr);
const storageBlockAmount = computed(() => props.currentCloudInstanceConfig?.storage_block_amount);

const diskTotalCount = computed(() => model.value.reduce((pre, cur) => pre + cur.disk_num, 0));

const handleDiskTypeChange = (val: CvmDataDiskType, index: number) => {
  model.value[index].disk_size = val ? CVM_DATA_DISK_INFO[val].min : 0;
};

const handleAdd = () => {
  model.value.push({ disk_type: undefined, disk_size: 0, disk_num: 1 });
};

const handleRemove = (index: number) => {
  model.value.splice(index, 1);
};

watch(model, () => formItem?.validate('change'), { deep: true });
</script>

<template>
  <div v-if="isItDeviceType && storageBlockAttr && storageBlockAmount" class="cvm-data-disk-item mb8">
    <bk-input class="form-control" :model-value="CVM_DATA_DISK_INFO[storageBlockAttr.type].disk_name" disabled />
    <hcm-form-number :model-value="storageBlockAttr.min_size" class="form-control" prefix="大小" suffix="GB" disabled />
    <hcm-form-number :model-value="storageBlockAmount" class="form-control small" suffix="块" disabled />
  </div>
  <bk-button v-if="model.length === 0" @click="handleAdd">
    <plus class="f24" />
  </bk-button>
  <div v-else class="cvm-data-disk-list">
    <div v-for="(dataDisk, index) in model" :key="`${dataDisk.disk_type}${index}`" class="cvm-data-disk-item">
      <bk-select
        v-model="dataDisk.disk_type"
        :popover-options="{ boundary: 'parent' }"
        class="form-control"
        @change="(val: CvmDataDiskType) => handleDiskTypeChange(val, index)"
      >
        <bk-option v-for="disk in dataDiskOptions" :key="disk.disk_type" :id="disk.disk_type" :name="disk.disk_name" />
      </bk-select>
      <hcm-form-number
        v-model="dataDisk.disk_size"
        class="form-control"
        prefix="大小"
        suffix="GB"
        :step="10"
        :max="CVM_DATA_DISK_INFO[dataDisk.disk_type]?.max ?? 32000"
        :min="CVM_DATA_DISK_INFO[dataDisk.disk_type]?.min ?? 0"
      />
      <hcm-form-number
        v-model="dataDisk.disk_num"
        class="form-control small"
        suffix="块"
        :max="20 - diskTotalCount + dataDisk.disk_num"
        :min="0"
      />
      <bk-button class="button" text :disabled="diskTotalCount === 20" @click="handleAdd">
        <i class="hcm-icon bkhcm-icon-plus-circle-shape"></i>
      </bk-button>
      <bk-button class="button" text @click="handleRemove(index)">
        <i class="hcm-icon bkhcm-icon-minus-circle-shape"></i>
      </bk-button>
      <i
        v-if="dataDisk.disk_type === CvmDataDiskType.CLOUD_SSD"
        class="hcm-icon bkhcm-icon-prompt text-gray cursor ml4 f16"
        v-bk-tooltips="{ content: 'SSD 云硬盘的运营成本约为高性能云盘的 4 倍，请合理评估使用。' }"
      ></i>
    </div>
  </div>
</template>

<style scoped lang="scss">
.f16 {
  font-size: 16px;
}

.f24 {
  font-size: 24px;
}

.cvm-data-disk-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cvm-data-disk-item {
  display: flex;
  align-items: center;
  gap: 8px;

  .form-control {
    width: 240px;

    &.small {
      width: 160px;
    }
  }

  .button {
    margin: 0 5px;

    .hcm-icon {
      color: #c4c6cc;
    }

    &.is-disabled {
      .hcm-icon {
        color: #eaebf0;
      }
    }
  }
}
</style>
