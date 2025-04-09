import { inject, provide, reactive } from 'vue';
import { permissionDialogSymbol } from '@/constants/provide-symbols';
import { type IVerifyResult } from '@/store/auth';

export type PermissionDialogContext = {
  isShow: boolean;
  permission: IVerifyResult['permission'];
  done?: () => void;
};

export type PermissionDialogOptions = Pick<PermissionDialogContext, 'done'>;

export const providePermissionDialog = () => {
  // 使用时需避免解构会丢失响应
  const context = reactive<PermissionDialogContext>({
    isShow: false,
    permission: null,
  });

  provide(permissionDialogSymbol, context);

  return context;
};

export default function usePermissionDialog() {
  const permissionDialogContext = inject<PermissionDialogContext>(permissionDialogSymbol);

  const show = (permission: PermissionDialogContext['permission'], options?: PermissionDialogOptions) => {
    permissionDialogContext.permission = permission;
    permissionDialogContext.done = options?.done;
    permissionDialogContext.isShow = true;
  };

  return {
    show,
  };
}
