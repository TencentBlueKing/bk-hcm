import { defineComponent, ref, reactive, watch } from 'vue';
import './index.scss';
import http from '@/http';
import {
  Loading,
  SearchSelect,
  Table,
  Button,
  Dialog,
  Form,
  Input,
  Select,
  Message,
} from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import MemberSelect from '@/components/MemberSelect';
import type { Column } from 'bkui-vue/lib/table/props';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { QueryRuleOPEnum } from '@/typings/common';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { VendorEnum } from '@/common/constant';
import { useRoute } from 'vue-router';
import { timeFormatter } from '@/common/util';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { FormItem } = Form;

export default defineComponent({
  setup() {
    const { t } = useI18n();
    const resourceAccountStore = useResourceAccountStore();
    const businessMapStore = useBusinessMapStore();
    const route = useRoute();
    const searchVal = ref('');
    const searchData = ref<Array<ISearchItem>>([
      {
        name: '账号 ID',
        id: 'id',
      },
    ]);
    const isLoading = ref(false);
    const columns = ref<Array<Column>>([
      {
        label: '账号名称',
        field: 'name',
      },
      {
        label: '账号 ID',
        field: 'id',
      },
      {
        label: '所属业务',
        field: 'bk_biz_ids',
        render: ({ data }: any) => (data?.bk_biz_ids.length > 0
          ? data?.bk_biz_ids
            .map((bk_biz_id: number) => {
              return businessMapStore.getNameFromBusinessMap(bk_biz_id);
            })
            ?.join(',')
          : '--'),
      },
      {
        label: '备注',
        field: 'memo',
        render: ({ cell }: any) => cell || '--',
      },
      {
        label: '负责人',
        field: 'managers',
        render: ({ data }: any) => data?.managers?.join(',') || '--',
      },
      {
        label: '更新人',
        field: 'reviser',
      },
      {
        label: '更新时间',
        field: 'updated_at',
        render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
      },
      {
        label: '操作',
        field: 'operation',
        render: ({ data }: any) => (
          <Button
            text
            theme='primary'
            onClick={() => handleModifyAccount(data)}>
            编辑
          </Button>
        ),
      },
    ]);
    const dataList = ref<any>([]);
    const pagination = reactive({
      start: 0,
      limit: 10,
      count: 100,
    });
    const handlePageLimitChange = (v: number) => {
      pagination.limit = v;
      getUserList();
    };
    const handlePageValueCHange = (v: number) => {
      pagination.start = (v - 1) * pagination.limit;
      getUserList();
    };
    const changeColumns = () => {
      [VendorEnum.TCLOUD, VendorEnum.AWS, VendorEnum.HUAWEI].includes(resourceAccountStore.resourceAccount.vendor)
        ? (columns.value = [
          {
            label: '账号名称',
            field: 'name',
            render: ({ data }: any) => {
              return (
                  <>
                    {data?.name}
                    {data?.account_type === 'current_account' ? (
                      <bk-tag theme='info' class='users-list-bk-tag'>
                        当前账号
                      </bk-tag>
                    ) : data?.account_type === 'main_account' ? (
                      <bk-tag theme='success' class='users-list-bk-tag'>
                        主账号
                      </bk-tag>
                    ) : (
                      ''
                    )}
                  </>
              );
            },
          },
          {
            label: '账号 ID',
            field: 'id',
          },
          {
            label: '所属业务',
            field: 'bk_biz_ids',
            render: ({ data }: any) => (data?.bk_biz_ids.length > 0
              ? data?.bk_biz_ids
                .map((bk_biz_id: number) => {
                  return businessMapStore.getNameFromBusinessMap(bk_biz_id);
                })
                ?.join(',')
              : '--'),
          },
          {
            label: '备注',
            field: 'memo',
            render: ({ cell }: any) => cell || '--',
          },
          {
            label: '负责人',
            field: 'managers',
            render: ({ data }: any) => data?.managers?.join(',') || '--',
          },
          {
            label: '更新人',
            field: 'reviser',
          },
          {
            label: '更新时间',
            field: 'updated_at',
            render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
          },
          {
            label: '操作',
            field: 'operation',
            render: ({ data }: any) => (
                <Button
                  text
                  theme='primary'
                  onClick={() => handleModifyAccount(data)}>
                  编辑
                </Button>
            ),
          },
        ])
        : (columns.value = [
          {
            label: '账号名称',
            field: 'name',
            render: ({ data }: any) => {
              return (
                  <>
                    {data?.name}
                    {
                      data?.account_type !== '' && (
                        <bk-tag
                          theme={data?.account_type === 'current_account' ? 'info' : 'success'}
                          class='users-list-bk-tag'>
                          { data?.account_type === 'current_account' ? '当前账号' : '主账号'}
                        </bk-tag>
                      )
                    }
                  </>
              );
            },
          },
          {
            label: '所属业务',
            field: 'bk_biz_ids',
            render: ({ data }: any) => (data?.bk_biz_ids.length > 0
              ? data?.bk_biz_ids
                .map((bk_biz_id: number) => {
                  return businessMapStore.getNameFromBusinessMap(bk_biz_id);
                })
                ?.join(',')
              : '--'),
          },
          {
            label: '备注',
            field: 'memo',
            render: ({ cell }: any) => cell || '--',
          },
          {
            label: '负责人',
            field: 'managers',
            render: ({ data }: any) => data?.managers?.join(',') || '--',
          },
          {
            label: '更新人',
            field: 'reviser',
          },
          {
            label: '更新时间',
            field: 'updated_at',
            render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
          },
          {
            label: '操作',
            field: 'operation',
            render: ({ data }: any) => (
                <Button
                  text
                  theme='primary'
                  onClick={() => handleModifyAccount(data)}>
                  编辑
                </Button>
            ),
          },
        ]);
    };
    const filter = reactive({
      op: QueryRuleOPEnum.AND,
      rules: [],
    });
    const getUserList = async (customRules: Array<{
      op: QueryRuleOPEnum;
      field: string;
      value: string | number;
    }> = []) => {
      isLoading.value = true;
      const [detailsRes, countRes] = await Promise.all([false, true].map(isCount => http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/sub_accounts/list`,
        {
          filter: {
            op: QueryRuleOPEnum.AND,
            rules: [...filter.rules, ...customRules],
          },
          page: {
            limit: isCount ? 0 : pagination.limit,
            start: isCount ? 0 : pagination.start,
            count: isCount,
          },
        },
      )));
      isLoading.value = false;
      dataList.value = detailsRes?.data?.details;
      pagination.count = countRes?.data?.count;
      changeColumns();
    };
    const isShowModifyUserDialog = ref(false);
    const isUserDialogLoading = ref(false);
    const userFormModel = reactive({
      bk_biz_ids: [],
      managers: [],
      memo: '',
      id: '',
    });
    const formRef = ref<InstanceType<typeof Form>>(null);
    const formRules = {};
    const clearUserFormParams = () => {
      Object.assign(userFormModel, {
        bk_biz_ids: [],
        managers: [],
        memo: '',
        id: '',
      });
    };
    const handleModifyAccount = (data: any) => {
      clearUserFormParams();
      console.log(data);
      isShowModifyUserDialog.value = true;
      Object.assign(userFormModel, {
        bk_biz_ids: data?.bk_biz_ids,
        managers: data?.managers,
        memo: data?.memo,
        id: data?.id,
      });
    };
    const handleModifyUserSubmit = async () => {
      await formRef.value.validate();
      try {
        isUserDialogLoading.value = true;
        await http.patch(
          `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/sub_accounts/${userFormModel.id}`,
          userFormModel,
        );
        Message({
          theme: 'success',
          message: t('编辑成功'),
        });
        isShowModifyUserDialog.value = false;
        getUserList();
      } catch (error) {
        console.log(error);
      } finally {
        isUserDialogLoading.value = false;
      }
    };

    watch(
      () => route.query.accountId,
      (newVal) => {
        // bug：一次变化，执行三次
        if (!newVal) {
          filter.rules = [];
        } else {
          filter.rules[0] = {
            op: QueryRuleOPEnum.EQ,
            field: 'account_id',
            value: newVal,
          };
        }
        getUserList();
      },
      {
        immediate: true,
      },
    );

    watch(
      () => searchVal.value,
      (vals) => {
        console.log(vals);
        filter.rules = Array.isArray(vals)
          ? [
            filter.rules[0],
            ...vals.map((val: any) => ({
              field: val?.id,
              op: QueryRuleOPEnum.EQ,
              value: val?.values?.[0]?.id,
            })),
          ]
          : [];
        getUserList();
      },
    );

    return () => (
      <div>
        <SearchSelect
          class='w500 common-search-selector'
          v-model={searchVal.value}
          data={searchData.value}
        />
        <Loading loading={isLoading.value}>
          <Table
            data={dataList.value}
            columns={columns.value}
            pagination={pagination}
            remotePagination
            onPageLimitChange={handlePageLimitChange}
            onPageValueChange={handlePageValueCHange}
            show-overflow-tooltip
            onColumnSort={() => {}}
            onColumnFilter={() => {}}></Table>
        </Loading>

        <Dialog
          isShow={isShowModifyUserDialog.value}
          width={680}
          title={'编辑用户'}
          isLoading={isUserDialogLoading.value}
          onConfirm={handleModifyUserSubmit}
          onClosed={() => (isShowModifyUserDialog.value = false)}
          theme='primary'>
          <Form
            v-model={userFormModel}
            formType='vertical'
            ref={formRef}
            rules={formRules}>
            <FormItem
              label='所属业务'
              class={'api-secret-selector'}
              property='bk_biz_ids'>
              <Select
                v-model={userFormModel.bk_biz_ids}
                showSelectAll
                multiple
                multipleMode='tag'
                collapseTags>
                {businessMapStore.businessList.map((businessItem) => {
                  return (
                    <bk-option
                      key={businessItem.id}
                      value={businessItem.id}
                      label={businessItem.name}></bk-option>
                  );
                })}
                <bk-option></bk-option>
              </Select>
            </FormItem>
            <FormItem
              label='负责人'
              class={'api-secret-selector'}
              property='managers'>
              <MemberSelect v-model={userFormModel.managers} />
            </FormItem>
            <FormItem label='备注'>
              <Input
                type={'textarea'}
                v-model={userFormModel.memo}
                maxlength={256}
                resize={false}
              />
            </FormItem>
          </Form>
        </Dialog>
      </div>
    );
  },
});
