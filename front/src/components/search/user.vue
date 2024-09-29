<script setup lang="ts">
import { computed } from 'vue';
import MemberSelect from '@/components/MemberSelect';

defineOptions({ name: 'hcm-search-user' });

const props = withDefaults(defineProps<{ multiple: boolean }>(), {
  multiple: true,
});

const model = defineModel<string[]>();

const defaultUserlist = computed(() => localModel.value.map((item) => ({ username: item, display_name: item })));

const localModel = computed({
  get() {
    if (props.multiple && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(val) {
    model.value = val;
  },
});
</script>

<template>
  <MemberSelect v-model="localModel" :default-userlist="defaultUserlist" clearable />
</template>
