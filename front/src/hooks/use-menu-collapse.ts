import { ref, type Ref } from 'vue';
import { useUserStore } from '@/store/user';

export const FAVORITES_SECTION_KEY = '__favorites__';

const STORAGE_KEY_PREFIX = 'hcm_menu_collapse_';

// 模块级单例状态 —— 所有消费者共享同一份数据
// 数据结构：Record<顶级菜单标识, 已折叠的 section key 列表>
const collapsedSections: Ref<Record<string, string[]>> = ref({});
let initialized = false;

const getStorageKey = () => {
  const { username } = useUserStore();
  return `${STORAGE_KEY_PREFIX}${username}`;
};

const persist = () => {
  localStorage.setItem(getStorageKey(), JSON.stringify(collapsedSections.value));
};

const load = () => {
  try {
    const raw = localStorage.getItem(getStorageKey());
    if (raw) collapsedSections.value = JSON.parse(raw);
  } catch {
    collapsedSections.value = {};
  }
};

/**
 * 通用菜单折叠状态管理，适用于"我的收藏"与普通二级菜单（bk-submenu）。
 * 通过 localStorage 按用户持久化各 section 的展开/收起状态。
 *
 * @param topMenuKey 顶级菜单标识（使用 Symbol.description）
 */
export const useMenuCollapse = () => {
  if (!initialized) {
    load();
    initialized = true;
  }

  const isCollapsed = (topMenuKey: string, sectionKey: string): boolean => {
    return collapsedSections.value[topMenuKey]?.includes(sectionKey) ?? false;
  };

  const setCollapsed = (topMenuKey: string, sectionKey: string, collapsed: boolean) => {
    if (!collapsedSections.value[topMenuKey]) {
      collapsedSections.value[topMenuKey] = [];
    }
    const list = collapsedSections.value[topMenuKey];
    const idx = list.indexOf(sectionKey);
    if (collapsed && idx === -1) {
      list.push(sectionKey);
    } else if (!collapsed && idx !== -1) {
      list.splice(idx, 1);
    }
    persist();
  };

  const toggleCollapse = (topMenuKey: string, sectionKey: string) => {
    setCollapsed(topMenuKey, sectionKey, !isCollapsed(topMenuKey, sectionKey));
  };

  /**
   * 根据持久化状态计算 bk-menu 的 openedKeys。
   * 逻辑：所有 groupKeys 中未被标记为折叠的即为展开。
   */
  const getOpenedKeys = (topMenuKey: string, allGroupKeys: string[]): string[] => {
    return allGroupKeys.filter((key) => !isCollapsed(topMenuKey, key));
  };

  return {
    isCollapsed,
    setCollapsed,
    toggleCollapse,
    getOpenedKeys,
  };
};
