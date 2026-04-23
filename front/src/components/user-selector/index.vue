<script setup lang="ts">
import { computed, h, ref, watchEffect, useId, nextTick, useAttrs } from 'vue';
import debounce from 'lodash/debounce';
import { TagInputColumn } from '@blueking/ediatable';
import { useUserStore, type IUserItem } from '@/store/user';
import { userSelectorRecentSelectedKey } from '@/constants/storage-symbols';
import type { DisplayType } from '@/components/form/typings';
import type { Rules } from '@blueking/ediatable';

export interface IUserSelectorProps {
  multiple?: boolean;
  disabled?: boolean;
  clearable?: boolean;
  allowCreate?: boolean;
  hasDeleteIcon?: boolean;
  trigger?: 'focus' | 'search';
  collapseTags?: boolean;
  placeholder?: string;
  display?: DisplayType;
  rules?: Rules;
  copyable?: boolean;
}

defineOptions({ name: 'user-selector' });

const model = defineModel<string | string[]>();
const props = withDefaults(defineProps<IUserSelectorProps>(), {
  multiple: true,
  allowCreate: true,
  hasDeleteIcon: true,
  clearable: true,
  trigger: 'focus',
  collapseTags: true,
  placeholder: '请输入',
});
const attrs = useAttrs();

const comp = computed(() => (props.display?.on === 'cell' ? TagInputColumn : 'bk-tag-input'));

const id = useId();
const activeSearchId = ref<string | null>(null);

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

const tagInputRef = ref();

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

    // 需要再次去重
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

  // 如果是单选，且当前输入值已存在，不再获取
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
  // 临时修复单选时，如果输入框中有值，失焦后不隐藏下拉列表的问题
  if (!props.multiple) {
    tagInputRef.value?.handleBlur();
  }
};

const handleClickMe = () => {
  if (props.multiple) {
    if (!localModel.value.includes(userStore.username)) {
      localModel.value = [...localModel.value, userStore.username];
    }
  } else {
    localModel.value = [userStore.username];
  }

  // 如果是在 cell 模式下 还要触发一次 getValue 校验
  if (props.display?.on === 'cell') {
    nextTick(() => {
      tagInputRef.value?.getValue();
    });
  } else {
    // blur触发强制隐藏，由于组件的实现问题，不隐藏当只有一个"我"选项时会出现一个空白
    tagInputRef.value?.handleBlur(); // FIXME: 这里需要确认一下
  }
};

defineExpose({
  getValue() {
    if (tagInputRef.value?.getValue) {
      return tagInputRef.value.getValue().then(() => model.value);
    }
    return model.value;
  },
});
</script>

<template>
  <div class="user-selector-wrap" :class="{ 'is-cell': display?.on === 'cell' }">
    <component
      :is="comp"
      class="user-selector"
      v-model="localModel"
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
    >
      <template #suffix>
        <div class="suffix">
          <div class="me" v-show="!(activeSearchId === id && userStore.searchLoading)" @click.stop="handleClickMe">
            我
          </div>
          <div class="loading" v-show="activeSearchId === id && userStore.searchLoading">
            <bk-loading :loading="userStore.searchLoading" mode="spin" size="mini" />
          </div>
        </div>
      </template>
    </component>
    <!-- cell 模式：TagInputColumn 不支持 suffix 插槽，用绝对定位模拟 -->
    <div v-if="display?.on === 'cell'" class="suffix-absolute">
      <div class="me" v-show="!(activeSearchId === id && userStore.searchLoading)" @click.stop="handleClickMe">我</div>
      <div class="loading" v-show="activeSearchId === id && userStore.searchLoading">
        <bk-loading :loading="userStore.searchLoading" mode="spin" size="mini" />
      </div>
    </div>
  </div>
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
