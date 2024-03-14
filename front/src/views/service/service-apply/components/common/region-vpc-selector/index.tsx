import { defineComponent, reactive, ref, watch } from 'vue';
// import components
import { Select } from 'bkui-vue';
// import stores
import { useResourceStore } from '@/store';
// import types
import { IPageQuery, QueryRuleOPEnum } from '@/typings';
// import utils
import { throttle } from 'lodash';
import './index.scss';

const { Option } = Select;

export default defineComponent({
  name: 'RegionVpcSelector',
  props: {
    modelValue: String, // 选中的vpc cloud_id
    accountId: String, // 云账号id
    region: String, // 云地域
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit }) {
    // use stores
    const resourceStore = useResourceStore();
    // define data
    const loading = ref(false);
    const selectedValue = ref('');
    const vpcListPage = reactive<IPageQuery>({
      start: 0,
      limit: 7,
    });
    const hasMoreData = ref(true);
    const vpcList = ref([]);

    // 获取vpc列表
    const getVpcList = async (region: string) => {
      loading.value = true;
      try {
        const [listRes, countRes] = await Promise.all(
          [false, true].map((isCount) =>
            resourceStore.list(
              {
                page: {
                  count: isCount,
                  start: isCount ? 0 : vpcListPage.start,
                  limit: isCount ? 0 : vpcListPage.limit,
                  sort: isCount ? null : 'created_at',
                  order: isCount ? null : 'DESC',
                },
                filter: {
                  op: QueryRuleOPEnum.AND,
                  rules: [
                    { op: QueryRuleOPEnum.EQ, field: 'account_id', value: props.accountId },
                    { op: QueryRuleOPEnum.EQ, field: 'region', value: region },
                  ],
                },
              },
              'vpcs',
            ),
          ),
        );
        vpcList.value = [...vpcList.value, ...listRes.data.details];
        if (vpcList.value.length < countRes.data.count) {
          hasMoreData.value = true;
        } else {
          hasMoreData.value = false;
        }
      } finally {
        loading.value = false;
      }
    };

    // 清空选项
    const handleClear = () => {
      selectedValue.value = '';
    };

    // 滚动加载
    const handleScrollEnd = throttle(() => {
      if (hasMoreData.value) {
        vpcListPage.start += vpcListPage.limit;
        getVpcList(props.region);
      }
    }, 500);

    watch(
      () => props.region,
      (val) => {
        selectedValue.value = '';
        if (!val) {
          vpcList.value = [];
          return;
        }
        // 当云地域变更时, 获取新的vpc列表
        getVpcList(val);
      },
    );

    watch(selectedValue, (val) => {
      const vpcDetail = vpcList.value.find((vpc) => vpc.cloud_id === val);
      emit('update:modelValue', val);
      emit('change', vpcDetail);
    });

    return () => (
      <div class='region-vpc-selector'>
        <Select
          v-model={selectedValue.value}
          scrollLoading={loading.value}
          onClear={handleClear}
          onScroll-end={handleScrollEnd}>
          {vpcList.value.map(({ id, name, cloud_id }) => (
            <Option key={id} id={cloud_id} name={`${cloud_id} ${name}`} />
          ))}
        </Select>
      </div>
    );
  },
});
