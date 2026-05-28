<script setup lang="ts">
import { computed, h, ref, watchEffect, useId, nextTick, useAttrs } from 'vue';
import debounce from 'lodash/debounce';
import BkUserSelector from '@blueking/bk-user-selector';
import '@blueking/bk-user-selector/vue3/vue3.css';
import { TagInputColumn } from '@blueking/ediatable';
import { useUserStore, type IUserItem } from '@/store/user';
import { userSelectorRecentSelectedKey } from '@/constants/storage-symbols';
import type { DisplayType } from '@/components/form/typings';
import type { Rules } from '@blueking/ediatable';

defineOptions({ name: 'user-selector' });

const model = defineModel<string | string[]>();

const props = withDefaults(defineProps<IUserSelectorProps>(), {
  multiple: true,
  allowCreate: true,
  clearable: true,
  placeholder: '请输入',
  fastSelect: true,
  hasDeleteIcon: true,
  trigger: 'focus',
  collapseTags: true,
});

const emit = defineEmits<{
  change: [val: string | string[]];
}>();

export interface IUserSelectorProps {
  multiple?: boolean;
  disabled?: boolean;
  clearable?: boolean;
  placeholder?: string;
  fastSelect?: boolean;
  allowCreate?: boolean;
  hasDeleteIcon?: boolean;
  trigger?: 'focus' | 'search';
  collapseTags?: boolean;
  display?: DisplayType;
  rules?: Rules;
  copyable?: boolean;
}

const attrs = useAttrs();
const userStore = useUserStore();

// ====== cell 模式（TagInputColumn）专用逻辑 ======
const isCellMode = computed(() => props.display?.on === 'cell');
const comp = computed(() => (isCellMode.value ? TagInputColumn : null));

const id = useId();
const activeSearchId = ref<string | null>(null);

// 记住最近选择过的10个用户
const getRecent = () => JSON.parse(localStorage.getItem(userSelectorRecentSelectedKey)) || [];
const setRecent = (val: string[]) => localStorage.setItem(userSelectorRecentSelectedKey, JSON.stringify(val));
const saveRecent = (val: string[]) => {
  const lastSelected = getRecent();
  setRecent([...new Set([...val, ...lastSelected])].slice(0, 10));
};

