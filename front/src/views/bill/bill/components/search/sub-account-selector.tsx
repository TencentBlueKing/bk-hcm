import { VendorEnum } from '@/common/constant';
import { BILL_MAIN_ACCOUNTS_KEY } from '@/constants';
import { useSingleList } from '@/hooks/useSingleList';
import { QueryRuleOPEnum } from '@/typings';
import { Select } from 'bkui-vue';
import { isEqual } from 'lodash';
import { PropType, computed, defineComponent, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    vendor: Array as PropType<VendorEnum[]>,
    rootAccountId: Array as PropType<string[]>,
    autoSelect: Boolean,
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const router = useRouter();
    const route = useRoute();

    const selectedValue = ref(props.modelValue);
    const rules = computed(() => {
      const rules = [];
      if (props.vendor.length) rules.push({ field: 'vendor', op: QueryRuleOPEnum.IN, value: props.vendor });
      if (props.rootAccountId.length)
        rules.push({ field: 'parent_account_id', op: QueryRuleOPEnum.IN, value: props.rootAccountId });
      return rules;
    });

    const { dataList, isDataLoad, handleScrollEnd, handleRefresh } = useSingleList({
      url: '/api/v1/account/main_accounts/list',
      rules: () => rules.value,
      pagination: { limit: 10000 },
      immediate: true,
    });

    watch(
      () => props.modelValue,
      (val) => {
        selectedValue.value = val;
      },
    );

    watch(selectedValue, (val) => {
      router.push({
        query: { ...route.query, [BILL_MAIN_ACCOUNTS_KEY]: val.length ? btoa(JSON.stringify(val)) : undefined },
      });
      emit('update:modelValue', val);
    });

    watch(
      rules,
      (newRules, oldRules) => {
        !isEqual(newRules, oldRules) && handleRefresh();
      },
      { deep: true },
    );

    // 二级账号数据只会拉取一次, 所以监听第一次改变就行, 监听的目的是为了自动选中
    const unwatch = watch(
      dataList,
      () => {
        if (props.autoSelect && route.query[BILL_MAIN_ACCOUNTS_KEY]) {
          selectedValue.value = JSON.parse(atob(route.query[BILL_MAIN_ACCOUNTS_KEY] as string));
          unwatch();
        }
      },
      { deep: true },
    );

    return () => (
      <Select
        v-model={selectedValue.value}
        multiple
        multipleMode='tag'
        collapseTags
        clearable
        onScroll-end={handleScrollEnd}
        loading={isDataLoad.value}
        scrollLoading={isDataLoad.value}>
        {dataList.value.map(({ id, cloud_id }) => (
          <Select.Option key={id} id={id} name={cloud_id} />
        ))}
      </Select>
    );
  },
});
