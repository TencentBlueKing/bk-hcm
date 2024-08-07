import { useCommonStore } from '@/store';
import usePagePermissionStore from '@/store/usePagePermissionStore';
// import { Verify } from '@/typings';
import { ref } from 'vue';

export type AuthVerifyDataType = {
  permissionAction: Record<string, boolean>;
  urlParams: {
    system_id: string;
    actions: Array<{
      id: string;
      name: string;
      related_resource_types: Array<any>;
    }>;
  };
};

export type PermissionParamsType = {
  system_id: string;
  actions: Array<{
    id: string;
    name: string;
    related_resource_types: Array<any>;
  }>;
};

type paramsType = {
  action: string;
  resource_type: string;
  bk_biz_id?: number;
};
const showPermissionDialog = ref(false);
const authVerifyData = ref<AuthVerifyDataType>({ permissionAction: {}, urlParams: {} });
const permissionParams = ref<PermissionParamsType>({ system_id: '', actions: [] });

export enum IAM_CODE {
  Success = 0,
  NoPermission = 2000009,
}

// 权限hook
export function useVerify() {
  const commonStore = useCommonStore();
  const { setHasPagePermission, setPermissionMsg, logout } = usePagePermissionStore();

  // 根据参数获取权限
  const getAuthVerifyData = async (authData: any[]) => {
    if (!authData) return;
    // 格式化参数
    const params = authData?.reduce(
      (p, v) => {
        const resourceData: paramsType = {
          action: v.action,
          resource_type: v.type,
        };
        if (v.bk_biz_id) {
          // 业务需要传业务id
          resourceData.bk_biz_id = v.bk_biz_id;
        }
        p.resources.push(resourceData);
        return p;
      },
      { resources: [] },
    );
    let res;
    try {
      res = await commonStore.authVerify(params);
    } catch (err: any) {
      switch (err.code) {
        case IAM_CODE.NoPermission:
          setHasPagePermission(false);
          setPermissionMsg(err.message);
          break;
        default:
          logout();
      }
    }

    if (res?.data?.permission) {
      // 没有权限才需要获取跳转链接参数
      // 每个操作对应的参数
      const systemId = res.data.permission.system_id;
      const urlParams = res.data.permission.actions.reduce((p: any, e: any) => {
        p[e.id] = {
          system_id: systemId,
          actions: [e],
        };
        return p;
      }, {});
      authVerifyData.value.urlParams = urlParams;
    }
    // permissionAction 用于判断按钮状态 仅针对操作按钮有用
    const permissionAction = res?.data?.results.reduce((p: any, e: any, i: number) => {
      // 将数组转成对象
      p[`${authData[i].id}`] = e.authorized;
      return p;
    }, {});
    authVerifyData.value.permissionAction = permissionAction;
    commonStore.addAuthVerifyData(authVerifyData); // 全局变量管理
    return res?.data;
  };

  // 获取操作跳转链接
  const getActionPermission = async (params: any) => {
    const res = await commonStore.authActionUrl(params);
    return res;
  };

  // 关闭弹窗
  const handlePermissionDialog = () => {
    showPermissionDialog.value = false;
  };

  // 申请权限跳转
  const handlePermissionConfirm = (url: string) => {
    window.open(url);
    handlePermissionDialog();
  };

  // 处理鉴权 actionName根据接口返回值传入
  const handleAuth = (actionName: string) => {
    if (!authVerifyData.value?.permissionAction) return;
    const actionItem = authVerifyData.value?.urlParams[actionName];
    if (!actionItem) return;
    permissionParams.value = actionItem;
    showPermissionDialog.value = true;
  };

  return {
    getAuthVerifyData,
    getActionPermission,
    handlePermissionDialog,
    handlePermissionConfirm,
    handleAuth,
    authVerifyData,
    permissionParams,
    showPermissionDialog,
  };
}
