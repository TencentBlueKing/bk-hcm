import { defineComponent, watch } from 'vue';
import { Select } from 'bkui-vue';
import { QueryRuleOPEnum } from '@/typings';
import { useSingleList } from '@/hooks/useSingleList';
import { getBusinessApiPath } from '@/utils';
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
    const { dataList, isDataLoad, handleScrollEnd, handleRefresh } = useSingleList({
      url: `/api/v1/cloud/${getBusinessApiPath()}/vpcs/list`,
      rules: () => [
        { op: QueryRuleOPEnum.EQ, field: 'account_id', value: props.accountId },
        { op: QueryRuleOPEnum.EQ, field: 'region', value: props.region },
      ],
    });

    const handleChange = (v: string) => {
      const vpcDetail = dataList.value.find((vpc) => vpc.cloud_id === v);
      emit('change', vpcDetail);
      emit('update:modelValue', v);
    };

    const handleClear = () => {
      emit('update:modelValue', '');
    };

    watch(
      [() => props.modelValue, () => props.region],
      async ([vpcId, region]) => {
        if (!region) return;
        await handleRefresh();
        if (vpcId) {
          handleChange(vpcId);
        }
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class='region-vpc-selector'>
        <Select
          modelValue={props.modelValue}
          onChange={handleChange}
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
