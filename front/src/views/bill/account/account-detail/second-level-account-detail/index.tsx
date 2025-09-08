import { defineComponent, provide, ref, watch } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import useBillStore, { IMainAccountDetail } from '@/store/useBillStore';
import { Message, Button } from 'bkui-vue';
import { BILL_VENDORS_MAP } from '../../account-manage/constants';
import { SITE_TYPE_MAP } from '@/common/constant';
import { timeFormatter } from '@/common/util';
import { useVerify } from '@/hooks';
import PermissionDialog from '@/components/permission-dialog';
import { MENU_SERVICE_TICKET_DETAILS } from '@/constants/menu-symbol';
import routerAction from '@/router/utils/action';

export default defineComponent({
  props: {
    accountId: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const detail = ref<IMainAccountDetail>({});
    const billStore = useBillStore();

    const {
      showPermissionDialog,
      handlePermissionConfirm,
      handlePermissionDialog,
      handleAuth,
      permissionParams,
      authVerifyData,
    } = useVerify();
    // provide 预鉴权参数
    provide('authAction', { authVerifyData, handleAuth, authId: 'main_account_edit' });

    const getDetail = async () => {
      const { data } = await billStore.main_account_detail(props.accountId);
      detail.value = data;
    };
    watch(
      () => props.accountId,
      async () => {
        await getDetail();
      },
      {
        immediate: true,
        deep: true,
      },
    );
    const handleUpdate = async (val: any) => {
      const { data } = await billStore.update_main_account({
        id: props.accountId,
        ...detail.value,
        ...val,
      });
      Message({
        message: (
          <span>
            修改申请已提交，审批通过后生效。审批信息
            <Button
              theme='primary'
              text
              onClick={() => {
                routerAction.open({
                  name: MENU_SERVICE_TICKET_DETAILS,
                  query: {
                    id: data.id,
                  },
                });
              }}>
              链接
            </Button>
          </span>
        ),
        theme: 'success',
      });
      // router.push({
      //   path: '/service/ticket/detail',
      //   query: {
      //     ...route.query,
      //     id: data.id,
      //   },
      // });
    };
    return () => (
      <div class={'account-detail-wrapper'}>
        <p class={'sub-title'}>帐号信息</p>
        <DetailInfo
          detail={detail.value}
          col={1}
          onChange={handleUpdate}
          fields={[
            { prop: 'vendor', name: '云厂商', render: () => BILL_VENDORS_MAP[detail.value.vendor] },
            { prop: 'parent_account_id', name: '一级账号ID' },
            { prop: 'id', name: '二级帐号ID' },
            { prop: 'name', name: '二级帐号名称' },
            { prop: 'cloud_id', name: '云账号id' },
            { prop: 'site', name: '站点类型', render: () => SITE_TYPE_MAP[detail.value.site] },
            { prop: 'email', name: '帐号邮箱', edit: true },
            { prop: 'managers', name: '主负责人', edit: true, type: 'member' },
            { prop: 'bak_managers', name: '备份负责人', edit: true, type: 'member' },
            {
              prop: 'op_product_id',
              name: '业务',
            },
            { prop: 'memo', name: '备注', edit: true },
            { prop: 'creator', name: '创建者' },
            { prop: 'reviser', name: '修改者' },
            { prop: 'created_at', name: '创建时间', render: () => timeFormatter(detail.value.created_at) },
            { prop: 'updated_at', name: '修改时间', render: () => timeFormatter(detail.value.updated_at) },
          ]}
        />
        {/* 申请权限 */}
        <PermissionDialog
          v-model:isShow={showPermissionDialog.value}
          params={permissionParams.value}
          onCancel={handlePermissionDialog}
          onConfirm={handlePermissionConfirm}
        />
      </div>
    );
  },
});
