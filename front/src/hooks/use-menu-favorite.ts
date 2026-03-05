import { ref, type Ref } from 'vue';
import type { IMenu } from '@/common/menu-service';
import { useUserStore } from '@/store/user';

type ClickSource = 'favorites' | 'menu';

const STORAGE_KEY_PREFIX = 'hcm_menu_fav_';

// 模块级单例状态 —— 所有消费者共享同一份数据
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

const key = (sym: symbol): string => sym.description ?? '';

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
