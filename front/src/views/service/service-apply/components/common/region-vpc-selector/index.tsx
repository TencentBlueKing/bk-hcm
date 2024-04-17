import { defineComponent, ref, watch } from 'vue';
// import components
import { Select } from 'bkui-vue';
// import types
import { QueryRuleOPEnum } from '@/typings';
// import hooks
import useSelectOptionListWithScroll from '@/hooks/useSelectOptionListWithScroll';
import './index.scss';

const { Option } = Select;

export default defineComponent({
  name: 'RegionVpcSelector',
  props: {
    modelValue: String, // 选中的vpc cloud_id
    accountId: String, // 云账号id
    region: String, // 云地域
    isDisabled: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit }) {
    const selectedValue = ref('');

    const [isLoading, optionList, initState, getOptionList, handleOptionListScrollEnd] = useSelectOptionListWithScroll(
      'vpcs',
      [],
      false,
    );

    // 清空选项
    const handleClear = () => {
      selectedValue.value = '';
    };

    watch(
      () => props.region,
      async (val) => {
        if (!val) return;
        // 初始化状态
        selectedValue.value = '';
        initState();
        // 当云地域变更时, 获取新的vpc列表
        await getOptionList([
          { op: QueryRuleOPEnum.EQ, field: 'account_id', value: props.accountId },
          { op: QueryRuleOPEnum.EQ, field: 'region', value: val },
        ]);

        // 如果为edit, 需要回填
        if (props.modelValue) selectedValue.value = props.modelValue;
      },
      {
        immediate: true,
      },
    );

    watch(selectedValue, (val) => {
      // 更新父组件中的数据cloud_vpc_id
      emit('update:modelValue', val);
      // 将选中的vpc信息回传给父组件
      const vpcDetail = optionList.value.find((vpc) => vpc.cloud_id === val);
      emit('change', vpcDetail);
    });

    return () => (
      <div class='region-vpc-selector'>
        <Select
          v-model={selectedValue.value}
          scrollLoading={isLoading.value}
          onClear={handleClear}
          onScroll-end={handleOptionListScrollEnd}
          disabled={props.isDisabled}>
          {optionList.value.map(({ id, name, cloud_id }) => (
            <Option key={id} id={cloud_id} name={`${cloud_id} ${name}`} />
          ))}
        </Select>
      </div>
    );
  },
});
