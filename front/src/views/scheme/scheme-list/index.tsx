import { defineComponent, ref, reactive, computed, onMounted } from "vue";
import { useRouter } from 'vue-router';
import { Message } from 'bkui-vue/lib';
import { Plus } from "bkui-vue/lib/icon";
import { useSchemeStore } from '@/store';
import { QueryFilterType, IPageQuery, QueryRuleOPEnum } from '@/typings/common';
import { ICollectedSchemeItem, ISchemeListItem } from '@/typings/scheme';
import { getScoreColor } from '@/common/util';
import SearchInput from "../components/search-input/index";
import CloudServiceTag from "../components/cloud-service-tag";

import './index.scss';

export default defineComponent({
  name: 'scheme-list-page',
  setup () {
    const schemeStore = useSchemeStore();

    const router = useRouter();

    const searchStr = ref('');
    const collections = ref<{ id: number; res_id: string; }[]>([]);
    const collectPending = ref(false);
    const pagination = reactive({
        current: 1,
        count: 0,
        limit: 10,
    });

    const tableListLoading = ref(false);
    const tableListData = ref<ISchemeListItem[]>([]);

    const tableCols = [
      {
        label: '方案名称',
        minWidth: 200,
        render: ({ data }: { data: ISchemeListItem }) => {
          return (
            <div class="scheme-name">
              <i
                class={['hcm-icon', 'collect-icon', collections.value.findIndex(item => item.res_id === data.id) > -1 ? 'bkhcm-icon-collect' : 'bkhcm-icon-not-favorited']}
                onClick={() => handleToggleCollection(data)}/>
              <bk-button text theme="primary" onClick={() => { goToDetail(data.id) }}>{data.name}</bk-button>
            </div>
          )
        },
      },
      {
        label: '标签'
      },
      {
        label: '业务类型',
        render: ({ data }: { data: ISchemeListItem }) => {
          return data.biz_type
        }
      },
      {
        label: '用户分布地区',
        render: ({ data }: { data: ISchemeListItem }) => {
          return data.user_distribution.map(item => `${item.name}, ${item.children.map(ch => ch.name).join(', ')}`).join('; ')
        }
      },
      {
        label: '部署架构',
        render: ({ data }: { data: ISchemeListItem }) => {
          return data.deployment_architecture.join(', ')
        }
      },
      {
        label: '云厂商',
        render: ({ data }: { data: ISchemeListItem }) => {
          return <div class="verdors-list">{ data.vendors.map(item => <CloudServiceTag type={item} />) }</div>
        }
      },
      {
        label: '综合评分',
        render: ({ data }: { data: ISchemeListItem }) => {
          return <span style={{ color: getScoreColor(data.composite_score) }}>{data.composite_score || '-'}</span>
        }
      },
      {
        label: '创建人',
        render: ({ data }: { data: ISchemeListItem }) => {
          return data.creator
        }
      },
      {
        label: '更新时间',
        render: ({ data }: { data: ISchemeListItem }) => {
          return data.updated_at
        }
      },
      {
        label: '操作',
        render: ({ data }: { data: ISchemeListItem }) => {
          return <bk-button text theme="primary" onClick={() => handleDelScheme(data)}>删除</bk-button>
        }
      },
    ]

    // 加载表格当前页数据
    const getTableListData = async () => {
      tableListLoading.value = true;
      const [collectionRes, allUnCollectedRes] = await Promise.all([
        schemeStore.listCollection(),
        getUnCollectedScheme([], { start: 0, limit: 0, count: true })
      ]);
      collections.value = collectionRes.data.map((item: ICollectedSchemeItem) => ({ id: item.id, res_id: item.res_id }));
      const collectionIds = collections.value.map(item => item.res_id);
      pagination.count = allUnCollectedRes.data.count;

      const currentPageStartNum = (pagination.current - 1) * pagination.limit;
      const currentPageCollectedIdsLength = collectionIds.length - currentPageStartNum;

      if (currentPageCollectedIdsLength > 0 && currentPageCollectedIdsLength < pagination.limit) {
        // 当前页中收藏方案和非收藏方案混排
        const ids = collectionIds.slice(currentPageStartNum);
        const [collectedRes, unCollectedRes] = await Promise.all([
          getCollectedSchemes(ids),
          getUnCollectedScheme(collectionIds, { start: 0, limit: pagination.limit - ids.length })
        ]);
        tableListData.value = [...collectedRes.data.details, ...unCollectedRes.data.details];
      } else if (currentPageCollectedIdsLength >= pagination.limit) {
        // 当前页中只有收藏方案
        const ids = collectionIds.slice(currentPageStartNum, currentPageStartNum + pagination.limit);
        const res = await getCollectedSchemes(ids);
        tableListData.value = res.data.details;
      } else {
        // 当前页中只有非收藏方案
        const res = await getUnCollectedScheme(collectionIds, { start: currentPageStartNum - collectionIds.length, limit: pagination.limit })
        tableListData.value = res.data.details;
      }

      tableListLoading.value = false;
    }

    // 获取已收藏方案列表
    const getCollectedSchemes = (ids: string[]) => {
        const filterQuery: QueryFilterType = {
          op: 'and',
          rules: [{ field: 'id', op: QueryRuleOPEnum.IN, value: ids }]
        };
        const pageQuery: IPageQuery = {
          start: 0,
          limit: ids.length
        };

        return schemeStore.listCloudSelectionScheme(filterQuery, pageQuery);
    }

    // 获取未被收藏的方案列表
    const getUnCollectedScheme = (ids: string[], pageQuery: IPageQuery) => {
      const filterQuery: QueryFilterType = {
        op: 'and',
        rules: []
      };

      if (ids.length > 0) {
        filterQuery.rules.push({ field: 'id', op: QueryRuleOPEnum.NIN, value: ids })
      }

      return schemeStore.listCloudSelectionScheme(filterQuery, pageQuery)
    };

    // 跳转创建方案
    const goToCreate = () => {
      router.push({ name: 'scheme-recommendation' });
    };

    // 跳转方案详情
    const goToDetail = (id: string) => {
      router.push({ name: 'scheme-detail', query: { sid: id } })
    }

    // 搜索方案
    const handleSearch = () => {};

    // 收藏/取消收藏
    const handleToggleCollection = async(scheme: ISchemeListItem) => {
      if (collectPending.value) {
        return;
      }

      collectPending.value = true;
      const index = collections.value.findIndex(item => item.res_id === scheme.id);
      if (index > -1) {
        await schemeStore.deleteCollection(collections.value[index].id);
        collections.value.splice(index, 1);
        Message({
          theme: 'success',
          message: '取消收藏成功',
        });
      } else {
        await schemeStore.createCollection(scheme.id);
        collections.value.push({ id: 0, res_id: scheme.id }); // @todo 收藏成功后, 需要后台接口返回收藏ID
        Message({
          theme: 'success',
          message: '收藏成功',
        });
      }
      collectPending.value = false;

    };

    // 删除方案
    const handleDelScheme = (scheme: ISchemeListItem) => {
      console.log(scheme);
    };

    const handlePageValueChange = (val: number) => {
      console.log('page change', val);
      pagination.current = val;
      getTableListData();
    };

    const handlePageLimitChange = (val: number) => {
      console.log('page limit change', val);
      pagination.current = 1;
      pagination.limit = val;
      getTableListData();
    }

    const handleColumnSort = (val: string) => {
      console.log('col sort', val)
    }

    onMounted(async() => {
      getTableListData();
    });

    return () => (
      <div class="scheme-list-page">
        <div class="operate-wrapper">
          <bk-button class="create-btn" theme="primary" onClick={goToCreate}>
            <Plus class="plus-icon" />
            创建部署方案
          </bk-button>
          <SearchInput v-model={searchStr.value} width={400} onSearch={handleSearch} />
        </div>
        <div class="scheme-table-wrapper">
          <bk-table
            data={tableListData.value}
            pagination={pagination}
            remote-pagination
            pagination-height={60}
            border={['outer']}
            columns={tableCols}
            onPageValueChange={handlePageValueChange}
            onPageLimitChange={handlePageLimitChange}
            onColumnSort={handleColumnSort}>
          </bk-table>
        </div>
      </div>
    );
  },
});
