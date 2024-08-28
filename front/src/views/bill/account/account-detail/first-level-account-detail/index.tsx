import { computed, defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import useBillStore, { IRootAccountDetail } from '@/store/useBillStore';
import { Button, Form, Input, Message } from 'bkui-vue';
import { timeFormatter } from '@/common/util';
import { VendorEnum } from '@/common/constant';
import CommonDialog from '@/components/common-dialog';
import {
  ValidateStatus,
  useSecretExtension,
} from '@/views/resource/resource-manage/account/createAccount/components/accountForm/useSecretExtension';
import { SecretModel } from '@/typings/account';

const { FormItem } = Form;

export default defineComponent({
  props: {
    accountId: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const detail = ref<IRootAccountDetail>({});
    const billStore = useBillStore();
    const getDetail = async () => {
      const { data } = await billStore.root_account_detail(props.accountId);
      detail.value = data;
    };
    const isEditDialogShow = ref(false);
    const buttonLoading = ref(false);
    const formDiaRef = ref(null);

    const initSecretModel: SecretModel = {
      secretId: '',
      secretKey: '',
      subAccountId: '',
      iamUserName: '',
    };

    const { curExtension, isValidateLoading, handleValidate, isValidateDiasbled } = useSecretExtension({
      vendor: detail.value.vendor as VendorEnum,
    }, true);

    const secretModel = reactive<SecretModel>({
      ...initSecretModel,
    });

    watch(
      () => props.accountId,
      () => {
        getDetail();
      },
      {
        immediate: true,
        deep: true,
      },
    );
    const computedExtension = computed(() => {
      let extension: any[] = [];

      switch (detail.value.vendor) {
        case 'aws':
          extension = [
            { prop: 'cloud_account_id', name: '一级账号ID', render: () => detail.value.extension?.cloud_account_id },
            { prop: 'cloud_iam_username', name: 'IAM用户名', render: () => detail.value.extension?.cloud_iam_username },
            { prop: 'cloud_secret_id', name: '密钥ID', render: () => detail.value.extension?.cloud_secret_id },
            // { prop: 'cloud_secret_key', name: '密钥', render: () => detail.value.extension?.cloud_secret_key },
          ];
          break;
        case 'gcp':
          extension = [
            { prop: 'email', name: '邮箱', render: () => detail.value.extension?.email },
            { prop: 'cloud_project_id', name: '云项目ID', render: () => detail.value.extension?.cloud_project_id },
            { prop: 'cloud_project_name', name: '云项目名', render: () => detail.value.extension?.cloud_project_name },
            {
              prop: 'cloud_service_account_id',
              name: '服务账号ID',
              render: () => detail.value.extension?.cloud_service_account_id,
            },
            {
              prop: 'cloud_service_account_name',
              name: '服务账号名',
              render: () => detail.value.extension?.cloud_service_account_name,
            },
            {
              prop: 'cloud_service_secret_id',
              name: '服务密钥ID',
              render: () => detail.value.extension?.cloud_service_secret_id,
            },
            // {
            //   prop: 'cloud_service_secret_key',
            //   name: '服务密钥',
            //   render: () => detail.value.extension?.cloud_service_secret_key,
            // },
          ];
          break;
        case 'azure':
          extension = [
            { prop: 'display_name_name', name: '显示名称', render: () => detail.value.extension?.display_name_name },
            { prop: 'cloud_tenant_id', name: '租户ID', render: () => detail.value.extension?.cloud_tenant_id },
            {
              prop: 'cloud_subscription_id',
              name: '订阅ID',
              render: () => detail.value.extension?.cloud_subscription_id,
            },
            {
              prop: 'cloud_subscription_name',
              name: '订阅名',
              render: () => detail.value.extension?.cloud_subscription_name,
            },
            {
              prop: 'cloud_application_id',
              name: '应用ID',
              render: () => detail.value.extension?.cloud_application_id,
            },
            {
              prop: 'cloud_application_name',
              name: '应用名',
              render: () => detail.value.extension?.cloud_application_name,
            },
            {
              prop: 'cloud_client_secret_id',
              name: '客户端密钥ID',
              render: () => detail.value.extension?.cloud_client_secret_id,
            },
            // {
            //   prop: 'cloud_client_secret_key',
            //   name: '客户端密钥',
            //   render: () => detail.value.extension?.cloud_client_secret_key,
            // },
          ];
          break;
        case 'huawei':
          extension = [
            // {
            //   prop: 'cloud_main_account_name',
            //   name: '主账号名',
            //   render: () => detail.value.extension?.cloud_main_account_name,
            // },
            {
              prop: 'cloud_sub_account_id',
              name: '子账号ID',
              render: () => detail.value.extension?.cloud_sub_account_id,
            },
            {
              prop: 'cloud_sub_account_name',
              name: '子账号名',
              render: () => detail.value.extension?.cloud_sub_account_name,
            },
            { prop: 'cloud_secret_id', name: '密钥ID', render: () => detail.value.extension?.cloud_secret_id },
            // { prop: 'cloud_secret_key', name: '密钥', render: () => detail.value.extension?.cloud_secret_key },
            { prop: 'cloud_iam_user_id', name: 'IAM用户ID', render: () => detail.value.extension?.cloud_iam_user_id },
            { prop: 'cloud_iam_username', name: 'IAM用户名', render: () => detail.value.extension?.cloud_iam_username },
          ];
          break;
        case 'zenlayer':
        case 'kaopu':
          extension = [
            { prop: 'cloud_account_id', name: '一级账号ID', render: () => detail.value.extension?.cloud_account_id },
          ];
          break;
      }

      return extension;
    });
    const handleUpdate = async (val: any) => {
      await billStore.root_account_update(props.accountId, val);
      Message({
        message: '更新成功',
        theme: 'success',
      });
      getDetail();
    };
    return () => (
      <div class={'account-detail-wrapper'}>
        <p class={'sub-title'}>帐号信息</p>

        <DetailInfo
          wide
          detail={detail.value}
          onChange={handleUpdate}
          fields={[
            { prop: 'name', name: '一级账号名称', edit: true },
            { prop: 'cloud_id', name: '一级账号ID' },
            { prop: 'email', name: '账号邮箱' },
            { prop: 'managers', name: '主负责人', edit: true, type: 'member' },
            { prop: 'bak_managers', name: '备份负责人', edit: true, type: 'member' },
            { prop: 'memo', name: '备注', edit: true },
            { prop: 'creator', name: '创建者' },
            { prop: 'reviser', name: '修改者' },
            { prop: 'created_at', name: '创建时间', render: () => timeFormatter(detail.value.created_at) },
            { prop: 'updated_at', name: '修改时间', render: () => timeFormatter(detail.value.updated_at) },
          ]}
        />
        <p class={'sub-title'}>
          API 密钥
          {![VendorEnum.KAOPU, VendorEnum.ZENLAYER].includes(detail.value.vendor as VendorEnum) && (
            <span
              class={'edit-icon'}
              onClick={() => {
                isEditDialogShow.value = true;
              }}>
              <i class={'hcm-icon bkhcm-icon-bianji mr6'} />
              编辑
            </span>
          )}
        </p>
        <div class={'detail-info'}>
          <DetailInfo detail={detail.value} fields={computedExtension.value} wide />
        </div>

        {/* <Dialog isShow={isEditDialogShow.value} title='编辑API密钥'>
          1231
        </Dialog> */}
        <CommonDialog v-model:isShow={isEditDialogShow.value} title={'编辑API密钥'} dialogType='operation'>
          {{
            default: () => (
              <>
                <Form labelWidth={130} model={secretModel} ref={formDiaRef} formType='vertical'>
                  {Object.entries(curExtension.value.input).map(([property, { label }]) => (
                    <FormItem label={label} property={property}>
                      <Input
                        v-model={curExtension.value.input[property].value}
                        type={
                          property === 'cloud_service_secret_key' && detail.value.vendor === VendorEnum.GCP
                            ? 'textarea'
                            : 'text'
                        }
                        rows={8}
                      />
                    </FormItem>
                  ))}
                  {[curExtension.value.output1, curExtension.value.output2].map((output) =>
                    Object.entries(output).map(([property, { label, placeholder }]) => (
                      <FormItem label={label} property={property}>
                        <Input v-model={output[property].value} readonly placeholder={placeholder} />
                      </FormItem>
                    )),
                  )}
                </Form>
              </>
            ),
            footer: () => (
              <div class={'validate-btn-container'}>
                <Button
                  theme='primary'
                  class={'validate-btn'}
                  loading={isValidateLoading.value}
                  onClick={() => handleValidate()}
                  disabled={isValidateDiasbled.value}>
                  账号校验
                </Button>
                <Button
                  theme='primary'
                  disabled={isValidateDiasbled.value || curExtension.value.validatedStatus !== ValidateStatus.YES}
                  loading={buttonLoading.value}
                  onClick={async () => {
                    try {
                      buttonLoading.value = true;
                      await handleUpdate({
                        extension: secretModel,
                      });
                    } finally {
                      buttonLoading.value = false;
                    }
                  }}>
                  {'确认'}
                </Button>
                <Button class='ml10' onClick={() => (isEditDialogShow.value = false)}>
                  {'取消'}
                </Button>
              </div>
            ),
          }}
        </CommonDialog>
      </div>
    );
  },
});
