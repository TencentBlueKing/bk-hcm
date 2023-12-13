import { Button, Exception, Select } from 'bkui-vue';
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
    val: '按成本评分排序',
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
    const isDes = ref(true);

    watch(
      () => [
        sortChoice.value,
        isDes.value,
      ],
      () => {
        schemeStore.sortSchemes(sortChoice.value, isDes.value);
      },
      {
        deep: true,
        immediate: true,
      },
    );

    return () => (
      <div class={'scheme-preview-container'}>
        <div class={'scheme-preview-header'}>
          <div class={'scheme-preview-header-title'}>推荐方案</div>
          <Info
            class={'scheme-preview-header-tip'}
            v-bk-tooltips={{
              content: '本方案由系统的算法推荐出来，算法所使用的数据是平台的公共数据，暂不支持自定义数据。推荐的依据为业务分布地区，业务的类型，网络延迟，默认的用户分布占比等。',
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
          <Button onClick={() => isDes.value = !isDes.value}  v-bk-tooltips={{
            content: isDes.value ? '降序' : '升序',
          }}>
            <i class={`${isDes.value ? 'hcm-icon bkhcm-icon-jiangxu' : 'icon hcm-icon bkhcm-icon-shengxu'}`}/>
          </Button>
        </div>
        <div class={'scheme-preview-content'}>
          {schemeStore.recommendationSchemes.length > 0
            ? schemeStore.recommendationSchemes.map((
              { composite_score, cost_score, net_score, result_idc_ids, cover_rate, id },
              idx,
            ) => (
                  <SchemePreviewTableCard
                    key={result_idc_ids.join(',') + id}
                    compositeScore={composite_score}
                    costScore={cost_score}
                    netScore={net_score}
                    resultIdcIds={result_idc_ids}
                    idx={idx}
                    onViewDetail={(idx: number) => props.onViewDetail(idx)}
                    coverRate={cover_rate}
                  />
            ))
            : (
              <Exception
                type="search-empty"
                scene="page"
                description="暂无推荐结果"
              ></Exception>
            )}
        </div>
      </div>
    );
  },
});
