<script setup lang="ts">
import { computed, ref, useAttrs, watchEffect } from 'vue';
import { ModelProperty } from '@/model/typings';
import { SelectColumn } from '@blueking/ediatable';
import { DisplayType } from './typings';

defineOptions({ name: 'hcm-form-list' });

const model = defineModel<string | number | (string | number)[]>();
const props = withDefaults(
  defineProps<{
    list: ModelProperty['list'];
    clearable?: boolean;
    multiple?: boolean;
    display?: DisplayType;
  }>(),
  {
    clearable: false,
    multiple: false,
  },
);
const attrs = useAttrs();

const comp = computed(() => (props.display?.on === 'cell' ? SelectColumn : 'bk-select'));

const localList = ref<ModelProperty['list']>([]);
const loading = ref(false);

watchEffect(async () => {
  if (typeof props.list === 'function') {
    loading.value = true;
    try {
      localList.value = await props.list();
    } finally {
      loading.value = false;
    }
  } else {
    localList.value = props.list;
  }
});

const localModel = computed({
  get() {
    if (props.multiple && model.value && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(val) {
    model.value = val;
  },
});

const selectColumnRef = ref();

defineExpose({
  getValue() {
    if (selectColumnRef.value?.getValue) {
      return selectColumnRef.value.getValue().then(() => model.value);
    }
    return model.value;
  },
});
</script>

<template>
  <component
    :is="comp"
    v-model="localModel"
    ref="selectColumnRef"
    :list="localList"
    :clearable="clearable"
    :multiple="multiple"
    :multiple-mode="multiple ? 'tag' : 'default'"
    :loading="loading"
    v-bind="attrs"
  />
</template>
