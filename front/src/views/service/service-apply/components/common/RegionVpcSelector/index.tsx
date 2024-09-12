import { PropType, defineComponent, ref, watch } from 'vue';
import { Select } from 'bkui-vue';
import { QueryRuleOPEnum } from '@/typings';
import { useSingleList } from '@/hooks/useSingleList';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { VendorEnum } from '@/common/constant';

const { Option } = Select;

export default defineComponent({
  name: 'RegionVpcSelector',
  props: {
    modelValue: String, // 选中的vpc cloud_id
    accountId: String, // 云账号id
    vendor: String as PropType<VendorEnum>, // 云厂商, 如果传递此参数, 则使用 /vpcs/with/subnet_count/list, 且更换option显示的内容
    region: String, // 云地域
    isDisabled: { type: Boolean, default: false },
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit, expose }) {
    const cloudVpcId = ref('');

    const { getBusinessApiPath } = useWhereAmI();
    const businessApiPath = getBusinessApiPath();
    const { dataList, isDataLoad, handleScrollEnd, handleRefresh } = useSingleList({
      url: () =>
        props.vendor
          ? `/api/v1/web/${businessApiPath}vendors/${props.vendor}/vpcs/with/subnet_count/list`
          : `/api/v1/cloud/${businessApiPath}vpcs/list`,
      rules: () => [
        { op: QueryRuleOPEnum.EQ, field: 'account_id', value: props.accountId },
        { op: QueryRuleOPEnum.EQ, field: 'region', value: props.region },
      ],
    });

    // clear-handler - 清空数据
    const handleClear = () => {
      cloudVpcId.value = '';
      emit('update:modelValue', '');
    };

    // change-handler - cloudVpcId 变更时, 通过 cloud_id 找到对应的 vpc, 传递给父组件
    const handleChange = () => {
      const vpcDetail = dataList.value.find((vpc) => vpc.cloud_id === cloudVpcId.value);
      emit('change', vpcDetail);
    };

    // region 变更时, 刷新 vpc 列表
    watch(
      () => props.region,
      (val) => {
        if (!val) return;
        handleRefresh();
      },
    );

    // 监听 vpc 列表, 如果父组件传递了 props, 则选中, 防止编辑操作回显数据时丢失
    watch(
      dataList,
      () => {
        cloudVpcId.value = props.modelValue ? props.modelValue : '';
        handleChange();
      },
      { deep: true },
    );

    watch(cloudVpcId, (val) => {
      handleChange();
      emit('update:modelValue', val);
    });

    watch(
      () => props.modelValue,
      (val) => {
        cloudVpcId.value = val;
      },
    );

    // 暴露方法给父组件, 用于数据刷新
    expose({ handleRefresh });

    return () => (
      <div class='region-vpc-selector'>
        <Select
          v-model={cloudVpcId.value}
          onClear={handleClear}
          onScroll-end={handleScrollEnd}
          loading={isDataLoad.value}
          scrollLoading={isDataLoad.value}
          disabled={props.isDisabled}>
          {dataList.value.map(({ id, name, cloud_id, extension }) => {
            if (props.vendor) {
              const cidrs = extension?.cidr?.map((obj: any) => obj.cidr).join(',') || '';
              return <Option key={id} id={cloud_id} name={`${cloud_id} ${name} ${cidrs}`} />;
            }
            return <Option key={id} id={cloud_id} name={`${cloud_id} ${name}`} />;
          })}
        </Select>
      </div>
    );
  },
});
