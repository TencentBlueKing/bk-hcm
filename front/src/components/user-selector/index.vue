<script setup lang="ts">
import { computed, h, ref, watchEffect } from 'vue';
import debounce from 'lodash/debounce';
import { useUserStore, type IUserItem } from '@/store/user';
import { userSelectorRecentSelectedKey } from '@/constants/storage-symbols';

defineOptions({ name: 'user-selector' });

export interface IUserSelectorProps {
  multiple?: boolean;
  disabled?: boolean;
  clearable?: boolean;
  allowCreate?: boolean;
  hasDeleteIcon?: boolean;
  trigger?: 'focus' | 'search';
  collapseTags?: boolean;
  placeholder?: string;
}

const props = withDefaults(defineProps<IUserSelectorProps>(), {
  multiple: true,
  allowCreate: true,
  hasDeleteIcon: true,
  clearable: true,
  trigger: 'focus',
  collapseTags: true,
  placeholder: '请输入',
});

const model = defineModel<string | string[]>();

const userStore = useUserStore();

// 记住最近选择过的10个用户
const getRecent = () => JSON.parse(localStorage.getItem(userSelectorRecentSelectedKey)) || [];
const setRecent = (val: string[]) => localStorage.setItem(userSelectorRecentSelectedKey, JSON.stringify(val));
const saveRecent = (val: string[]) => {
  const lastSelected = getRecent();
  setRecent([...new Set([...val, ...lastSelected])].slice(0, 10));
};

const localModel = computed<string[]>({
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
    // 更新最近选择
    saveRecent(val);

    // 如果是单选返回单个值
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
  // 获取默认数据带display_name的完整用户信息
  const defaultUsers = [...new Set([...localModel.value, ...getRecent(), userStore.username])];
  const newUsers = defaultUsers.filter(
    (username: string) => !userList.value.some((oldItem) => oldItem.username === username),
  );

  // 只获取不存在列表中的新用户
  if (newUsers.length) {
    // 在全局store中查询，存在则直接使用，不存在则请求
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
      // allowCreate为true时，允许输入不存在的用户，此时查询结果为空，为了防止重复请求需要创建数据放入到列表中
      if (!newUserList.length) {
        newUserList = searchUsers.map((username) => ({ username, display_name: username }));
      }
    }

    userList.value = [...userList.value, ...existUserList, ...newUserList];
  }
});

const handleInput = debounce(async (inputValue: string) => {
  const value = inputValue.toLowerCase().trim();
  if (!value) {
    return;
  }

  // 如果是单选，且当前输入值已存在，不再获取
  if (!props.multiple && userList.value.some((item) => item.username === value)) {
    return;
  }

  const list = await userStore.search(value);
  const newList = list.filter((item) => !userList.value.some((oldItem) => oldItem.username === item.username));
  userList.value = [...userList.value, ...newList];
}, 500);

const handleClickMe = () => {
  if (props.multiple) {
    if (!localModel.value.includes(userStore.username)) {
      localModel.value = [...localModel.value, userStore.username];
    }
  } else {
    localModel.value = [userStore.username];
  }
};
</script>

<template>
  <bk-tag-input
    class="user-selector"
    v-model="localModel"
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
    @input="handleInput"
  >
    <template #suffix>
      <div class="suffix">
        <div class="me" v-show="!userStore.searchLoading" @click="handleClickMe">我</div>
        <div class="loading" v-show="userStore.searchLoading">
          <bk-loading :loading="userStore.searchLoading" mode="spin" size="mini" />
        </div>
      </div>
    </template>
  </bk-tag-input>
</template>

<style lang="scss" scoped>
.user-selector {
  .suffix {
    margin-left: auto;
    margin-right: 5px;
    display: flex;
    align-items: center;
    .me {
      color: $default-color;
      cursor: pointer;
      z-index: 1;
    }
    .loading {
      transform: scale(0.75);
    }
  }
}
</style>
