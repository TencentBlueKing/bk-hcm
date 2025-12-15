<script setup lang="ts">
import { computed, ref, useAttrs, watchEffect } from 'vue';
import { useFormItem } from 'bkui-vue/lib/shared';
import { useConfigRequirementStore, type IRequirementItem } from '@/store/config/requirement';

defineOptions({ name: 'hcm-form-req-type' });

const model = defineModel<number | number[] | string | string[]>();

const props = withDefaults(
  defineProps<{
    multiple?: boolean;
    clearable?: boolean;
    disabled?: boolean;
    useNameValue?: boolean;
    appearance?: 'card';
    filter?: (list: IRequirementItem[]) => IRequirementItem[];
  }>(),
  {
    multiple: false,
  },
);

const emit = defineEmits<(e: 'change', val: number, oldVal: number) => void>();

const formItem = useFormItem();
const attrs = useAttrs();

const list = ref<IRequirementItem[]>([]);

const localModel = computed({
  get() {
    if (props.multiple && model.value && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(value) {
    if (!props.useNameValue) {
      const newVal = Array.isArray(value) ? value.map((val) => Number(val)) : Number(value);
      model.value = newVal;
    } else {
      model.value = value as string | string[];
    }
  },
});

const configRequirementStore = useConfigRequirementStore();

watchEffect(async () => {
  list.value = await configRequirementStore.getRequirementType();
  if (props.filter) {
    list.value = props.filter(list.value);
  }
});

const options: Record<IRequirementItem['require_type'], { tags: string[]; icon: string; recommend: boolean }> = {
  1: { tags: ['提前预测', '全业务场景'], icon: 'bkhcm-icon-regular', recommend: false },
  2: { tags: ['专项资源', '限期申请'], icon: 'bkhcm-icon-lantern', recommend: false },
  3: { tags: ['专项资源', '限期申请'], icon: 'bkhcm-icon-host-multi', recommend: false },
  6: { tags: ['即时申领', '资源周转'], icon: 'bkhcm-icon-rolling-server', recommend: true },
  7: { tags: ['即时申领', '小额紧急'], icon: 'bkhcm-icon-green-channel-rocket', recommend: true },
  8: { tags: ['2026春节紧急资源', '按量计费'], icon: 'bkhcm-icon-spring-pool', recommend: false },
  9: { tags: ['提前预测', '短租场景'], icon: 'bkhcm-icon-short-rental', recommend: false },
};

const handleSelect = (item: IRequirementItem) => {
  if (model.value === item.require_type) {
    return;
  }
  emit('change', item.require_type, model.value as number);
  model.value = item.require_type;
  formItem?.validate('change');
};
</script>

<template>
  <bk-select
    v-if="!appearance"
    v-model="localModel"
    :list="list"
    :clearable="clearable"
    :multiple="multiple"
    :multiple-mode="multiple ? 'tag' : 'default'"
    :id-key="!useNameValue ? 'require_type' : 'require_name'"
    :display-key="'require_name'"
    v-bind="attrs"
  />
  <div v-else-if="appearance === 'card'" class="req-type-card">
    <div
      :class="['card-item', { selected: req.require_type === model }]"
      v-for="req in list"
      :key="req.require_type"
      @click="handleSelect(req)"
    >
      <div class="type-icon">
        <i :class="['hcm-icon', options[req.require_type]?.icon || 'bkhcm-icon-regular', 'req-type-icon']"></i>
      </div>
      <div class="type-name">{{ req.require_name }}</div>
      <div class="tag-list">
        <div class="tag-item" v-for="(tag, index) in options[req.require_type]?.tags" :key="index">{{ tag }}</div>
      </div>
      <div class="recommend-tag" v-if="options[req.require_type]?.recommend">推荐</div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.req-type-card {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;

  .card-item {
    position: relative;
    display: flex;
    flex-direction: column;
    width: 200px;
    height: 102px;
    background: #fff;
    border: 1px solid #dcdee5;
    padding: 12px;
    border-radius: 2px;
    background-image: url('@/assets/image/req-type-card-bg.svg'), linear-gradient(180deg, #d8e7ff 0%, #fff 65%, #fff);
    background-size: contain, cover;
    background-repeat: no-repeat;
    cursor: pointer;

    &:hover {
      border: 1px solid #a3c5fd;
      box-shadow: 0 2px 6px 0 #0000001a;
    }

    &.selected {
      padding: 11px;
      border: 2px solid #3a84ff;

      .recommend-tag {
        top: -2px;
        right: -2px;
      }
    }

    .type-icon {
      line-height: normal;

      .req-type-icon {
        font-size: 18px;
        color: #699df4;
      }
    }

    .type-name {
      font-size: 14px;
      color: #313238;
    }

    .recommend-tag {
      display: flex;
      align-items: center;
      font-size: 12px;
      color: #fff;
      background: #f59500;
      border-radius: 2px;
      position: absolute;
      top: -1px;
      right: -1px;
      height: 16px;
      padding: 0 6px;
    }
  }

  .tag-list {
    display: flex;
    gap: 6px;
    margin-top: auto;

    .tag-item {
      display: flex;
      height: 22px;
      align-items: center;
      justify-content: center;
      font-size: 12px;
      color: #299e56;
      background: #daf6e5;
      border-radius: 11px;
      padding: 0 6px;
      white-space: nowrap;
    }
  }
}
</style>
