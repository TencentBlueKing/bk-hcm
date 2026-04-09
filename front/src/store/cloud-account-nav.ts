import { ref } from 'vue';
import { defineStore } from 'pinia';
import type { TabName } from '@/views/cloud-account-manage/typings';

export interface INavIntent {
  targetTab: TabName;
  detailCloudId?: string;
  filter?: Record<string, any>;
}

export const useCloudAccountNavStore = defineStore('cloudAccountNav', () => {
  /** 当前待消费的导航意图，null 表示无待处理的跨 Tab 跳转 */
  const navIntent = ref<INavIntent | null>(null);

  /**
   * 发起跨 Tab 跳转意图
   * 由父容器的 switchToXxxTab 调用
   */
  const setNavIntent = (intent: INavIntent) => {
    navIntent.value = intent;
  };

  /**
   * 消费导航意图
   * 目标 Tab 在数据就绪后调用，获取并清空意图
   * @returns 当前意图，若无则返回 null
   */
  const consumeNavIntent = (targetTab: string): INavIntent | null => {
    if (navIntent.value && navIntent.value.targetTab === targetTab) {
      const intent = { ...navIntent.value };
      navIntent.value = null;
      return intent;
    }
    return null;
  };

  /**
   * 窥探当前意图（不消费）
   * 用于目标 Tab 在数据未就绪时判断是否有待处理的跳转
   */
  const peekNavIntent = (targetTab: string): INavIntent | null => {
    if (navIntent.value && navIntent.value.targetTab === targetTab) {
      return { ...navIntent.value };
    }
    return null;
  };

  /**
   * 强制清除导航意图
   * 当用户手动切换 Tab（非跨 Tab 跳转）时，清除残留意图
   */
  const clearNavIntent = () => {
    navIntent.value = null;
  };

  return {
    navIntent,
    setNavIntent,
    consumeNavIntent,
    peekNavIntent,
    clearNavIntent,
  };
});
