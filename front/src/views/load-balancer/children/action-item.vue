<script setup lang="ts">
import { ref } from 'vue';
import type { ActionItemType } from '../typing';

import { AngleDown } from 'bkui-vue/lib/icon';

interface IProps {
  action: ActionItemType;
  disabled?: boolean;
}

defineProps<IProps>();

const isDropdownMenuShow = ref(false);
const handleDropdownItemClick = (action: ActionItemType) => {
  isDropdownMenuShow.value = false;
  action.handleClick();
};
</script>

<template>
  <!-- custom-render -->
  <template v-if="action.render">
    <component :is="action.render()" />
  </template>
  <template v-else>
    <!-- dropdown -->
    <bk-dropdown
      v-if="action.type === 'dropdown'"
      :is-show="isDropdownMenuShow"
      v-bind="action.displayProps"
      :disabled="disabled"
      trigger="manual"
      :popover-options="{ forceClickoutside: true }"
    >
      <bk-button :disabled="disabled" @click="isDropdownMenuShow = true">
        {{ action.label }}
        <angle-down class="f26" />
      </bk-button>
      <template #content>
        <bk-dropdown-menu class="dropdown-menu">
          <template v-for="childAction in action.children" :key="childAction.value">
            <hcm-auth v-if="childAction.authSign" :sign="childAction.authSign()" v-slot="{ noPerm }">
              <bk-dropdown-item
                v-bind="childAction.displayProps"
                :disabled="noPerm || childAction.disabled?.()"
                @click="handleDropdownItemClick(childAction)"
              >
                {{ childAction.label }}
              </bk-dropdown-item>
            </hcm-auth>
            <bk-dropdown-item
              v-else
              v-bind="childAction.displayProps"
              :disabled="childAction.disabled?.()"
              @click="handleDropdownItemClick(childAction)"
            >
              {{ childAction.label }}
            </bk-dropdown-item>
          </template>
        </bk-dropdown-menu>
      </template>
    </bk-dropdown>
    <!-- button -->
    <bk-button v-else class="button" v-bind="action.displayProps" :disabled="disabled" @click="action.handleClick()">
      <component v-if="action.prefix" :is="action.prefix()" class="f26" />
      {{ action.label }}
    </bk-button>
  </template>
</template>

<style scoped lang="scss">
.dropdown-menu {
  display: flex;
  flex-direction: column;

  :deep(.bk-dropdown-item) {
    width: 100%;
  }
}

.button {
  min-width: 64px;
}

.f26 {
  font-size: 26px;
}
</style>
