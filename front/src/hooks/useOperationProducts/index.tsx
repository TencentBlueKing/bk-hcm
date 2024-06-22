import useBillStore from '@/store/useBillStore';
import { Select } from 'bkui-vue';
import { defineComponent, onMounted, reactive, ref, watch } from 'vue';
const { Option } = Select;

export const useOperationProducts = () => {
  const billStore = useBillStore();
  const list = ref([]);
  const pagination = reactive({
    limit: 50,
    start: 0,
  });
  const allCounts = ref(0);

  const getList = async () => {
    const [detailRes, countRes] = await Promise.all(
      [false, true].map((isCount) =>
        billStore.list_operation_products({
          page: {
            ...pagination,
            limit: isCount ? 0 : pagination.limit,
            count: isCount,
          },
        }),
      ),
    );
    allCounts.value = countRes.data.count;
    list.value = detailRes.data.details;
  };

  onMounted(() => {
    getList();
  });

  const OperationProductsSelector = defineComponent({
    props: {
      modelValue: String,
    },
    setup(props, { emit }) {
      const selectedVal = ref(props.modelValue);
      const isScrollLoading = ref(false);

      watch(
        () => selectedVal.value,
        (val) => {
          emit('update:modelValue', val);
        },
      );

      return () => (
        <div class={'selector-wrapper'}>
          <Select
            v-model={selectedVal.value}
            scrollLoading={isScrollLoading.value}
            filterable
            onScroll-end={async () => {
              if (list.value.length >= allCounts.value || isScrollLoading.value) return;
              isScrollLoading.value = true;
              pagination.start += pagination.limit;
              const { data } = await billStore.list_operation_products({
                page: {
                  start: pagination.start,
                  count: false,
                  limit: pagination.limit,
                },
              });
              list.value.push(...data.details);
              isScrollLoading.value = false;
            }}>
            {list.value.map(({ op_product_name, op_product_id }) => (
              <Option name={op_product_name} id={op_product_id} key={op_product_id} />
            ))}
          </Select>
        </div>
      );
    },
  });

  return {
    OperationProductsSelector,
  };
};
