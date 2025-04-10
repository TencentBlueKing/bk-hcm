import { PropType, computed, defineComponent, watch } from 'vue';
import { Select } from 'bkui-vue';
import { QueryRuleOPEnum } from '@/typings';
import { useSingleList } from '@/hooks/useSingleList';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { VendorEnum } from '@/common/constant';

const { Option } = Select;

interface IVpcItem {
  id: string;
  vendor: string;
  account_id: string;
  cloud_id: string;
  name: string;
  region: string;
  category: string;
  memo: string;
  bk_biz_id: number;
  bk_cloud_id: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

interface IVpcWithSubnetCountItem extends IVpcItem {
  extension?: any; // 具体见 docs\api-docs\web-server\docs\resource\list_vpc_with_subnet_count.md
  subnet_count: number;
  current_zone_subnet_count: number;
}

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
    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        const vpcDetail = dataList.value.find((vpc) => vpc.cloud_id === val);
        emit('change', vpcDetail);
        emit('update:modelValue', val);
      },
    });

    const { getBusinessApiPath } = useWhereAmI();
    const businessApiPath = getBusinessApiPath();
    const { dataList, isDataLoad, handleRefresh } = useSingleList<IVpcWithSubnetCountItem>({
      url: () =>
        props.vendor
          ? `/api/v1/web/${businessApiPath}vendors/${props.vendor}/vpcs/with/subnet_count/list`
          : `/api/v1/cloud/${businessApiPath}vpcs/list`,
      rules: () => [
        { op: QueryRuleOPEnum.EQ, field: 'account_id', value: props.accountId },
        { op: QueryRuleOPEnum.EQ, field: 'region', value: props.region },
      ],
      rollRequestConfig: { enabled: true, limit: 50 },
    });

    // clear-handler - 清空数据
    const handleClear = () => {
      selected.value = '';
    };

    // region 变更时, 刷新 vpc 列表
    watch(
      () => props.region,
      (val) => {
        if (!val) return;
        handleRefresh();
      },
    );

    watch(
      () => props.modelValue,
      (val) => {
        selected.value = val;
      },
    );

    // 暴露方法给父组件, 用于数据刷新
    expose({ handleRefresh });

    return () => (
      <div class='region-vpc-selector'>
        <Select v-model={selected.value} onClear={handleClear} loading={isDataLoad.value} disabled={props.isDisabled}>
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
