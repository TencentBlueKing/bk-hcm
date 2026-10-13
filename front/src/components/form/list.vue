<script setup lang="ts">
import { computed, ref, useAttrs, useSlots, watchEffect } from 'vue';
import { ModelProperty } from '@/model/typings';
import { SelectColumn } from '@blueking/ediatable';
import { DisplayType } from './typings';

export type ListGeneratorFactory<T = Record<string, any>> = (
  keywordOrOptions?: string | { ids?: (string | number)[]; [key: string]: any },
) => AsyncGenerator<T[], void>;

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
const slots = useSlots();

const comp = computed(() => (props.display?.on === 'cell' ? SelectColumn : 'bk-select'));

const isGeneratorMode = computed(() => !!props.listGenerator);
const idKey = computed(() => (attrs.idKey as string) || 'id');
const localList = ref<Array<Record<string, any>>>([]);
const loading = ref(false);
const scrollLoading = ref(false);
const pinnedIds = ref<Set<string | number>>(new Set());
let currentGenerator: AsyncGenerator<Record<string, any>[], void> | null = null;

const filterPinned = (items: Record<string, any>[]) => items.filter((item) => !pinnedIds.value.has(item[idKey.value]));

const loadFirstPage = async (generator: AsyncGenerator<Record<string, any>[], void>) => {
  const result = await generator.next();
  localList.value = result.done ? [] : (result.value as Record<string, any>[]);
};
const appendSelectedItems = async () => {
  if (!model.value) return;
  const ids = (Array.isArray(model.value) ? model.value : [model.value]).filter(
    (id) => !localList.value.some((item) => item[idKey.value] === id),
  );
  if (ids.length === 0) return;
  const dataGen = props.listGenerator?.({ ids });
  const dataList: Record<string, any>[] = [];
  for await (const items of dataGen) {
    dataList.push(...items);
  }
  if (dataList.length > 0) {
    localList.value = [...dataList, ...localList.value];
    dataList.forEach((item) => pinnedIds.value.add(item[idKey.value]));
  }
};

watchEffect(async () => {
  if (isGeneratorMode.value) {
    loading.value = true;
    pinnedIds.value = new Set();
    try {
      currentGenerator = props.listGenerator?.();
      await loadFirstPage(currentGenerator);
      await appendSelectedItems();
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
  pinnedIds.value = new Set();
  try {
    currentGenerator = props.listGenerator?.(keyword);
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
      localList.value = [...localList.value, ...filterPinned(result.value as Record<string, any>[])];
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
  >
    <template v-for="(slotFn, slotName) in slots" :key="slotName" #[slotName]="slotProps">
      <component :is="slotFn" v-if="slotFn" v-bind="slotProps" />
    </template>
  </component>
</template>

<style lang="scss" scoped>
.bk-ediatable-select {
  :deep(.bk-select-tag) {
    background-color: transparent;
  }
}
</style>
