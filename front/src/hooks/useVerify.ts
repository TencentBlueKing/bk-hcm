import { useCommonStore } from '@/store';
// import { Verify } from '@/typings';
import {
  ref,
  Ref,
} from 'vue';

const authVerifyData = ref<any>({ permissionAction: '', urlParams: '' });
const permissionParams = ref({ system_id: '', actions: [] });

// 权限hook
export function useVerify(showPermissionDialog?: Ref<boolean>) {
  const commonStore = useCommonStore();

  // 根据参数获取权限
  const getAuthVerifyData = async (action: any[]) => {
    if (!action) return;
    // 格式化参数
    const params = action?.reduce((p, v) => {
      p.resources.push({
        action: v.action,
        resource_type: v.type,
      });
      return p;
    }, { resources: [] });
    const res = await commonStore.authVerify(params);
    if (res.data.permission) {    // 没有权限才需要获取跳转链接参数
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
    // permissionAction 用于判断按钮状态
    const permissionAction = res.data.results.reduce((p: any, e: any, i: number) => {    // 将数组转成对象
      p[`${action[i].id}`] = e.authorized;
      return p;
    }, {});

    authVerifyData.value.permissionAction  = permissionAction;
    commonStore.addAuthVerifyData(authVerifyData);    // 全局变量管理
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
    if (!authVerifyData.value?.permission) return;
    const actionItem = authVerifyData.value?.permission?.actions.filter((e: any) => e.id === actionName);
    if (!actionItem.length) return;
    permissionParams.value = {
      system_id: authVerifyData.value?.permission.system_id,
      actions: actionItem,
    };
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
  };
}
