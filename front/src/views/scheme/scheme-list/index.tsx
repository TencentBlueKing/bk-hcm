import { defineComponent, ref, reactive, computed, onMounted } from "vue";
import { useRouter } from 'vue-router';
import { Plus } from "bkui-vue/lib/icon";
import { useSchemeStore } from '@/store';
import { QueryFilterType, IPageQuery, QueryRuleOPEnum } from '@/typings/common';
import { ICollectedSchemeItem, ISchemeListItem } from '@/typings/scheme';
import SearchInput from "../components/search-input/index";

import './index.scss';

export default defineComponent({
  name: 'scheme-list-page',
  setup () {
    const schemeStore = useSchemeStore();

    const router = useRouter();

    const searchStr = ref('');
    const collectionLoading = ref(false);
    let collectionIds = reactive<string[]>([]);
    let collectionList = ref<ICollectedSchemeItem[]>([]);
    let schemeList = ref<ISchemeListItem[]>([]);
    const schemeLoading = ref(false);
    const collectPending = ref(false);
    const pagination = reactive({
        current: 1,
        count: 0,
        limit: 10,
    });

    const tableCols = [
      {
        label: '方案名称',
        minWidth: 200,
        render: ({ data }: { data: ISchemeListItem }) => {
          return (
            <div class="scheme-name">
              <i
                class={['hcm-icon', 'collect-icon', collectionIds.includes(data.id) ? 'bkhcm-icon-collect' : 'bkhcm-icon-not-favorited']}
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
          return data.vendors.join(', ')
        }
      },
      {
        label: '综合评分',
        render: ({ data }: { data: ISchemeListItem }) => {
          return data.composite_score
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

    const tableListData = computed(() => {
      // const collectionLen = collectionList.length;
      console.log([...collectionList.value, ...schemeList.value])
      return [...collectionList.value, ...schemeList.value]
    })

    // 加载已收藏方案
    const getSchemeCollection = async () => {
      collectionLoading.value = true
      const colRes = await schemeStore.listCollection();
      collectionIds = colRes.data.map((item: ICollectedSchemeItem) => item.res_id)
      if (collectionIds.length > 0) {
        const filterQuery: QueryFilterType = {
          op: 'and',
          rules: [{ field: 'id', op: QueryRuleOPEnum.IN, value: collectionIds }]
        };
        const pageQuery: IPageQuery = {
          start: 0,
          limit: collectionIds.length
        };
        const schRes = await schemeStore.listCloudSelectionScheme(filterQuery, pageQuery);
        collectionList.value = schRes.data.details;
        console.log(colRes, schRes);
      } else {
        collectionList.value = [];
      }
      collectionLoading.value = false;
    }

    // 加载排除已收藏方案的列表
    const getSchemeList = async () => {
      schemeLoading.value = true;
      const filterQuery: QueryFilterType = {
        op: 'and',
        rules: [{ field: 'id', op: QueryRuleOPEnum.NIN, value: collectionIds }]
      };
      const pageQuery: IPageQuery = {
        start: (pagination.current - 1) * pagination.limit,
        limit: pagination.limit
      };
      const [listRes, countRes] = await Promise.all([
         schemeStore.listCloudSelectionScheme(filterQuery, pageQuery),
         schemeStore.listCloudSelectionScheme(filterQuery, { start: 0, limit: 0, count: true }),
      ])
      schemeList.value = listRes.data.details;
      pagination.count = countRes.data.count;
      schemeLoading.value = false;
    }

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
      if (collectionIds.includes(scheme.id)) {
        await schemeStore.deleteCollection(scheme.id);
      } else {
        await schemeStore.createCollection(scheme.id);
      }
      collectPending.value = true;

    };

    // 删除方案
    const handleDelScheme = (scheme: ISchemeListItem) => {
      console.log(scheme);
    };

    const handlePageValueChange = (val: number) => {
      console.log('page change', val);
    };

    const handlePageLimitChange = (val: number) => {
      console.log('page limit change', val);
    }

    const handleColumnSort = (val: string) => {
      console.log('col sort', val)
    }

    onMounted(async() => {
      await getSchemeCollection();
      getSchemeList();
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
