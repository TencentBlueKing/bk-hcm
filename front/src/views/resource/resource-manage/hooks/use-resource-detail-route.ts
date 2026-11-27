import { computed, type Ref } from 'vue';
import { useRoute } from 'vue-router';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { MENU_RESOURCE_DETAIL } from '@/constants/menu-symbol';
import {
  AUTH_BIZ_DELETE_IAAS_RESOURCE,
  AUTH_BIZ_UPDATE_IAAS_RESOURCE,
  AUTH_DELETE_IAAS_RESOURCE,
  AUTH_UPDATE_IAAS_RESOURCE,
} from '@/constants/auth-symbols';

// 详情组件由资源纳管详情（/resource/manage/:resourceType/details/:id）与业务详情（/business/:type/detail/:id）共用
// 两侧旧的 ?id= 地址都在路由层转成 params.id
export const useResourceDetailRoute = () => {
  const route = useRoute();
  const isResourceDetail = computed(() => route.name === MENU_RESOURCE_DETAIL);
  const detailId = computed(() => route.params.id as string);
  return { isResourceDetail, detailId };
};

/**
 * 资源纳管详情按账号鉴权，业务详情按业务鉴权
 * 两侧按路由区分：getBizsId 会回退到 localStorage，资源侧也可能取到业务 ID
 * 账号 / 业务 ID 未就绪前不鉴权，按钮按 authPending 禁用
 */
export const useResourceDetailAuth = (accountId: Ref<string | undefined>) => {
  const { isResourceDetail } = useResourceDetailRoute();
  const { getBizsId } = useWhereAmI();

  const relationId = computed(() => (isResourceDetail.value ? accountId.value : getBizsId()));
  const authPending = computed(() => !relationId.value);
  const updateSign = computed(() => ({
    type: isResourceDetail.value ? AUTH_UPDATE_IAAS_RESOURCE : AUTH_BIZ_UPDATE_IAAS_RESOURCE,
    relation: [relationId.value],
  }));
  const deleteSign = computed(() => ({
    type: isResourceDetail.value ? AUTH_DELETE_IAAS_RESOURCE : AUTH_BIZ_DELETE_IAAS_RESOURCE,
    relation: [relationId.value],
  }));

  return { isResourceDetail, authPending, updateSign, deleteSign };
};
