<script lang="ts" setup>
import { computed, ref, watchEffect, defineExpose, PropType, reactive, watch } from 'vue';
import { useAccountStore } from '@/store';
import { SelectColumn } from '@blueking/ediatable';
import { useRouter, useRoute } from 'vue-router';
import { isEmpty, localStorageActions } from '@/common/util';
import { useI18n } from 'vue-i18n';

const props = defineProps({
  modelValue: [Number, String, Array] as PropType<number | string | Array<number | string>>,
  authed: Boolean as PropType<boolean>,
  autoSelect: Boolean as PropType<boolean>,
  isAudit: Boolean as PropType<boolean>,
  isEditable: Boolean as PropType<boolean>,
  multiple: Boolean as PropType<boolean>,
  clearable: Boolean as PropType<boolean>,
  isShowAll: Boolean as PropType<boolean>,
  notAutoSelectAll: Boolean as PropType<boolean>,
  saveBizs: Boolean as PropType<boolean>,
  bizsKey: String as PropType<string>,
  apiMethod: Function as PropType<(...args: any) => Promise<any>>,
});
const emit = defineEmits(['update:modelValue']);

const router = useRouter();
const route = useRoute();
const { t } = useI18n();
const accountStore = useAccountStore();
const businessList = ref([]);
const defaultBusiness = ref();
const loading = ref(null);

watchEffect(async () => {
  loading.value = true;
  let req = props.authed ? accountStore.getBizListWithAuth : accountStore.getBizList;
  req = props.apiMethod || req;
  if (props.isAudit) {
    req = accountStore.getBizAuditListWithAuth;
  }

  const res = await req();
  loading.value = false;
  businessList.value = res?.data;

  // 支持全选
  if (props.isShowAll) {
    businessList.value.unshift({ name: t('全部'), id: 'all' });
  }

  let id = null;
  // 自动选中的默认值
  if (props.autoSelect && !isEmpty(businessList.value)) {
    // 支持全选, 若开启 notAutoSelectAll 则选中第一个业务; 若未开启, 则选中"全选"
    id =
      props.isShowAll && props.notAutoSelectAll && businessList.value[1]
        ? businessList.value[1].id
        : businessList.value[0]?.id;
  }

  // 处理多选情况，注意默认值可能为 null，此时需要转为空数组
  if (props.multiple) {
    id = id ? [id] : [];
  }

  // 开启 saveBizs, 则自动选中上一次选中的业务
  if (props.saveBizs) {
    const urlBizs = route.query[props.bizsKey] as string;

    // 优先使用 url 中的业务id, 其次是持久化的, 最后是默认值
    id = urlBizs
      ? JSON.parse(atob(urlBizs))
      : localStorageActions.get(props.bizsKey, (value) => JSON.parse(atob(value))) || id;
  }

  // 设置选中的值
  defaultBusiness.value = id;
  selectedValue.value = id;
});

const computedBuinessList = computed(() => {
  return businessList.value.map(({ name, id }) => ({
    value: id,
    label: name,
  }));
});

const selectedValue = computed({
  get() {
    if (!isEmpty(props.modelValue)) {
      return props.modelValue;
    }
    if (props.isShowAll) {
      if (props.multiple && Array.isArray(props.modelValue) && props.modelValue.length === 0) {
        return ['all'];
      }
      if (!props.multiple && props.modelValue === '') {
        return 'all';
      }
    }
    return props.multiple ? [] : null;
  },
  set(val) {
    let selectedValue = val;
    if (props.isShowAll) {
      if (props.multiple && Array.isArray(selectedValue)) {
        if (selectedValue[selectedValue.length - 1] === 'all') {
          selectedValue = [];
        } else if (selectedValue.includes('all')) {
          const index = selectedValue.findIndex((val) => val === 'all');
          selectedValue.splice(index, 1);
        }
      }
      if (!props.multiple && selectedValue === 'all') {
        selectedValue = '';
      }
    }
    emit('update:modelValue', selectedValue);
  },
});

// 记录业务id
watch(
  selectedValue,
  (val) => {
    if (!props.saveBizs) return;

    const query = { ...route.query };
    const encodedBizs = btoa(JSON.stringify(val));

    // 多选
    if (props.multiple) {
      // 未选时, 不用存业务id
      query[props.bizsKey] = val.length > 0 ? encodedBizs : undefined;
    }
    // 单选
    else {
      query[props.bizsKey] = encodedBizs || undefined;
    }

    // 持久化处理
    if (query[props.bizsKey]) {
      localStorageActions.set(props.bizsKey, query[props.bizsKey]);
    } else {
      localStorageActions.remove(props.bizsKey);
    }
    // 记录业务id 到 url 上
    router.push({ query });
  },
  { deep: true },
);

const rules = reactive([
  {
    validator: (value: string) => Boolean(value),
    message: '请选择业务',
  },
]);

defineExpose({
  businessList,
  defaultBusiness,
  rules,
});
</script>

<template>
  <select-column
    v-model="selectedValue"
    v-if="isEditable"
    :list="computedBuinessList"
    filterable
    :loading="loading"
    :rules="rules"
  ></select-column>

  <bk-select v-else v-model="selectedValue" :multiple="multiple" filterable :loading="loading" :clearable="clearable">
    <bk-option v-for="item in businessList" :key="item.id" :value="item.id" :label="item.name" />
  </bk-select>
</template>
