<script setup lang="ts">
import { computed, watch } from 'vue';
import { useUserStore } from '@/store';
import MemberSelect from '@/components/MemberSelect';

defineOptions({ name: 'hcm-search-user' });

const props = withDefaults(defineProps<{ multiple: boolean; defaultCurrent: boolean }>(), {
  multiple: true,
  defaultCurrent: true,
});

const userStore = useUserStore();
const model = defineModel<string[]>();

const defaultUserlist = computed(() => [
  {
    username: userStore.username,
    display_name: userStore.username,
  },
]);

watch(
  () => userStore.username,
  (val) => {
    if (props.defaultCurrent) {
      model.value = [val];
    }
  },
  { immediate: true },
);
</script>

<template>
  <MemberSelect v-model="model" :default-userlist="defaultUserlist" clearable />
</template>
