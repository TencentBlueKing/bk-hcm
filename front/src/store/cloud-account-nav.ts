import { ref } from 'vue';
import { defineStore } from 'pinia';

/**
 * 云账号管理 — 跨 Tab 导航状态管理
 *
 * 设计目的：
 *   之前跨 Tab 跳转时，filter 和 detailCloudId 都挂在 URL query 上，
 *   会与各子模块自身的 route.query watcher 产生竞态：
 *     1. Tab 切换时 handleTabChange 清除了 filter（或被新 Tab 的 watcher 覆写）；
 *     2. detailCloudId 作为一次性消费参数写在 URL 里，在异步数据加载完成前可能被
 *        其他 query 更新覆盖掉；
 *     3. 云密钥使用 secretFilter 而非 filter，跨 Tab 时两个 key 互相干扰。
 *
 *   现在将这些"跨 Tab 通信"参数放到 Pinia store 中：
 *     - 写入端（发起跳转的组件）调用 setNavIntent()
 *     - 读取端（目标 Tab）在首次渲染 / 数据加载后调用 consumeNavIntent()
 *     - 一旦消费即清空，不会重复触发
 */
export interface INavIntent {
  /** 目标 Tab 名称 */
  targetTab: 'secondary-account' | 'tertiary-account' | 'cloud-secret';
  /** 要自动打开详情弹窗的云账号 ID（由目标 Tab 在 fullList 中匹配） */
  detailCloudId?: string;
  /** 要注入到目标 Tab 搜索组件的筛选条件 */
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
