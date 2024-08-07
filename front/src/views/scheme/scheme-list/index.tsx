import { defineComponent, ref, reactive, watch, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { Plus, EditLine } from 'bkui-vue/lib/icon';
import { InfoBox, Message } from 'bkui-vue';
import { useSchemeStore, useAccountStore } from '@/store';
import { QueryFilterType, IPageQuery, QueryRuleOPEnum, RulesItem } from '@/typings/common';
import { ICollectedSchemeItem, ISchemeListItem, IBizType } from '@/typings/scheme';
import { VENDORS } from '@/common/constant';
import { DEPLOYMENT_ARCHITECTURE_MAP } from '@/constants';
import CloudServiceTag from '../components/cloud-service-tag';
import SchemeEditDialog from '../components/scheme-edit-dialog';
import { useVerify } from '@/hooks';
import ErrorPage from '@/views/error-pages/403';
import moment from 'moment';

import './index.scss';
import PermissionDialog from '@/components/permission-dialog';

export default defineComponent({
  name: 'SchemeListPage',
  setup() {
    const schemeStore = useSchemeStore();
    const accountStore = useAccountStore();

    const router = useRouter();

    const searchValue = ref([]);
    const bizList = ref([]);
    const bizLoading = ref(false);
    const collections = ref<{ id: number; res_id: string }[]>([]);
    const collectPending = ref(false);
    const tableListLoading = ref(false);
    const tableListData = ref<ISchemeListItem[]>([]);
    const isEditDialogOpen = ref(false);
    const selectedScheme = ref<{ id: string; name: string; bk_biz_id: number }>({
      id: '',
      name: '',
      bk_biz_id: 0,
    });
    const pagination = reactive({
      current: 1,
      count: 0,
      limit: 10,
    });
    const sortConfig = reactive({
      field: '',
      order: '',
    });
    const filterConfigs = reactive<{ field: string; value: string[] }[]>([]);
    const searchData = ref([
      { id: 'name', name: '方案名称' },
      // { id: 'bk_biz_id', name: '业务id' },
      { id: 'creator', name: '创建人' },
    ]);
    const {
      authVerifyData,
      handleAuth,
      handlePermissionConfirm,
      handlePermissionDialog,
      showPermissionDialog,
      permissionParams,
    } = useVerify();
    if (!authVerifyData.value.permissionAction.cloud_selection_find) return () => <ErrorPage />;

    const tableCols = ref([
      {
        label: '方案名称',
        minWidth: 200,
        showOverflowTooltip: true,
        render: ({ data }: { data: ISchemeListItem }) => {
          return (
            <div class='scheme-name'>
              <i
                class={[
                  'hcm-icon',
                  'collect-icon',
                  collections.value.findIndex((item) => item.res_id === data.id) > -1
                    ? 'bkhcm-icon-collect'
                    : 'bkhcm-icon-not-favorited',
                ]}
                onClick={() => handleToggleCollection(data)}
              />
              <span
                class='name-text'
                onClick={() => {
                  goToDetail(data.id);
                }}>
                {data.name}
              </span>
              <span
                class={`edit-icon ${
                  authVerifyData.value.permissionAction.cloud_selection_edit ? '' : 'hcm-no-permision-text-btn'
                }`}
                onClick={() => {
                  if (authVerifyData.value.permissionAction.cloud_selection_edit) handleOpenEditDialog(data);
                  else handleAuth('cloud_selection_edit');
                }}>
                <EditLine />
              </span>
            </div>
          );
        },
      },
      // {
      //   label: '标签',
      //   field: 'bk_biz_id',
      //   render: ({ data }: { data: ISchemeListItem }) => {
      //     if (bizLoading.value) {
      //       return <bk-loading loading theme="primary" mode="spin" size="mini" />;
      //     }
      //     if (data) {
      //       if (data.bk_biz_id < 1) {
      //         return '--';
      //       }
      //       const biz = bizList.value.find(item => item.id === data.bk_biz_id);
      //       const name = biz ? biz.name : '--';
      //       return <span class="tag">{name}</span>;
      //     }
      //   },
      // },
      {
        label: '业务类型',
        field: 'biz_type',
        showOverflowTooltip: true,
        render: ({ data }: { data: ISchemeListItem }) => {
          return data.biz_type;
        },
      },
      {
        label: '用户分布地区',
        showOverflowTooltip: true,
        render: ({ data }: { data: ISchemeListItem }) => {
          return data.user_distribution.map((item) => `${item.name}`).join('; ');
        },
      },
      {
        label: '部署架构',
        field: 'deployment_architecture',
        filter: {
          filterFn: () => true,
          list: Object.keys(DEPLOYMENT_ARCHITECTURE_MAP).map((key) => {
            return { text: DEPLOYMENT_ARCHITECTURE_MAP[key], value: key };
          }),
        },
        render: ({ data }: { data: ISchemeListItem }) => {
          return data.deployment_architecture.map((item) => DEPLOYMENT_ARCHITECTURE_MAP[item]).join(', ');
        },
      },
      {
        label: '云厂商',
        field: 'vendors',
        filter: {
          filterFn: () => true,
          list: VENDORS.map((item) => {
            return { text: item.name, value: item.id };
          }),
        },
        render: ({ data }: { data: ISchemeListItem }) => {
          return (
            <div class='vendors-list'>
              {data.vendors.map((item) => (
                <CloudServiceTag type={item} />
              ))}
            </div>
          );
        },
      },
      {
        label: '综合评分',
        width: 120,
        field: 'composite_score',
        sort: true,
        render: ({ data }: { data: ISchemeListItem }) => {
          return <span class='composite-score'>{data.composite_score || '-'}</span>;
        },
      },
      {
        label: '创建人',
        render: ({ data }: { data: ISchemeListItem }) => {
          return data.creator;
        },
      },
      {
        label: '更新时间',
        field: 'updated_at',
        sort: true,
        render: ({ data }: { data: ISchemeListItem }) => {
          return moment(data.updated_at).format('YYYY-MM-DD HH:mm:ss');
        },
      },
      {
        label: '操作',
        width: 120,
        render: ({ data }: { data: ISchemeListItem }) => {
          return (
            <bk-button
              text
              theme='primary'
              onClick={() => {
                if (!authVerifyData.value.permissionAction.cloud_selection_delete) handleAuth('cloud_selection_delete');
                else handleDelScheme(data);
              }}
              class={authVerifyData.value.permissionAction.cloud_selection_delete ? '' : 'hcm-no-permision-text-btn'}>
              删除
            </bk-button>
          );
        },
      },
    ]);

    watch(
      () => searchValue.value,
      () => {
        pagination.current = 1;
        getTableData();
      },
    );

    const getTableData = () => {
      if (searchValue.value.length > 0 || sortConfig.field || filterConfigs.length > 0) {
        getSearchTableData();
      } else {
        getNormalTableData();
      }
    };

    // 获取全部业务列表
    const getBizList = async () => {
      bizLoading.value = true;
      const res = await accountStore.getBizListWithAuth();
      bizList.value = res.data;
      const col = tableCols.value.find((item) => item.field === 'bk_biz_id');
      if (col) {
        const list = res.data.map((item: { id: string; name: string }) => {
          const { id, name } = item;
          return { text: name, value: id };
        });
        col.filter = {
          filterFn: () => true,
          maxHeight: 300,
          list,
        };
      }
      bizLoading.value = false;
    };

    // 获取业务类型列表
    const getBizTypeList = async () => {
      const pageQuery: IPageQuery = {
        count: false,
        start: 0,
        limit: 500,
      };
      const res = await schemeStore.listBizTypes(pageQuery);

      // 业务类型列筛选配置
      const col = tableCols.value.find((item) => item.field === 'biz_type');
      if (col) {
        const list = res.data.details.map((item: IBizType) => {
          const { biz_type } = item;
          return { text: biz_type, value: biz_type };
        });
        col.filter = {
          filterFn: () => true,
          maxHeight: 300,
          list,
        };
      }
    };

    // 加载表格当前页数据
    const getNormalTableData = async () => {
      tableListLoading.value = true;
      // 获取已收藏列表，未收藏列表count
      const [collectionRes, allUnCollectedRes] = await Promise.all([
        schemeStore.listCollection(),
        getUnCollectedScheme([], { start: 0, limit: 0, count: true }),
      ]);
      collections.value = collectionRes.data.map((item: ICollectedSchemeItem) => {
        return { id: item.id, res_id: item.res_id };
      });
      const collectionIds = collections.value.map((item) => item.res_id);
      pagination.count = allUnCollectedRes.data.count;

      const currentPageStartNum = (pagination.current - 1) * pagination.limit;
      const currentPageCollectedIdsLength = collectionIds.length - currentPageStartNum;

      if (currentPageCollectedIdsLength > 0 && currentPageCollectedIdsLength < pagination.limit) {
        // 当前页中收藏方案和非收藏方案混排
        const ids = collectionIds.slice(currentPageStartNum);
        const [collectedRes, unCollectedRes] = await Promise.all([
          getCollectedSchemes(ids),
          getUnCollectedScheme(collectionIds, {
            start: 0,
            limit: pagination.limit - ids.length,
          }),
        ]);
        tableListData.value = [...collectedRes.data.details, ...unCollectedRes.data.details];
      } else if (currentPageCollectedIdsLength >= pagination.limit) {
        // 当前页中只有收藏方案
        const ids = collectionIds.slice(currentPageStartNum, currentPageStartNum + pagination.limit);
        const res = await getCollectedSchemes(ids);
        tableListData.value = res.data.details;
      } else {
        // 当前页中只有非收藏方案
        const res = await getUnCollectedScheme(collectionIds, {
          start: currentPageStartNum - collectionIds.length,
          limit: pagination.limit,
        });
        tableListData.value = res.data.details;
      }

      tableListLoading.value = false;
    };

    // 搜索表格数据
    const getSearchTableData = async () => {
      tableListLoading.value = true;
      const resWithCount = await getUnCollectedScheme([], {
        start: 0,
        limit: 0,
        count: true,
      });

      const pageQuery: IPageQuery = {
        start: (pagination.current - 1) * pagination.limit,
        limit: pagination.limit,
      };

      if (sortConfig.field) {
        pageQuery.sort = sortConfig.field;
        pageQuery.order = sortConfig.order.toUpperCase();
      }

      const res = await getUnCollectedScheme([], pageQuery);
      pagination.count = resWithCount.data.count;
      tableListData.value = res.data.details;
      tableListLoading.value = false;
    };

    // 获取已收藏方案列表
    const getCollectedSchemes = (ids: string[]) => {
      const filterQuery: QueryFilterType = {
        op: 'and',
        rules: [{ field: 'id', op: QueryRuleOPEnum.IN, value: ids }],
      };
      const pageQuery: IPageQuery = {
        start: 0,
        limit: ids.length,
      };

      return schemeStore.listCloudSelectionScheme(filterQuery, pageQuery);
    };

    // 获取未被收藏的方案列表
    const getUnCollectedScheme = (ids: string[], pageQuery: IPageQuery) => {
      const rules = searchValue.value
        .filter((item) => item.values?.length > 0)
        .map((item) => {
          if (['composite_score', 'bk_biz_id'].includes(item.id)) {
            return {
              field: item.id,
              op: QueryRuleOPEnum.EQ,
              value: Number(item.values[0].id),
            };
          }
          return {
            field: item.id,
            op: QueryRuleOPEnum.CIS,
            value: item.values[0].id,
          };
        });

      if (filterConfigs.length > 0) {
        filterConfigs.forEach((filter) => {
          if (['vendors', 'deployment_architecture'].includes(filter.field)) {
            const multiFieldsRule: { op: QueryRuleOPEnum; rules: RulesItem[] } = { op: QueryRuleOPEnum.OR, rules: [] };
            filter.value.forEach((val) => {
              multiFieldsRule.rules.push({
                field: filter.field,
                op: QueryRuleOPEnum.JSON_CONTAINS,
                value: val,
              });
            });
            // @ts-ignore
            rules.push(multiFieldsRule);
          } else {
            rules.push({
              field: filter.field,
              op: QueryRuleOPEnum.IN,
              value: filter.value,
            });
          }
        });
      }

      const filterQuery: QueryFilterType = {
        op: 'and',
        rules,
      };

      if (ids.length > 0) {
        filterQuery.rules.push({
          field: 'id',
          op: QueryRuleOPEnum.NIN,
          value: ids,
        });
      }

      return schemeStore.listCloudSelectionScheme(filterQuery, pageQuery);
    };

    // 跳转创建方案
    const goToCreate = () => {
      router.push({ name: 'scheme-recommendation' });
    };

    // 跳转方案详情
    const goToDetail = (id: string) => {
      router.push({ name: 'scheme-detail', query: { sid: id } });
    };

    // 收藏/取消收藏
    const handleToggleCollection = async (scheme: ISchemeListItem) => {
      if (collectPending.value) {
        return;
      }

      collectPending.value = true;
      const index = collections.value.findIndex((item) => item.res_id === scheme.id);
      if (index > -1) {
        await schemeStore.deleteCollection(collections.value[index].id);
        collections.value.splice(index, 1);
        Message({
          theme: 'success',
          message: '取消收藏成功',
        });
      } else {
        const res = await schemeStore.createCollection(scheme.id);
        collections.value.push({ id: res.data.id, res_id: scheme.id }); // @todo 收藏成功后, 需要后台接口返回收藏ID
        Message({
          theme: 'success',
          message: '收藏成功',
        });
      }
      collectPending.value = false;
    };

    const handleOpenEditDialog = (scheme: ISchemeListItem) => {
      isEditDialogOpen.value = true;
      selectedScheme.value = {
        id: scheme.id,
        name: scheme.name,
        bk_biz_id: scheme.bk_biz_id,
      };
    };

    // 删除方案
    const handleDelScheme = (scheme: ISchemeListItem) => {
      InfoBox({
        title: '请确认是否删除',
        subTitle: `将删除【${scheme.name}】`,
        headerAlign: 'center',
        footerAlign: 'center',
        contentAlign: 'center',
        onConfirm() {
          schemeStore.deleteCloudSelectionScheme([scheme.id]).then(() => {
            if (tableListData.value.length === 1 && pagination.current !== 1) {
              pagination.current = 1;
            }
            getTableData();
            Message({
              theme: 'success',
              message: '删除成功',
            });
          });
        },
      });
    };

    const saveSchemeFn = (data: { name: string; bk_biz_id: number }) => {
      return schemeStore.updateCloudSelectionScheme(selectedScheme.value.id, data);
    };

    const handleConfirm = () => {
      Message({
        theme: 'success',
        message: '方案编辑成功',
      });
      isEditDialogOpen.value = false;
      getTableData();
    };

    const handlePageValueChange = (val: number) => {
      pagination.current = val;
      getTableData();
    };

    const handlePageLimitChange = (val: number) => {
      pagination.current = 1;
      pagination.limit = val;
      getTableData();
    };

    // 列排序
    const handleColumnSort = ({ type, column }: { type: string; column: { field: string } }) => {
      if (type !== 'null') {
        sortConfig.field = column.field;
        sortConfig.order = type;
      } else {
        sortConfig.field = '';
        sortConfig.order = '';
      }
      getTableData();
    };

    const handleColumnFilter = ({ checked, column }: { checked: string[]; column: { field: string } }) => {
      const index = filterConfigs.findIndex((filter) => filter.field === column.field);
      if (index > -1) {
        if (checked.length > 0) {
          filterConfigs.splice(index, 1, {
            field: column.field,
            value: checked,
          });
        } else {
          filterConfigs.splice(index, 1);
        }
      } else if (checked.length > 0) {
        filterConfigs.push({ field: column.field, value: checked });
      }
      pagination.current = 1;
      getTableData();
    };

    onMounted(() => {
      getBizList();
      getBizTypeList();
      getNormalTableData();
    });

    return () => (
      <div class='scheme-list-page'>
        <div class='operate-wrapper'>
          <bk-button
            class={`create-btn ${
              authVerifyData.value.permissionAction.cloud_selection_recommend ? '' : 'hcm-no-permision-btn'
            }`}
            theme='primary'
            onClick={() => {
              if (authVerifyData.value.permissionAction.cloud_selection_recommend) goToCreate();
              else handleAuth('cloud_selection_create');
            }}>
            <Plus class='plus-icon' />
            创建选型方案
          </bk-button>
          <bk-search-select
            v-model={searchValue.value}
            class={'scheme-search-select'}
            data={searchData.value}></bk-search-select>
        </div>
        <div class='scheme-table-wrapper'>
          <bk-loading loading={tableListLoading.value}>
            <bk-table
              data={tableListData.value}
              pagination={pagination}
              remote-pagination
              pagination-height={60}
              border={['outer']}
              columns={tableCols.value}
              onPageValueChange={handlePageValueChange}
              onPageLimitChange={handlePageLimitChange}
              onColumnSort={handleColumnSort}
              onColumnFilter={handleColumnFilter}></bk-table>
          </bk-loading>
        </div>
        <SchemeEditDialog
          v-model:show={isEditDialogOpen.value}
          title='编辑方案'
          schemeData={selectedScheme.value || {}}
          confirmFn={saveSchemeFn}
          onConfirm={handleConfirm}
        />
        <PermissionDialog
          isShow={showPermissionDialog.value}
          onConfirm={handlePermissionConfirm}
          onCancel={handlePermissionDialog}
          params={permissionParams.value}
        />
      </div>
    );
  },
});
