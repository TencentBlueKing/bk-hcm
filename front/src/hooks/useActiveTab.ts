import { ref, watchEffect } from 'vue';
import { useRouter, useRoute } from 'vue-router';

/**
 * Tab组件切换Panel, url中query参数(type)为标识
 */
export default (initialValue: string) => {
  const router = useRouter();
  const route = useRoute();
  const activeTab = ref(initialValue);

  // change-handler
  const handleActiveTabChange = (v: string) => {
    // tab切换
    activeTab.value = v;
    // 路由切换
    router.push({ query: { ...route.query, type: v } });
  };

  // 监听route.query.type的变化, tab状态保持
  watchEffect(() => {
    handleActiveTabChange(route.query.type as string);
  });

  return {
    activeTab,
    handleActiveTabChange,
  };
};
