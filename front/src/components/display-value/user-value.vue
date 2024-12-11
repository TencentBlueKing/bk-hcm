<script setup lang="ts">
import { computed, watchEffect } from 'vue';
import CombineRequest from '@blueking/combine-request';
import { useUserStore } from '@/store/user';

const props = defineProps<{ value: string | string[] }>();

const localValue = computed(() => {
  if (!props.value) {
    return [];
  }
  return Array.isArray(props.value) ? props.value : [props.value];
});

const displayValue = computed(() => {
  const names = localValue.value.map((username) => {
    // 每次从全局store中查询获取
    const user = userStore.userList.find((user) => user.username === username);
    if (!user) {
      return '--';
    }
    return `${user.username}(${user.display_name})`;
  });
  return names?.join?.(';');
});

const userStore = useUserStore();

const combineRequest = CombineRequest.setup(Symbol.for('user-value'), (users) => {
  const uniqueUsers = [...new Set((users as string[][]).reduce((acc, cur) => acc.concat(cur), []))];
  userStore.getUserByName(uniqueUsers);
});

watchEffect(() => {
  if (!localValue.value.length) {
    return;
  }

  const newUsers = localValue.value.filter(
    (username) => !userStore.userList.some((item) => item.username === username),
  );

  if (newUsers.length) {
    combineRequest.add(newUsers);
  }
});
</script>

<template>
  {{ displayValue }}
</template>
