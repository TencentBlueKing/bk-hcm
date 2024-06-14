import { computed, defineComponent, ref, watch } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import useBillStore from '@/store/useBillStore';
import { Message } from 'bkui-vue';
import { timeFormatter } from '@/common/util';

export default defineComponent({
  props: {
    accountId: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const detail = ref({});
    const billStore = useBillStore();
    const getDetail = async () => {
      const { data } = await billStore.root_account_detail(props.accountId);
      detail.value = data;
    };
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
            { prop: 'cloud_secret_key', name: '密钥', render: () => detail.value.extension?.cloud_secret_key },
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
            {
              prop: 'cloud_service_secret_key',
              name: '服务密钥',
              render: () => detail.value.extension?.cloud_service_secret_key,
            },
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
            {
              prop: 'cloud_client_secret_key',
              name: '客户端密钥',
              render: () => detail.value.extension?.cloud_client_secret_key,
            },
          ];
          break;
        case 'huawei':
          extension = [
            {
              prop: 'cloud_main_account_name',
              name: '主账号名',
              render: () => detail.value.extension?.cloud_main_account_name,
            },
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
            { prop: 'cloud_secret_key', name: '密钥', render: () => detail.value.extension?.cloud_secret_key },
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
    const handleUpdate = async (val) => {
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
            { prop: 'id', name: 'id' },
            { prop: 'name', name: '名字', edit: true },
            { prop: 'vendor', name: '云厂商' },
            { prop: 'cloud_id', name: '云ID' },
            { prop: 'email', name: '邮箱', edit: true },
            { prop: 'managers', name: '负责人', edit: true, type: 'member' },
            { prop: 'bak_managers', name: '备份负责人', edit: true, type: 'member' },
            { prop: 'site', name: '站点' },
            // { prop: 'dept_id', name: '组织架构ID' },
            { prop: 'memo', name: '备注', edit: true },
            { prop: 'creator', name: '创建者' },
            { prop: 'reviser', name: '修改者' },
            { prop: 'created_at', name: '创建时间', render: () => timeFormatter(detail.value.created_at) },
            { prop: 'updated_at', name: '修改时间', render: () => timeFormatter(detail.value.updated_at) },
          ]}
        />
        <p class={'sub-title'}>
          API 密钥
          <span class={'edit-icon'}>
            <i class={'hcm-icon bkhcm-icon-bianji mr6'} />
            编辑
          </span>
        </p>
        <div class={'detail-info'}>
          <DetailInfo detail={detail.value} fields={computedExtension.value} wide />
        </div>
      </div>
    );
  },
});
