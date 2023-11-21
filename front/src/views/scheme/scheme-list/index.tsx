import { defineComponent, ref, reactive } from "vue";
import { useRouter } from 'vue-router';
import { Plus } from "bkui-vue/lib/icon";
import SearchInput from "../components/search-input/index";

import './index.scss';

export default defineComponent({
  name: 'scheme-list-page',
  setup () {

    const router = useRouter();

    const searchStr = ref('');
    const pagination = reactive({
        location: 'left',
        align: 'right',
        start: 1,
        limit: 10,
        count: 0,
    });

    const tableCols = [
      {
        label: '方案名称'
      },
      {
        label: '标签'
      },
      {
        label: '业务类型'
      },
      {
        label: '用户分布地区'
      },
      {
        label: '部署架构'
      },
      {
        label: '云厂商'
      },
      {
        label: '综合评分'
      },
      {
        label: '创建人'
      },
      {
        label: '更新时间'
      },
      {
        label: '操作'
      },
    ]

    const goToCreate = () => {
      router.push({ name: 'scheme-recommendation' });
    };

    const handleSearch = () => {};

    return () => (
      <div class="scheme-list-page">
        <div class="operate-wrapper">
          <bk-button class="create-btn" theme="primary" onClick={goToCreate}>
            <Plus class="plus-icon" />
            创建部署方案
          </bk-button>
          <SearchInput v-model={searchStr.value} width={400} onSearch={handleSearch} />
        </div>
        <div class="scheme-table-rapper">
          <bk-table
            pagination={pagination}
            pagination-height={60}
            border={['outer']}
            columns={tableCols} />
        </div>
      </div>
    )
  },
});
