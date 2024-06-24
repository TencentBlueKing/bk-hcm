import { defineComponent, ref, watch } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import useBillStore, { IMainAccountDetail } from '@/store/useBillStore';
import { Message } from 'bkui-vue';
import { BILL_VENDORS_MAP } from '../../account-manage/constants';
import { SITE_TYPE_MAP } from '@/common/constant';

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
      await billStore.update_main_account({
        id: props.accountId,
        ...detail.value,
        ...val,
      });
      Message({
        message: '更新已提交，等待单据审批',
        theme: 'success',
      });
      // router.push({
      //   path: '/service/my-apply/detail',
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
          wide
          onChange={handleUpdate}
          fields={[
            { prop: 'vendor', name: '云厂商', render: () => BILL_VENDORS_MAP[detail.value.vendor] },
            { prop: 'parent_account_id', name: '一级账号ID' },
            { prop: 'id', name: '二级帐号ID' },
            { prop: 'cloud_id', name: '云账号id' },
            { prop: 'site', name: '站点类型', render: () => SITE_TYPE_MAP[detail.value.site] },
            { prop: 'email', name: '帐号邮箱', edit: true },
            { prop: 'managers', name: '主负责人', edit: true, type: 'member' },
            { prop: 'bak_managers', name: '备份负责人', edit: true, type: 'member' },
            // { prop: 'business_type', name: '业务类型' },
            // { prop: 'dept_id', name: '组织架构', edit: true },
            {
              prop: 'op_product_id',
              name: '业务',
            },
            // { prop: 'status', name: '账号状态' },
            { prop: 'memo', name: '备注', edit: true },
          ]}
        />
        {/* <p class={'sub-title'}>
          API 密钥
          <span class={'edit-icon'}>
            <i class={'hcm-icon bkhcm-icon-bianji mr6'} />
            编辑
          </span>
        </p>
        <DetailInfo detail={detail.value} fields={computedExtension.value} wide /> */}
      </div>
    );
  },
});
