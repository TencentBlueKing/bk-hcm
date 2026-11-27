import type { NavigationGuardWithThis, RouteLocationNormalized } from 'vue-router';

/**
 * 路由守卫：将 query.id 规范化到 params.id
 * 兼容旧的详情页 URL 格式 (detail?id=xxx) → 新格式 (detail/xxx)
 */
export const normalizeDetailId: NavigationGuardWithThis<undefined> = (to: RouteLocationNormalized) => {
  if (!to.params.id && to.query.id) {
    const { id, ...restQuery } = to.query;
    return { name: to.name!, params: { ...to.params, id: id as string }, query: restQuery };
  }
};
