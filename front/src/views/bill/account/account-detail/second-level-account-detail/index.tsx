import { computed, defineComponent, provide, reactive, ref, watch } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import useBillStore, { IMainAccountDetail } from '@/store/useBillStore';
import { Message, Button, Form, Input } from 'bkui-vue';
import { BILL_VENDORS_MAP } from '../../account-manage/constants';
import { SITE_TYPE_MAP } from '@/common/constant';
import { timeFormatter } from '@/common/util';
import { useVerify } from '@/hooks';
import PermissionDialog from '@/components/permission-dialog';
import CommonDialog from '@/components/common-dialog';
import HcmFormUser from '@/components/form/user.vue';
import isEqual from 'lodash/isEqual';
import { MENU_SERVICE_TICKET_DETAILS } from '@/constants/menu-symbol';
import routerAction from '@/router/utils/action';

const { FormItem } = Form;

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

    // 账号信息编辑弹窗相关
    const isAccountEditDialogShow = ref(false);
    const buttonLoading = ref(false);

    // 账号信息编辑表单
    const accountEditForm = reactive({
      email: '',
      managers: [] as string[],
      bak_managers: [] as string[],
      memo: '',
    });

    // 重置账号编辑表单
    const resetAccountEditForm = () => {
      accountEditForm.email = detail.value.email || '';
      accountEditForm.managers = detail.value.managers ? [...detail.value.managers] : [];
      accountEditForm.bak_managers = detail.value.bak_managers ? [...detail.value.bak_managers] : [];
      accountEditForm.memo = detail.value.memo || '';
    };

    // 判断表单数据是否有变化
    const isFormChanged = computed(
      () =>
        !isEqual(
          {
            email: accountEditForm.email,
            managers: accountEditForm.managers,
            bak_managers: accountEditForm.bak_managers,
            memo: accountEditForm.memo,
          },
          {
            email: detail.value.email || '',
            managers: detail.value.managers || [],
            bak_managers: detail.value.bak_managers || [],
            memo: detail.value.memo || '',
          },
        ),
    );

    // 打开账号信息编辑弹窗
    const openAccountEditDialog = () => {
      resetAccountEditForm();
      isAccountEditDialogShow.value = true;
    };

    // 提交账号信息编辑（触发审批）
    const handleAccountUpdate = async () => {
      try {
        buttonLoading.value = true;
        const { data } = await billStore.update_main_account({
          id: props.accountId,
          ...detail.value,
          email: accountEditForm.email,
          managers: accountEditForm.managers,
          bak_managers: accountEditForm.bak_managers,
          memo: accountEditForm.memo,
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
        isAccountEditDialogShow.value = false;
        getDetail();
      } finally {
        buttonLoading.value = false;
      }
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
    return () => (
      <div class={'account-detail-wrapper'}>
        <p class={'sub-title'}>
          帐号信息
          <span class={'edit-icon'} onClick={openAccountEditDialog}>
            <i class={'hcm-icon bkhcm-icon-bianji mr6'} />
            编辑
          </span>
        </p>
        <DetailInfo
          detail={detail.value}
          col={1}
          fields={[
            {
              prop: 'vendor',
              name: '云厂商',
              render: () => (BILL_VENDORS_MAP as Record<string, string>)[detail.value.vendor ?? ''],
            },
            { prop: 'parent_account_id', name: '一级账号ID' },
            { prop: 'id', name: '二级帐号ID' },
            { prop: 'name', name: '二级帐号名称' },
            { prop: 'cloud_id', name: '云账号id' },
            {
              prop: 'site',
              name: '站点类型',
              render: () => (SITE_TYPE_MAP as Record<string, string>)[detail.value.site ?? ''],
            },
            { prop: 'email', name: '帐号邮箱' },
            { prop: 'managers', name: '主负责人' },
            { prop: 'bak_managers', name: '备份负责人' },
            {
              prop: 'op_product_id',
              name: '业务',
            },
            { prop: 'memo', name: '备注' },
            { prop: 'creator', name: '创建者', render: () => <hcm-user-value value={detail.value.creator} /> },
            { prop: 'reviser', name: '修改者', render: () => <hcm-user-value value={detail.value.reviser} /> },
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

        {/* 账号信息编辑弹窗 */}
        <CommonDialog
          v-model:isShow={isAccountEditDialogShow.value}
          title={'编辑账号信息'}
          dialogType='operation'
          width={680}>
          {{
            default: () => (
              <Form labelWidth={130} model={accountEditForm} formType='vertical'>
                <FormItem label='账号邮箱' property='email'>
                  <Input v-model={accountEditForm.email} placeholder='请输入账号邮箱' />
                </FormItem>
                <FormItem label='主负责人' property='managers'>
                  <HcmFormUser v-model={accountEditForm.managers} />
                </FormItem>
                <FormItem label='备份负责人' property='bak_managers'>
                  <HcmFormUser v-model={accountEditForm.bak_managers} />
                </FormItem>
                <FormItem label='备注' property='memo'>
                  <Input v-model={accountEditForm.memo} type='textarea' placeholder='请输入备注' rows={3} />
                </FormItem>
              </Form>
            ),
            footer: () => (
              <div class={'validate-btn-container'}>
                <Button
                  theme='primary'
                  loading={buttonLoading.value}
                  disabled={!isFormChanged.value}
                  onClick={handleAccountUpdate}>
                  提交
                </Button>
                <Button class='ml10' onClick={() => (isAccountEditDialogShow.value = false)}>
                  取消
                </Button>
              </div>
            ),
          }}
        </CommonDialog>
      </div>
    );
  },
});
