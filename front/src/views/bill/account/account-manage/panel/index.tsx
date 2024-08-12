import { PropType, defineComponent, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Button } from 'bkui-vue';
import FirstLevelAccountDetail from '../../account-detail/first-level-account-detail';
import SecondLevelAccountDetail from '../../account-detail/second-level-account-detail';
import CommonSideslider from '@/components/common-sideslider';
import ErrorPage from '@/views/error-pages/403';

import { useI18n } from 'vue-i18n';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { AccountLevelEnum, searchData, secondarySearchData } from '../constants';

export default defineComponent({
  props: { accountLevel: String as PropType<AccountLevelEnum>, authVerifyData: Object },
  setup(props) {
    if (
      props.accountLevel === AccountLevelEnum.SecondLevel &&
      !props.authVerifyData.permissionAction?.main_account_find
    ) {
      return () => <ErrorPage urlKeyId='main_account_find' />;
    }

    const router = useRouter();
    const route = useRoute();
    const { t } = useI18n();

    const { columns } = useColumns(props.accountLevel);

    const isSideSliderShow = ref(false);
    const curAccount = ref({});

    const { CommonTable } = useTable({
      tableOptions: {
        columns: [
          {
            label: props.accountLevel === AccountLevelEnum.FirstLevel ? '一级帐号名称' : '二级帐号名称',
            field: 'name',
            render: ({ data }: any) => (
              <Button
                text
                theme='primary'
                onClick={() => {
                  // SideSlider展示详情(可编辑)
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
        searchData: props.accountLevel === AccountLevelEnum.FirstLevel ? searchData : secondarySearchData,
      },
      requestOption: {
        type: props.accountLevel === AccountLevelEnum.FirstLevel ? 'account/root_accounts' : 'account/main_accounts',
        sortOption: { sort: 'created_at', order: 'DESC' },
        dataPath: 'data.details',
      },
    });

    return () => (
      <>
        <CommonTable>
          {{
            operation: () => (
              <Button
                theme='primary'
                onClick={() => {
                  router.push({
                    path:
                      props.accountLevel === AccountLevelEnum.FirstLevel
                        ? '/bill/account-manage/first-account'
                        : '/bill/account-manage/second-account',
                    query: { ...route.query },
                  });
                }}>
                {props.accountLevel === AccountLevelEnum.FirstLevel ? t('录入一级账号') : t('创建二级账号')}
              </Button>
            ),
          }}
        </CommonTable>

        {/* 一级账号详情及编辑 */}
        <CommonSideslider
          v-model:isShow={isSideSliderShow.value}
          width={640}
          title={props.accountLevel === AccountLevelEnum.FirstLevel ? t('一级账号详情') : t('二级账号详情')}
          noFooter={true}>
          {props.accountLevel === AccountLevelEnum.FirstLevel ? (
            <FirstLevelAccountDetail accountId={curAccount.value.id} />
          ) : (
            <SecondLevelAccountDetail accountId={curAccount.value.id} />
          )}
        </CommonSideslider>
      </>
    );
  },
});
