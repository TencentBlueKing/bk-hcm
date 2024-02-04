<script lang="ts" setup>
import { computed, watch } from 'vue';
import { AuditActionEnum, AuditActionNameEnum } from '../constants';

const props = defineProps({
  type: String,
});

const emit = defineEmits(['input']);

const typeActions = {
  account: [
    { id: AuditActionEnum.CREATE, name: AuditActionNameEnum.CREATE },
    { id: AuditActionEnum.UPDATE, name: AuditActionNameEnum.UPDATE },
    { id: AuditActionEnum.DELETE, name: AuditActionNameEnum.DELETE },
  ],
  cvm: [
    { id: AuditActionEnum.CREATE, name: AuditActionNameEnum.CREATE },
    { id: AuditActionEnum.ASSIGN, name: AuditActionNameEnum.ASSIGN },
    { id: AuditActionEnum.UPDATE, name: AuditActionNameEnum.UPDATE },
    { id: AuditActionEnum.REBOOT, name: AuditActionNameEnum.REBOOT },
    { id: AuditActionEnum.START, name: AuditActionNameEnum.START },
    { id: AuditActionEnum.STOP, name: AuditActionNameEnum.STOP },
    { id: AuditActionEnum.RESET_PWD, name: AuditActionNameEnum.RESET_PWD },
    { id: AuditActionEnum.DELETE, name: AuditActionNameEnum.DELETE },
  ],
  vpc: [
    { id: AuditActionEnum.CREATE, name: AuditActionNameEnum.CREATE },
    { id: AuditActionEnum.APPLY, name: AuditActionNameEnum.APPLY },
    { id: AuditActionEnum.ASSIGN, name: AuditActionNameEnum.ASSIGN },
    { id: AuditActionEnum.UPDATE, name: AuditActionNameEnum.UPDATE },
    { id: AuditActionEnum.DELETE, name: AuditActionNameEnum.DELETE },
    { id: AuditActionEnum.BIND, name: AuditActionNameEnum.BIND },
  ],
  security_group: [
    { id: AuditActionEnum.CREATE, name: AuditActionNameEnum.CREATE },
    { id: AuditActionEnum.UPDATE, name: AuditActionNameEnum.UPDATE },
    { id: AuditActionEnum.DELETE, name: AuditActionNameEnum.DELETE },
  ],
  disk: [
    { id: AuditActionEnum.MOUNT, name: AuditActionNameEnum.MOUNT },
    { id: AuditActionEnum.UNMOUNT, name: AuditActionNameEnum.UNMOUNT },
    { id: AuditActionEnum.RECYCLE, name: AuditActionNameEnum.RECYCLE },
  ],
  gcp_firewall_rule: [
    { id: AuditActionEnum.CREATE, name: AuditActionNameEnum.CREATE },
    { id: AuditActionEnum.UPDATE, name: AuditActionNameEnum.UPDATE },
    { id: AuditActionEnum.DELETE, name: AuditActionNameEnum.DELETE },
  ],
  route_table: [
    { id: AuditActionEnum.CREATE, name: AuditActionNameEnum.CREATE },
    { id: AuditActionEnum.UPDATE, name: AuditActionNameEnum.UPDATE },
    { id: AuditActionEnum.DELETE, name: AuditActionNameEnum.DELETE },
  ],
  image: [
    { id: AuditActionEnum.CREATE, name: AuditActionNameEnum.CREATE },
    { id: AuditActionEnum.DELETE, name: AuditActionNameEnum.DELETE },
  ],
  network_interface: [
    { id: AuditActionEnum.CREATE, name: AuditActionNameEnum.CREATE },
    { id: AuditActionEnum.UPDATE, name: AuditActionNameEnum.UPDATE },
    { id: AuditActionEnum.DELETE, name: AuditActionNameEnum.DELETE },
  ],
  subnet: [
    { id: AuditActionEnum.CREATE, name: AuditActionNameEnum.CREATE },
    { id: AuditActionEnum.UPDATE, name: AuditActionNameEnum.UPDATE },
    { id: AuditActionEnum.DELETE, name: AuditActionNameEnum.DELETE },
    { id: AuditActionEnum.ASSIGN, name: AuditActionNameEnum.ASSIGN },
  ],
};

const actions = computed(() => {
  return typeActions[props.type] || [];
});

const selectedValue = computed({
  get() {
    return actions.value?.[0]?.id || '';
  },
  set(val) {
    emit('input', val);
  },
});

watch(selectedValue, (v) => console.log(v));
</script>

<template>
  <bk-select :key="props.type" v-model="selectedValue" filterable :multiple="false">
    <bk-option v-for="(item, index) in actions" :key="index" :value="item.id" :label="item.name" />
  </bk-select>
</template>
