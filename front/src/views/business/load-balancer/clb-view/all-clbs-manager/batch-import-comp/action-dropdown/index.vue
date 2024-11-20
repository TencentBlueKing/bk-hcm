<script setup lang="ts">
import { ref } from 'vue';
import { Dropdown, Button } from 'bkui-vue';
import PermissionDialog from '@/components/permission-dialog';

import { useI18n } from 'vue-i18n';
import { useVerify } from '@/hooks';
import { Action } from '../types';

const { DropdownMenu, DropdownItem } = Dropdown;

defineOptions({ name: 'ActionDropdown' });
const emit = defineEmits<(e: 'click', action: Action) => void>();

const { t } = useI18n();

// 权限校验
const {
  showPermissionDialog,
  handlePermissionConfirm,
  handlePermissionDialog,
  handleAuth,
  permissionParams,
  authVerifyData,
} = useVerify();

const isShow = ref(false);
const actionList = ref([
  { action: Action.CREATE_LISTENER_OR_URL_RULE, label: t('批量创建监听器及规则') },
  { action: Action.BIND_RS, label: t('批量绑定RS') },
]);

const handleShowDropdownMenu = () => {
  if (!authVerifyData.value?.permissionAction?.load_balancer_update) {
    handleAuth('clb_resource_operate');
  } else {
    isShow.value = true;
  }
};

const handleClick = (action: Action) => {
  isShow.value = false;
  emit('click', action);
};
</script>

<template>
  <Dropdown
    placement="bottom-start"
    trigger="manual"
    :is-show="isShow"
    :popover-options="{ forceClickoutside: true }"
    @hide="isShow = false"
  >
    <Button
      :class="{ 'hcm-no-permision-btn': !authVerifyData?.permissionAction?.load_balancer_update }"
      @click="handleShowDropdownMenu"
    >
      {{ t('批量导入') }}
    </Button>
    <template #content>
      <DropdownMenu>
        <DropdownItem v-for="{ action, label } in actionList" :key="action" @click="handleClick(action)">
          {{ label }}
        </DropdownItem>
      </DropdownMenu>
    </template>
  </Dropdown>

  <PermissionDialog
    v-model:is-show="showPermissionDialog"
    :params="permissionParams"
    @cancel="handlePermissionDialog"
    @confirm="handlePermissionConfirm"
  />
</template>

<style scoped></style>
