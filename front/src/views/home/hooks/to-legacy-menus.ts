import type { RouteRecordRaw } from 'vue-router';
import type { IMenu } from '@/common/menu-service';

/**
 * 当前侧栏只认路由形状的菜单（分组 = { children, meta.groupTitle }），这里把 IMenu 转成该形状。
 * 结果只用于侧栏渲染，不注册进路由；侧栏改为直接按 IMenu 渲染后删除本文件。
 */
export const toLegacyMenus = (menus: IMenu[], views: RouteRecordRaw[]): RouteRecordRaw[] => {
  const ungrouped = Symbol('ungrouped');
  const groups = new Map<string | symbol, RouteRecordRaw[]>();

  menus.forEach((menu) => {
    const view = views.find((item) => item.name === menu.route?.name);
    if (!view) return;

    // 去掉 children：侧栏对顶层元素按「有 children 即分组」判定，页签子路由会被误当成分组
    const { children: _children, ...rest } = view;
    const item = {
      ...rest,
      meta: {
        ...view.meta,
        title: menu.i18n,
        icon: menu.icon ? `hcm-icon ${menu.icon}` : view.meta?.icon,
      },
    } as RouteRecordRaw;

    const key = menu.group ?? ungrouped;
    groups.set(key, [...(groups.get(key) ?? []), item]);
  });

  return [...groups].flatMap(([group, items]) =>
    group === ungrouped
      ? items
      : [{ path: String(group), children: items, meta: { groupTitle: String(group) } } as RouteRecordRaw],
  );
};
