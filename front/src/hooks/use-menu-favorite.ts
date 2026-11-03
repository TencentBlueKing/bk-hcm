import { ref, type Ref } from 'vue';
import type { IMenu } from '@/common/menu-service';
import { useUserStore } from '@/store/user';

/** 最近一次点亮当前页的位置：收藏区或常规菜单。只用于激活项弱化，不落盘 */
type ClickSource = 'favorites' | 'menu';

const STORAGE_KEY_PREFIX = 'hcm_menu_fav_';

// 模块级单例 —— 侧栏、收藏区、菜单项共享同一份数据
// 结构：Record<一级菜单 Symbol.description, 收藏项 description 列表>（按收藏顺序）
const favorites: Ref<Record<string, string[]>> = ref({});
const lastClickSource: Ref<ClickSource> = ref('menu');
let initialized = false;

const getStorageKey = () => {
  const { username } = useUserStore();
  return `${STORAGE_KEY_PREFIX}${username}`;
};

const persist = () => {
  localStorage.setItem(getStorageKey(), JSON.stringify(favorites.value));
};

const load = () => {
  try {
    const raw = localStorage.getItem(getStorageKey());
    if (raw) favorites.value = JSON.parse(raw);
  } catch {
    favorites.value = {};
  }
};

// Symbol 不能 JSON 序列化，用 description 作存储键（menu-symbol.ts 里每个 Symbol 的 description 唯一）
const key = (sym: symbol): string => sym.description ?? '';

/**
 * 侧栏菜单收藏。按用户写入 localStorage，按一级菜单分组。
 * 同一菜单在收藏区与常规菜单各有一份入口，lastClickSource 记录最近一次点击来源，
 * 供另一处把激活项弱化（menu.vue / menu-favorites.vue）。
 */
export const useMenuFavorite = () => {
  if (!initialized) {
    load();
    initialized = true;
  }

  const isFavorite = (topMenuId: symbol, menuId: symbol): boolean => {
    const section = favorites.value[key(topMenuId)];
    return section?.includes(key(menuId)) ?? false;
  };

  const addFavorite = (topMenuId: symbol, menuId: symbol) => {
    const k = key(topMenuId);
    const mk = key(menuId);
    if (!favorites.value[k]) {
      favorites.value[k] = [];
    }
    if (!favorites.value[k].includes(mk)) {
      favorites.value[k].push(mk);
      persist();
    }
  };

  const removeFavorite = (topMenuId: symbol, menuId: symbol) => {
    const k = key(topMenuId);
    const section = favorites.value[k];
    if (!section) return;
    const idx = section.indexOf(key(menuId));
    if (idx !== -1) {
      section.splice(idx, 1);
      persist();
    }
  };

  const toggleFavorite = (topMenuId: symbol, menuId: symbol) => {
    if (isFavorite(topMenuId, menuId)) {
      removeFavorite(topMenuId, menuId);
    } else {
      addFavorite(topMenuId, menuId);
    }
  };

  const getFavoriteCount = (topMenuId: symbol): number => {
    return favorites.value[key(topMenuId)]?.length ?? 0;
  };

  /**
   * 按收藏顺序还原菜单项。allMenus 应已按权限过滤。
   * 菜单已删除、或 description 对不上的收藏会顺手清掉并写回 localStorage。
   */
  const getFavoriteMenus = (topMenuId: symbol, allMenus: IMenu[]): IMenu[] => {
    const k = key(topMenuId);
    const section = favorites.value[k];
    if (!section?.length) return [];

    const matched: IMenu[] = [];
    const valid: string[] = [];
    for (const desc of section) {
      const menu = allMenus.find((m) => key(m.id as symbol) === desc);
      if (menu) {
        matched.push(menu);
        valid.push(desc);
      }
    }

    if (valid.length !== section.length) {
      favorites.value[k] = valid;
      persist();
    }

    return matched;
  };

  const setClickSource = (source: ClickSource) => {
    lastClickSource.value = source;
  };

  return {
    favorites,
    lastClickSource,
    isFavorite,
    addFavorite,
    removeFavorite,
    toggleFavorite,
    getFavoriteCount,
    getFavoriteMenus,
    setClickSource,
  };
};
