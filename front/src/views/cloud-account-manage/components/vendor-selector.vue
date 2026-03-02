<script setup lang="ts">
import { computed, ref } from 'vue';
import { AngleDown } from 'bkui-vue/lib/icon';
import { VendorEnum, VendorMap } from '@/common/constant';
import tcloudVendorIcon from '@/assets/image/vendor-tcloud.svg';
import awsVendorIcon from '@/assets/image/vendor-aws.svg';
import azureVendorIcon from '@/assets/image/vendor-azure.svg';
import gcpVendorIcon from '@/assets/image/vendor-gcp.svg';
import huaweiVendorIcon from '@/assets/image/vendor-huawei.svg';

interface IProps {
  modelValue?: VendorEnum;
  disabled?: boolean;
}

const props = withDefaults(defineProps<IProps>(), {
  modelValue: VendorEnum.TCLOUD,
  disabled: true, // 一期只支持腾讯云，默认禁用切换
});

const emit = defineEmits<{
  'update:modelValue': [value: VendorEnum];
  change: [value: VendorEnum];
}>();

// 下拉框展开状态
const isOpen = ref(false);

// 展开/收起事件处理
const handleToggle = (val: boolean) => {
  isOpen.value = val;
};

// 云厂商图标映射（仅支持的厂商）
const vendorIconMap: Partial<Record<VendorEnum, string>> = {
  [VendorEnum.TCLOUD]: tcloudVendorIcon,
  [VendorEnum.AWS]: awsVendorIcon,
  [VendorEnum.GCP]: gcpVendorIcon,
  [VendorEnum.AZURE]: azureVendorIcon,
  [VendorEnum.HUAWEI]: huaweiVendorIcon,
};

// 云厂商列表配置
const vendorList = computed(() => [
  {
    id: VendorEnum.TCLOUD,
    name: VendorMap[VendorEnum.TCLOUD],
    icon: vendorIconMap[VendorEnum.TCLOUD],
  },
  {
    id: VendorEnum.AWS,
    name: VendorMap[VendorEnum.AWS],
    icon: vendorIconMap[VendorEnum.AWS],
  },
  {
    id: VendorEnum.GCP,
    name: VendorMap[VendorEnum.GCP],
    icon: vendorIconMap[VendorEnum.GCP],
  },
  {
    id: VendorEnum.AZURE,
    name: VendorMap[VendorEnum.AZURE],
    icon: vendorIconMap[VendorEnum.AZURE],
  },
  {
    id: VendorEnum.HUAWEI,
    name: VendorMap[VendorEnum.HUAWEI],
    icon: vendorIconMap[VendorEnum.HUAWEI],
  },
]);

// 当前选中的云厂商
const currentVendor = computed(() => {
  return vendorList.value.find((v) => v.id === props.modelValue) || vendorList.value[0];
});

// 选择变化处理
const handleChange = (value: VendorEnum) => {
  emit('update:modelValue', value);
  emit('change', value);
};
</script>

<template>
  <bk-select
    :model-value="modelValue"
    :list="vendorList"
    :disabled="disabled"
    :clearable="false"
    id-key="id"
    display-key="name"
    class="vendor-selector"
    @change="handleChange"
    @toggle="handleToggle"
  >
    <!-- 自定义触发器：显示图标 + 名称 + 箭头 -->
    <template #trigger>
      <div class="vendor-trigger">
        <div class="bk-select-tag" :class="{ 'is-disabled': disabled }" style="width: 100%; padding-right: 5px">
          <img :src="currentVendor.icon" class="vendor-icon" alt="" />
          <span class="vendor-name">{{ currentVendor.name }}</span>
          <AngleDown class="arrow-icon" :class="{ 'is-open': isOpen }" />
        </div>
      </div>
    </template>

    <!-- 自定义选项渲染：显示图标 + 名称 -->
    <template #optionRender="{ item }">
      <div class="vendor-option">
        <img :src="item.icon" class="vendor-icon" alt="" />
        <span class="vendor-name">{{ item.name }}</span>
      </div>
    </template>
  </bk-select>
</template>

<style lang="scss" scoped>
.vendor-selector {
  width: 120px;
  padding: 4px 0;
}

.bk-select-tag {
  display: flex;
  align-items: center;
}

.vendor-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 100%;
  cursor: pointer;

  .vendor-icon {
    width: 16px;
    height: 16px;
    flex-shrink: 0;
    margin-right: 8px;
  }

  .vendor-name {
    font-size: 12px;
    color: #63656e;
    flex: 1;
  }

  .arrow-icon {
    font-size: 20px;
    color: #979ba5;
    flex-shrink: 0;
    transition: transform 0.2s ease-in-out;

    &.is-open {
      transform: rotate(180deg);
    }
  }
}

.vendor-option {
  display: flex;
  align-items: center;
  gap: 8px;

  .vendor-icon {
    width: 16px;
    height: 16px;
    flex-shrink: 0;
  }

  .vendor-name {
    font-size: 12px;
    color: #63656e;
  }
}
</style>
