import { Button, Select } from 'bkui-vue';
import { defineComponent, ref } from 'vue';
import './index.scss';
import { Info } from 'bkui-vue/lib/icon';
import SchemePreviewTableCard from './components/scheme-preview-table-card';

const { Option } = Select;

export const SchemeSortOptions = [
  {
    key: 1,
    val: '按综合评分排序',
  },
  {
    key: 2,
    val: '按网络评分排序',
  },
  {
    key: 3,
    val: '按方案成本排序',
  },
];

export default defineComponent({
  setup() {
    const sortChoice = ref(SchemeSortOptions[0].key);
    return () => <div class={'scheme-preview-container'}>
      <div class={'scheme-preview-header'}>
        <div class={'scheme-preview-header-title'}>
          推荐方案
        </div>
        <Info
          class={'scheme-preview-header-tip'}
          v-bk-tooltips={{
            content: '待产品补充',
          }}
        />
        <Select
          class={'scheme-preivew-header-sort-selector'}
          v-model={sortChoice.value}
          clearable={false}
        >
          {{
            default: () => (SchemeSortOptions.map(({ key, val }) => (
              <Option
                value={key}
                label={val}
              ></Option>
            ))),
            suffix: () => (<Button>
              123
            </Button>),
          }}
        </Select>
      </div>
      <div class={'scheme-preview-content'}>
        {
          [1, 2, 3].map(_v => <SchemePreviewTableCard></SchemePreviewTableCard>)
        }
      </div>
    </div>;
  },
});
