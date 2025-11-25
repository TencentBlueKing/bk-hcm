<template>
  <Button class="mb10" theme="primary" outline :disabled="!vendor" @click="showConfigureSlider">
    <Plus />
    {{ t('添加') }}
  </Button>
  <bk-table
    class="config-list-table"
    :data="list"
    :columns="ruleColumns"
    :max-height="300"
    row-hover="auto"
    :stripe="true"
    show-overflow-tooltip
    :settings="settings"
  />
</template>

<script setup lang="ts">
import { h } from 'vue';
import { Button } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import { useI18n } from 'vue-i18n';
import type { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';

const props = withDefaults(defineProps<IConfigurationListProps>(), {
  list: () => [],
  vendor: '',
});

const emit = defineEmits(['showConfigureSlider', 'cloneData', 'removeData']);

const { t } = useI18n();

export interface IConfigurationListProps {
  list: ApplyClbModel[];
  vendor: string;
}

const { columns: ruleColumns, settings } = useColumns('configureList', false, props.vendor);
ruleColumns.push({
  label: '操作',
  width: 120,
  showOverflowTooltip: false,
  render: ({ data }: { data: any }) => {
    return h('div', { class: 'operation-column' }, [
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          class: 'mr10',
          onClick: () => handleClone(data),
        },
        '克隆',
      ),
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          class: 'mr10',
          onClick: () => handleRemove(data.rowKey),
        },
        '移除',
      ),
    ]);
  },
});
const showConfigureSlider = () => {
  emit('showConfigureSlider');
};

const handleClone = (data: ApplyClbModel) => {
  emit('cloneData', data);
};

const handleRemove = (key: string) => {
  emit('removeData', key);
};
</script>
