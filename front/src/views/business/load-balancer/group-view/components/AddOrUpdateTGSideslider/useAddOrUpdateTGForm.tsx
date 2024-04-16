import { ref, computed, watch, Ref } from 'vue';
// import components
import { Input, Select } from 'bkui-vue';
import AccountSelector from '@/components/account-selector/index.vue';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector';
import RegionVpcSelector from '@/views/service/service-apply/components/common/region-vpc-selector';
import RsConfigTable from '../RsConfigTable';
// import stores
import { useAccountStore, useLoadBalancerStore } from '@/store';
// import types and constants
import { TARGET_GROUP_PROTOCOLS, VendorEnum } from '@/common/constant';

const { Option } = Select;

export default (formData: any, updateCount: Ref<number>) => {
  // use stores
  const accountStore = useAccountStore();
  const loadBalancerStore = useLoadBalancerStore();

  const curVendor = ref(VendorEnum.TCLOUD);
  const curVpcId = ref('');

  const selectedBizId = computed({
    get() {
      return accountStore.bizs;
    },
    set(val) {
      formData.bk_biz_id = val;
    },
  });

  const disabledEdit = computed(() => updateCount.value === 2 && loadBalancerStore.currentScene !== 'edit');

  const formItemOptions = computed(() => [
    {
      label: '云账号',
      required: true,
      property: 'account_id',
      span: 12,
      content: () => (
        <AccountSelector
          v-model={formData.account_id}
          bizId={selectedBizId.value}
          type='resource'
          onChange={(account: { vendor: VendorEnum }) => (curVendor.value = account?.vendor)}
          disabled={disabledEdit.value}
        />
      ),
    },
    [
      {
        label: '目标组名称',
        required: true,
        property: 'name',
        span: 12,
        content: () => <Input v-model={formData.name} disabled={disabledEdit.value} />,
      },
      {
        label: '协议端口',
        required: true,
        span: 12,
        content: () => (
          <div class='flex-row'>
            <Select v-model={formData.protocol} disabled={disabledEdit.value}>
              {TARGET_GROUP_PROTOCOLS.map((protocol) => (
                <Option name={protocol} id={protocol} />
              ))}
            </Select>
            &nbsp;&nbsp;:&nbsp;&nbsp;
            <Input v-model={formData.port} disabled={disabledEdit.value} />
          </div>
        ),
      },
    ],
    [
      {
        label: '地域',
        required: true,
        property: 'region',
        span: 12,
        content: () => (
          <RegionSelector
            isDisabled={!formData.account_id || disabledEdit.value}
            v-model={formData.region}
            accountId={formData.account_id}
            vendor={curVendor.value}
            type='cvm'
          />
        ),
      },
      {
        label: '所属VPC',
        required: true,
        property: 'cloud_vpc_id',
        span: 12,
        content: () => (
          <RegionVpcSelector
            v-model={formData.cloud_vpc_id}
            accountId={formData.account_id}
            region={formData.region}
            isDisabled={(!formData.account_id && !formData.region) || disabledEdit.value}
            onChange={(vpcDetail: any) => (curVpcId.value = vpcDetail?.id || '')}
          />
        ),
      },
    ],
    {
      label: 'RS 配置',
      required: true,
      property: 'rs_list',
      span: 24,
      content: () => (
        <RsConfigTable v-model:rsList={formData.rs_list} accountId={formData.account_id} vpcId={curVpcId.value} />
      ),
    },
  ]);

  watch(
    () => formData.account_id,
    (v) => {
      !v && (formData.region = '');
    },
  );

  watch(
    () => formData.region,
    (v) => {
      !v && (formData.cloud_vpc_id = '');
    },
  );

  return {
    formItemOptions,
  };
};
