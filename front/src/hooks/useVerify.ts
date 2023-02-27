import { useCommonStore } from '@/store';
// import { Verify } from '@/typings';
import {
  onMounted,
  ref,
  Ref,
} from 'vue';

// 权限hook
export function useVerify(showPermissionDialog?: Ref<boolean>, actionData?: string[]) {
  const commonStore = useCommonStore();
  const authVerifyData = ref<any>({});
  const permissionParams = ref({ system_id: '', actions: [] });


  // 根据参数获取权限
  const getAuthVerifyData = async (action: string[]) => {
    if (!action) return;
    const params = action?.reduce((p, v) => {
      p.resources.push({
        action: v,
        resource_type: 'account',
      });
      return p;
    }, { resources: [] });
    const res = await commonStore.authVerify(params, action);
    authVerifyData.value = res.data;
    return res.data;
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
    const actionItem = authVerifyData.value?.permission?.actions.filter((e: any) => e.id === actionName);
    if (!authVerifyData.value?.permission || !actionItem.length) return;
    permissionParams.value = {
      system_id: authVerifyData.value?.permission.system_id,
      actions: actionItem,
    };
    showPermissionDialog.value = true;
  };
  // 处理页面需要鉴权的信息
  onMounted(() => {
    getAuthVerifyData(actionData);
  });

  return {
    getAuthVerifyData,
    getActionPermission,
    handlePermissionDialog,
    handlePermissionConfirm,
    handleAuth,
    authVerifyData,
    permissionParams,
  };
}
