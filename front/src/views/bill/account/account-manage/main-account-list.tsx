import { defineComponent, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Button } from 'bkui-vue';
import { useI18n } from 'vue-i18n';

import SecondLevelAccountDetail from '../account-detail/second-level-account-detail';
import CommonSideslider from '@/components/common-sideslider';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { secondarySearchData } from './constants';
import { MENU_BILL_MAIN_ACCOUNT_CREATE } from '@/constants/menu-symbol';

/**
 * 二级账号列表页 —— 独立路由页面
 */
export default defineComponent({
  name: 'MainAccountList',
  setup() {
    const router = useRouter();
    const { t } = useI18n();

    const { columns } = useColumns('secondaryAccount');

    const isSideSliderShow = ref(false);
    const curAccount = ref<any>({});

    const { CommonTable } = useTable({
      tableOptions: {
        columns: [
          {
            label: '二级帐号名称',
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
        searchData: secondarySearchData,
      },
      requestOption: {
        type: 'account/main_accounts',
        sortOption: { sort: 'created_at', order: 'DESC' },
        dataPath: 'data.details',
      },
    });

    const handleCreate = () => {
      router.push({ name: MENU_BILL_MAIN_ACCOUNT_CREATE });
    };

    return () => (
      <div style={{ padding: '24px', height: '100%' }}>
        <div style={{ padding: '16px 24px', height: '100%', backgroundColor: '#fff' }}>
          <CommonTable>
            {{
              operation: () => (
                // TODO: 操作权限 —— 使用 hcm-auth 组件包裹
                <Button theme='primary' onClick={handleCreate}>
                  {t('创建二级账号')}
                </Button>
              ),
            }}
          </CommonTable>

          <CommonSideslider
            v-model:isShow={isSideSliderShow.value}
            width={640}
            title={t('二级账号详情')}
            noFooter={true}>
            <SecondLevelAccountDetail accountId={curAccount.value.id} />
          </CommonSideslider>
        </div>
      </div>
    );
  },
});
