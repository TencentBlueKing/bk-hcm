<script setup lang="ts">
import { computed, watchEffect } from 'vue';
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
    const user = userStore.userList.find((user) => user.username === username);
    if (!user) {
      return '--';
    }
    return `${user.username}(${user.display_name})`;
  });
  return names?.join?.(';');
});

const userStore = useUserStore();

watchEffect(() => {
  if (!localValue.value.length) {
    return;
  }

  const newUsers = localValue.value.filter(
    (username) => !userStore.userList.some((item) => item.username === username),
  );

  // TODO: 合并请求
  if (newUsers.length) {
    userStore.getUserByName(newUsers);
  }
});
</script>

<template>
  {{ displayValue }}
</template>
