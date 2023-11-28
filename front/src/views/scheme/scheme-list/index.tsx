import { defineComponent, ref, reactive, computed, onMounted } from "vue";
import { useRouter } from 'vue-router';
import { Plus } from "bkui-vue/lib/icon";
import { useSchemeStore } from '@/store';
import { QueryFilterType, IPageQuery } from '@/typings/common';
import { ICollectedSchemeItem, ISchemeListItem } from '@/typings/scheme';
import SearchInput from "../components/search-input/index";

import './index.scss';
import { data } from "autoprefixer";

export default defineComponent({
  name: 'scheme-list-page',
  setup () {
    const schemeStore = useSchemeStore();

    const router = useRouter();

    const searchStr = ref('');
    const collectionLoading = ref(false);
    let collectionList = ref<ICollectedSchemeItem[]>([]);
    let schemeList = ref<ISchemeListItem[]>([]);
    const schemeLoading = ref(false);
    const pagination = reactive({
        location: 'left',
        align: 'right',
        start: 1,
        limit: 10,
        count: 0,
    });

    const tableCols = [
      {
        label: '方案名称',
        render: ({ data }: { data: ISchemeListItem }) => {
          return <bk-button text theme="primary" onClick={() => { goToDetail(data.id) }}>{data.name}</bk-button>
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

    // 加载已收藏方案列表
    const getSchemeCollection = async () => {
      collectionLoading.value = true
      const res = await schemeStore.listCollection();
      collectionList.value = res.data;
      console.log(res);
      collectionLoading.value = false;
    }

    // 加载方案列表
    const getSchemeList = async () => {
      schemeLoading.value = true;
      const filterQuery: QueryFilterType = {
        op: 'and',
        rules: []
      };
      const pageQuery: IPageQuery = {
        start: 0,
        limit: pagination.limit
      };
      const res = await schemeStore.listCloudSelectionScheme(filterQuery, pageQuery);
      console.log(res);
      schemeList.value = res.data.details;
      schemeLoading.value = false;
    }

    const goToCreate = () => {
      router.push({ name: 'scheme-recommendation' });
    };

    const goToDetail = (id: string) => {
      router.push({ name: 'scheme-detail', query: { sid: id } })
    }

    const handleSearch = () => {};

    // 删除方案
    const handleDelScheme = (scheme: ISchemeListItem) => {
      console.log(scheme);
    };

    onMounted(() => {
      getSchemeCollection();
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
            pagination-height={60}
            border={['outer']}
            columns={tableCols} />
        </div>
      </div>
    );
  },
});
