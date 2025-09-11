<script setup lang="ts">
import { ref } from 'vue';
import type { ActionItemType } from '../typing';

import { AngleDown } from 'bkui-vue/lib/icon';

interface IProps {
  action: ActionItemType;
  disabled?: boolean;
  loading?: boolean;
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
      @hide="isDropdownMenuShow = false"
    >
      <bk-button :disabled="disabled" @click="isDropdownMenuShow = true">
        {{ action.label }}
        <angle-down class="f26" />
      </bk-button>
      <template #content>
        <bk-dropdown-menu class="dropdown-menu" @click="isDropdownMenuShow = false">
          <template v-for="childAction in action.children" :key="childAction.value">
            <bk-dropdown-item v-if="childAction.render" class="dropdown-item">
              <component :is="childAction.render()" :is-in-dropdown="true" @click="isDropdownMenuShow = false" />
            </bk-dropdown-item>
            <hcm-auth v-else-if="childAction.authSign" :sign="childAction.authSign()" v-slot="{ noPerm }">
              <bk-dropdown-item class="dropdown-item">
                <bk-button
                  class="button"
                  text
                  :disabled="noPerm || childAction.disabled?.()"
                  @click="handleDropdownItemClick(childAction)"
                  v-bind="childAction.displayProps"
                >
                  {{ childAction.label }}
                </bk-button>
              </bk-dropdown-item>
            </hcm-auth>
            <bk-dropdown-item v-else class="dropdown-item">
              <bk-button
                class="button"
                text
                :disabled="childAction.disabled?.()"
                @click="handleDropdownItemClick(childAction)"
                v-bind="childAction.displayProps"
              >
                {{ childAction.label }}
              </bk-button>
            </bk-dropdown-item>
          </template>
        </bk-dropdown-menu>
      </template>
    </bk-dropdown>
    <!-- button -->
    <bk-button
      v-else
      class="button"
      v-bind="action.displayProps"
      :disabled="disabled"
      :loading="loading"
      @click="action.handleClick()"
    >
      <component v-if="action.prefix" :is="action.prefix()" class="f26" />
      {{ action.label }}
    </bk-button>
  </template>
</template>

<style scoped lang="scss">
.button {
  min-width: 64px;
}

.dropdown-menu {
  display: flex;
  flex-direction: column;

  .dropdown-item {
    width: 100%;
    padding: 0;

    .button {
      padding: 0 16px;
      display: inline-block;
      width: 100%;
      text-align: left;
    }
  }
}

.f26 {
  font-size: 26px;
}
</style>
