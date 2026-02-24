import { defineComponent, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Button } from 'bkui-vue';
import { useI18n } from 'vue-i18n';

import FirstLevelAccountDetail from '../account-detail/first-level-account-detail';
import CommonSideslider from '@/components/common-sideslider';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { searchData } from './constants';
import { MENU_BILL_ROOT_ACCOUNT_CREATE } from '@/constants/menu-symbol';

/**
 * 一级账号列表页 —— 独立路由页面
 */
export default defineComponent({
  name: 'RootAccountList',
  setup() {
    const router = useRouter();
    const { t } = useI18n();

    const { columns } = useColumns('firstAccount');

    const isSideSliderShow = ref(false);
    const curAccount = ref<any>({});

    const { CommonTable } = useTable({
      tableOptions: {
        columns: [
          {
            label: '一级帐号名称',
            field: 'name',
            render: ({ data }: any) => (
              <Button
                text
                theme='primary'
                onClick={() => {
                  curAccount.value = data;
                  isSideSliderShow.value = true;
                }}>
                {data.name}
              </Button>
            ),
          },
          ...columns,
        ],
      },
      searchOptions: {
        searchData,
      },
      requestOption: {
        type: 'account/root_accounts',
        sortOption: { sort: 'created_at', order: 'DESC' },
        dataPath: 'data.details',
      },
    });

    const handleCreate = () => {
      router.push({ name: MENU_BILL_ROOT_ACCOUNT_CREATE });
    };

    return () => (
      <>
        <CommonTable>
          {{
            operation: () => (
              // TODO: 操作权限 —— 使用 hcm-auth 组件包裹
              <Button theme='primary' onClick={handleCreate}>
                {t('录入一级账号')}
              </Button>
            ),
          }}
        </CommonTable>

        <CommonSideslider v-model:isShow={isSideSliderShow.value} width={640} title={t('一级账号详情')} noFooter={true}>
          <FirstLevelAccountDetail accountId={curAccount.value.id} />
        </CommonSideslider>
      </>
    );
  },
});
