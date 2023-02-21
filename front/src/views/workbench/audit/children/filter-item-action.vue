<script lang="ts" setup>
import { computed } from 'vue';
import { AuditActionEnum, AuditActionNameEnum } from '../constants';

const props = defineProps({
  type: String,
});

const emit = defineEmits(['input']);

const baseActions = [
  { id: AuditActionEnum.CREATE, name: AuditActionNameEnum.CREATE },
  { id: AuditActionEnum.UPDATE, name: AuditActionNameEnum.UPDATE },
  { id: AuditActionEnum.DELETE, name: AuditActionNameEnum.DELETE },
];
const typeActions = {
  gcp_firewall_rule: [
    { id: AuditActionEnum.RESTART, name: AuditActionNameEnum.RESTART },
  ],
};

const actions = computed(() => {
  return [...baseActions, ...(typeActions[props.type] || [])];
});

const selectedValue = computed({
  get() {
    return actions.value?.[0]?.id;
  },
  set(val) {
    emit('input', val);
  },
});
</script>

<template>
  <bk-select
    v-model="selectedValue"
    filterable
    :multiple="false"
  >
    <bk-option
      v-for="(item, index) in actions"
      :key="index"
      :value="item.id"
      :label="item.name"
    />
  </bk-select>
</template>