const cellLocalModel = computed<string[]>({
  get() {
    if (!model.value) {
      return [];
    }
    if (!Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(val) {
    saveRecent(val);
    if (!props.multiple) {
      [model.value] = val;
    } else {
      model.value = val;
    }
  },
});

const userList = ref<IUserItem[]>([]);

const listTpl = (node: IUserItem, hl: (value: string) => string) => {
  const innerHTML = `${hl(node.username)}${node.display_name ? `(${hl(node.display_name)})` : ''}`;
  return h('div', { class: 'bk-selector-node' }, [
    h('span', {
      class: 'text',
      innerHTML,
    }),
  ]);
};

const tagTpl = (node: IUserItem) => {
  const tagContent = `${node.username}${node.display_name ? `(${node.display_name})` : ''}`;
  return h('div', { class: 'tag' }, [
    h('span', {
      class: 'text',
      innerHTML: tagContent,
    }),
  ]);
};

watchEffect(async () => {
  if (!isCellMode.value) return;

  const defaultUsers = [...new Set([...cellLocalModel.value, ...getRecent(), userStore.username])];
  const newUsers = defaultUsers.filter(
    (username: string) => !userList.value.some((oldItem) => oldItem.username === username),
  );

  if (newUsers.length) {
    const searchUsers: string[] = [];
    const existUserList: IUserItem[] = [];
    newUsers.forEach((username: string) => {
      const user = userStore.userList.find((oldItem) => oldItem.username === username);
      if (user) {
        existUserList.push(user);
      } else {
        searchUsers.push(username);
      }
    });

    let newUserList: IUserItem[] = [];
    if (searchUsers.length) {
      newUserList = await userStore.getUserByName(searchUsers);
      if (!newUserList.length) {
        newUserList = searchUsers.map((username) => ({ username, display_name: username }));
      }
    }

    const totalUserList = [...userList.value, ...existUserList, ...newUserList];
    const uniqueUserList = totalUserList.reduce((acc, cur) => {
      if (!acc.some((item) => item.username === cur.username)) {
        acc.push(cur);
      }
      return acc;
    }, []);

    userList.value = uniqueUserList;
  }
});

const handleInput = debounce(async (inputValue: string) => {
  const value = inputValue.toLowerCase().trim();
  if (!value) {
    return;
  }

  if (!props.multiple && userList.value.some((item) => item.username === value)) {
    return;
  }

  activeSearchId.value = id;

  const list = await userStore.search(value);
  const newList = list.filter((item) => !userList.value.some((oldItem) => oldItem.username === item.username));
  userList.value = [...userList.value, ...newList];

  activeSearchId.value = null;
}, 500);

const handleSelect = () => {
  if (!props.multiple) {
    tagInputRef.value?.handleBlur();
  }
};

const handleClickMe = () => {
  if (props.multiple) {
    if (!cellLocalModel.value.includes(userStore.username)) {
      cellLocalModel.value = [...cellLocalModel.value, userStore.username];
    }
  } else {
    cellLocalModel.value = [userStore.username];
  }

  if (isCellMode.value) {
    nextTick(() => {
      tagInputRef.value?.getValue();
    });
  } else {
    tagInputRef.value?.handleBlur();
  }
};

// ====== 默认模式（bk-user-selector）逻辑 ======
const tenantId = computed(() => userStore.tenantId);
const currentUserId = computed(() => props.fastSelect && userStore.username);
const apiBaseUrl = window.PROJECT_CONFIG.USER_MANAGE_URL;

const tagInputRef = ref();

const handleChange = (val: string | string[]) => {
  emit('change', val);
};

const focus = () => {
  tagInputRef.value?.focusInputTrigger?.();
};

defineExpose({
  getValue() {
    if (tagInputRef.value?.getValue) {
      return tagInputRef.value.getValue().then(() => model.value);
    }
    return model.value;
  },
  focus,
});
</script>

<template>
  <!-- cell 模式：使用 TagInputColumn，适配可编辑表格 -->
  <div v-if="isCellMode" class="user-selector-wrap is-cell">
    <component
      :is="comp"
      class="user-selector"
      v-model="cellLocalModel"
      ref="tagInputRef"
      :list="userList"
      :tpl="listTpl"
      :tag-tpl="tagTpl"
      :max-data="multiple ? -1 : 1"
      :allow-next-focus="multiple"
      :allow-auto-match="!multiple"
      :disabled="disabled"
      :clearable="clearable"
      :allow-create="allowCreate"
      :has-delete-icon="hasDeleteIcon"
      :trigger="trigger"
      :collapse-tags="collapseTags"
      :placeholder="placeholder"
      :show-clear-only-hover="true"
      :is-async-list="true"
      :display-key="'display_name'"
      :save-key="'username'"
      :search-key="['username', 'display_name']"
      :rules="rules"
      :copyable="false"
      v-bind="attrs"
      @input="handleInput"
      @select="handleSelect"
    />
    <div class="suffix-absolute">
      <div class="me" v-show="!(activeSearchId === id && userStore.searchLoading)" @click.stop="handleClickMe">我</div>
      <div class="loading" v-show="activeSearchId === id && userStore.searchLoading">
        <bk-loading :loading="userStore.searchLoading" mode="spin" size="mini" />
      </div>
    </div>
  </div>

  <!-- 默认模式：使用 bk-user-selector -->
  <bk-user-selector
    v-else
    class="user-selector"
    ref="tagInputRef"
    v-model="model"
    :multiple="multiple"
    :placeholder="placeholder"
    :tenant-id="tenantId"
    :current-user-id="currentUserId"
    :api-base-url="apiBaseUrl"
    :disabled="disabled"
    :clearable="clearable"
    v-bind="attrs"
    @change="handleChange"
  />
</template>

<style lang="scss" scoped>
.user-selector-wrap {
  position: relative;
  width: 100%;

  &.is-cell {
    :deep(.bk-tag-input-trigger .tag-list) {
      height: auto;
    }

    .suffix-absolute {
      position: absolute;
      right: 26px;
      top: 50%;
      transform: translateY(-50%);
      display: flex;
      align-items: center;
      z-index: 1;

      .me {
        color: $default-color;
        cursor: pointer;
      }

      .loading {
        transform: scale(0.75);
      }
    }
  }
}
</style>
