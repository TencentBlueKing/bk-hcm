import { computed, defineComponent, ref, watch } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import useBillStore from '@/store/useBillStore';

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
      const { data } = await billStore.main_account_detail(props.accountId);
      detail.value = data;
    };
    const computedExtension = computed(() => {
      let extension = [
        {
          prop: 'cloud_main_account_name',
          name: '二级账号名',
          render: () => detail.value.extension?.cloud_main_account_name,
        },
        {
          prop: 'cloud_main_account_id',
          name: '二级账号ID',
          render: () => detail.value.extension?.cloud_main_account_id,
        },
      ];
      switch (detail.value.vendor) {
        case 'aws':
        case 'huawei':
        case 'zenlayer':
        case 'kaopu':
          extension = [
            {
              prop: 'cloud_main_account_name',
              name: '二级账号名',
              render: () => detail.value.extension?.cloud_main_account_name,
            },
            {
              prop: 'cloud_main_account_id',
              name: '二级账号ID',
              render: () => detail.value.extension?.cloud_main_account_id,
            },
          ];
          break;
        case 'gcp':
          extension = [
            { prop: 'cloud_project_name', name: '云项目名', render: () => detail.value.extension?.cloud_project_name },
            { prop: 'cloud_project_id', name: '云项目ID', render: () => detail.value.extension?.cloud_project_id },
          ];
          break;
        case 'azure':
          extension = [
            {
              prop: 'cloud_subscription_name',
              name: '订阅名',
              render: () => detail.value.extension?.cloud_subscription_name,
            },
            {
              prop: 'cloud_subscription_id',
              name: '订阅ID',
              render: () => detail.value.extension?.cloud_subscription_id,
            },
          ];
          break;
      }
      return extension;
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
    return () => (
      <div class={'account-detail-wrapper'}>
        <p class={'sub-title'}>帐号信息</p>
        <DetailInfo
          detail={detail.value}
          wide
          fields={[
            { prop: 'vendor', name: '云厂商' },
            { prop: 'parent_account_id', name: '一级账号ID' },
            { prop: 'id', name: '二级帐号ID' },
            { prop: 'cloud_id', name: '云账号id' },
            { prop: 'site', name: '站点类型' },
            { prop: 'email', name: '帐号邮箱' },
            { prop: 'managers', name: '主负责人', edit: true },
            { prop: 'bak_managers', name: '备份负责人', edit: true },
            { prop: 'business_type', name: '业务类型' },
            { prop: 'dept_id', name: '组织架构', edit: true },
            { prop: 'op_product_id', name: '运营产品' },
            { prop: 'status', name: '账号状态' },
            { prop: 'memo', name: '备注' },
          ]}
        />
        <p class={'sub-title'}>API 密钥</p>
        <DetailInfo detail={detail.value} fields={computedExtension.value} wide />
      </div>
    );
  },
});
