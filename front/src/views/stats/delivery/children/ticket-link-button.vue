<script setup lang="ts">
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_TICKET_MANAGEMENT } from '@/constants/menu-symbol';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import useSearchQs from '@/hooks/use-search-qs';

const props = defineProps<{
  bizId: number;
  text?: string;
  filter?: Record<string, any>;
  type?: 'host_apply';
}>();
const handleClick = () => {
  const searchQs = useSearchQs();
  routerAction.open({
    name: MENU_BUSINESS_TICKET_MANAGEMENT,
    query: {
      [GLOBAL_BIZS_KEY]: props.bizId,
      type: props.type || 'host_apply',
      filter: props.filter ? searchQs.build(props.filter) : undefined,
    },
  });
};
</script>

<template>
  <bk-button text theme="primary" @click="handleClick">
    <slot>{{ text }}</slot>
  </bk-button>
</template>
