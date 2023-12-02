import { Exception, Select } from 'bkui-vue';
import { defineComponent, ref, watch } from 'vue';
import './index.scss';
import { Info } from 'bkui-vue/lib/icon';
import SchemePreviewTableCard from './components/scheme-preview-table-card';
import { useSchemeStore } from '@/store';

const { Option } = Select;

export const SchemeSortOptions = [
  {
    key: 'composite_score',
    val: '按综合评分排序',
  },
  {
    key: 'net_score',
    val: '按网络评分排序',
  },
  {
    key: 'cost_score',
    val: '按方案成本排序',
  },
];

export default defineComponent({
  props: {
    onViewDetail: {
      required: true,
      type: Function,
    },
  },
  setup(props) {
    const sortChoice = ref(SchemeSortOptions[0].key);
    const schemeStore = useSchemeStore();

    watch(
      () => sortChoice.value,
      (choice) => {
        schemeStore.sortSchemes(choice);
      },
      {
        deep: true,
      },
    );

    return () => (
      <div class={'scheme-preview-container'}>
        <div class={'scheme-preview-header'}>
          <div class={'scheme-preview-header-title'}>推荐方案</div>
          <Info
            class={'scheme-preview-header-tip'}
            v-bk-tooltips={{
              content: '待产品补充',
            }}
          />
          <Select
            class={'scheme-preivew-header-sort-selector'}
            v-model={sortChoice.value}
            clearable={false}>
            {{
              default: () => SchemeSortOptions.map(({ key, val }) => (
                  <Option value={key} label={val}></Option>
              )),
            }}
          </Select>
        </div>
        <div class={'scheme-preview-content'}>
          {schemeStore.recommendationSchemes.length > 0
            ? schemeStore.recommendationSchemes.map((
              { composite_score, cost_score, net_score, result_idc_ids, cover_rate },
              idx,
            ) => (
                  <SchemePreviewTableCard
                    compositeScore={composite_score}
                    costScore={cost_score}
                    netScore={net_score}
                    resultIdcIds={result_idc_ids}
                    idx={idx}
                    onViewDetail={props.onViewDetail}
                    coverRate={cover_rate}
                  />
            ))
            : (
              <Exception
                type="empty"
                scene="part"
                description="没有数据"
              ></Exception>
            )}
        </div>
      </div>
    );
  },
});
