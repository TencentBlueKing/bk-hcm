import { defineComponent, ref, watch } from 'vue';
// import components
import { Select } from 'bkui-vue';
// import types
import { QueryRuleOPEnum } from '@/typings';
// import hooks
import { useSingleList } from '@/hooks/useSingleList';
import './index.scss';

const { Option } = Select;

export default defineComponent({
  name: 'RegionVpcSelector',
  props: {
    modelValue: String, // 选中的vpc cloud_id
    accountId: String, // 云账号id
    region: String, // 云地域
    isDisabled: { type: Boolean, default: false },
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit }) {
    const selectedValue = ref('');

    const { dataList, isDataLoad, handleScrollEnd, handleRefresh } = useSingleList('vpcs', {
      rules: () => [
        { op: QueryRuleOPEnum.EQ, field: 'account_id', value: props.accountId },
        { op: QueryRuleOPEnum.EQ, field: 'region', value: props.region },
      ],
    });

    // 清空选项
    const handleClear = () => {
      selectedValue.value = '';
    };

    // 当地域变更时, 刷新列表
    watch(() => props.region, handleRefresh);

    watch(selectedValue, (val) => {
      // 更新父组件中的数据cloud_vpc_id
      emit('update:modelValue', val);
      // 将选中的vpc信息回传给父组件
      const vpcDetail = dataList.value.find((vpc) => vpc.cloud_id === val);
      emit('change', vpcDetail);
    });

    return () => (
      <div class='region-vpc-selector'>
        <Select
          v-model={selectedValue.value}
          onClear={handleClear}
          onScroll-end={handleScrollEnd}
          loading={isDataLoad.value}
          scrollLoading={isDataLoad.value}
          disabled={props.isDisabled}>
          {dataList.value.map(({ id, name, cloud_id }) => (
            <Option key={id} id={cloud_id} name={`${cloud_id} ${name}`} />
          ))}
        </Select>
      </div>
    );
  },
});
