import { VendorEnum } from '@/common/constant';
import { useSingleList } from '@/hooks/useSingleList';
import { QueryRuleOPEnum } from '@/typings';
import { decodeValueByAtob, encodeValueByBtoa } from '@/utils';
import { SelectColumn } from '@blueking/ediatable';
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
    // 保存至 url 上的 key
    urlKey: String,
    // 是否用于 ediatable
    isEditable: { type: Boolean, default: false },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, expose }) {
    const selectedValue = ref(props.modelValue);
    const selectRef = ref();
    const router = useRouter();
    const route = useRoute();

    const rules = computed(() => {
      const rules = [];
      if (props.vendor?.length) rules.push({ field: 'vendor', op: QueryRuleOPEnum.IN, value: props.vendor });
      if (props.rootAccountId?.length)
        rules.push({ field: 'parent_account_id', op: QueryRuleOPEnum.IN, value: props.rootAccountId });
      return rules;
    });

    const { dataList, isDataLoad, handleScrollEnd, handleRefresh } = useSingleList({
      url: '/api/v1/account/main_accounts/list',
      rules: () => rules.value,
      pagination: { limit: 10000 },
      immediate: true,
    });

    const getValue = () => {
      return selectRef.value.getValue().then(() => selectedValue.value);
    };

    expose({
      getValue,
    });

    watch(
      () => props.modelValue,
      (val) => (selectedValue.value = val),
      { deep: true },
    );

    watch(selectedValue, async (val) => {
      // async/await 避免因异步路由跳转导致取值错误
      if (props.urlKey) {
        await router.push({
          query: { ...route.query, [props.urlKey]: val.length ? encodeValueByBtoa(val) : undefined },
        });
      }
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
        if (props.autoSelect && props.urlKey && route.query[props.urlKey]) {
          selectedValue.value = decodeValueByAtob(route.query[props.urlKey] as string);
          unwatch();
        }
      },
      { deep: true },
    );

    if (props.isEditable) {
      return () => (
        <SelectColumn
          onScroll-end={handleScrollEnd}
          loading={isDataLoad.value}
          scrollLoading={isDataLoad.value}
          v-model={selectedValue.value}
          list={dataList.value.map(({ id, cloud_id }) => ({
            label: cloud_id as string,
            value: id as string,
            key: id as string,
          }))}
          ref={selectRef}
          rules={[
            {
              validator: (value: string) => Boolean(value),
              message: '二级账号不能为空',
            },
          ]}
        />
      );
    }

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
