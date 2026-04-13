<script setup lang="ts">
import { computed, ref, useAttrs, watchEffect } from 'vue';
import { ModelProperty } from '@/model/typings';
import { SelectColumn } from '@blueking/ediatable';
import { DisplayType } from './typings';

export type ListGeneratorFactory<T = Record<string, any>> = (keyword?: string) => AsyncGenerator<T[], void>;

defineOptions({ name: 'hcm-form-list' });

const model = defineModel<string | number | (string | number)[]>();
const props = withDefaults(
  defineProps<{
    list?: ModelProperty['list'];
    listGenerator?: ListGeneratorFactory;
    clearable?: boolean;
    multiple?: boolean;
    display?: DisplayType;
  }>(),
  {
    clearable: false,
    multiple: false,
  },
);
const emit = defineEmits<{
  change: [value: string | number, item: Record<string, any> | undefined];
}>();
const attrs = useAttrs();

const comp = computed(() => (props.display?.on === 'cell' ? SelectColumn : 'bk-select'));

const isGeneratorMode = computed(() => !!props.listGenerator);
const localList = ref<Array<Record<string, any>>>([]);
const loading = ref(false);
const scrollLoading = ref(false);
let currentGenerator: AsyncGenerator<Record<string, any>[], void> | null = null;

const loadFirstPage = async (generator: AsyncGenerator<Record<string, any>[], void>) => {
  const result = await generator.next();
  localList.value = result.done ? [] : (result.value as Record<string, any>[]);
};

watchEffect(async () => {
  if (isGeneratorMode.value) {
    loading.value = true;
    try {
      currentGenerator = props.listGenerator!();
      await loadFirstPage(currentGenerator);
    } finally {
      loading.value = false;
    }
  } else if (typeof props.list === 'function') {
    loading.value = true;
    try {
      localList.value = await props.list();
    } finally {
      loading.value = false;
    }
  } else {
    localList.value = props.list ?? [];
  }
});

const handleRemoteMethod = async (keyword: string) => {
  loading.value = true;
  try {
    currentGenerator = props.listGenerator!(keyword);
    await loadFirstPage(currentGenerator);
  } finally {
    loading.value = false;
  }
};

const handleScrollEnd = async () => {
  if (!currentGenerator || scrollLoading.value) return;
  scrollLoading.value = true;
  try {
    const result = await currentGenerator.next();
    if (!result.done) {
      localList.value = [...localList.value, ...(result.value as Record<string, any>[])];
    }
  } finally {
    scrollLoading.value = false;
  }
};

const handleChange = (value: string | number) => {
  const idKey = (attrs.idKey as string) || 'id';
  const item = localList.value.find((it) => it[idKey] === value);
  emit('change', value, item);
};

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
    :filterable="isGeneratorMode || undefined"
    :scroll-loading="isGeneratorMode ? scrollLoading : undefined"
    :remote-method="isGeneratorMode ? handleRemoteMethod : undefined"
    v-bind="attrs"
    @change="handleChange"
    @scroll-end="handleScrollEnd"
  />
</template>
