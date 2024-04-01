import { ref, computed, watch } from 'vue';
// import components
import { Input, Select } from 'bkui-vue';
import AccountSelector from '@/components/account-selector/index.vue';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector';
import VpcSelector from '@/components/vpc-selector/index.vue';
import RsConfigTable from '../RsConfigTable';
// import stores
import { useAccountStore, useBusinessStore } from '@/store';
// import types and constants
import { QueryRuleOPEnum } from '@/typings';
import { TARGET_GROUP_PROTOCOLS, VendorEnum } from '@/common/constant';
// import utils
import bus from '@/common/bus';

const { Option } = Select;

export default (formData: any) => {
  // use stores
  const accountStore = useAccountStore();
  const businessStore = useBusinessStore();

  const curVendor = ref(VendorEnum.TCLOUD);
  const rsTableList = ref([]);

  const selectedBizId = computed({
    get() {
      return accountStore.bizs;
    },
    set(val) {
      formData.bk_biz_id = val;
    },
  });

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
          onChange={(account: { vendor: VendorEnum }) => (curVendor.value = account.vendor)}
        />
      ),
    },
    [
      {
        label: '目标组名称',
        required: true,
        property: 'name',
        span: 12,
        content: () => <Input v-model={formData.name} />,
      },
      {
        label: '协议端口',
        required: true,
        span: 12,
        content: () => (
          <div class='flex-row'>
            <Select v-model={formData.protocol}>
              {TARGET_GROUP_PROTOCOLS.map((protocol) => (
                <Option name={protocol} id={protocol} />
              ))}
            </Select>
            &nbsp;&nbsp;:&nbsp;&nbsp;
            <Input v-model={formData.port} />
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
            isDisabled={!formData.account_id}
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
          <VpcSelector
            v-model={formData.cloud_vpc_id}
            isDisabled={!formData.account_id && !formData.region}
            region={formData.region}
            vendor={curVendor.value}
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
        <RsConfigTable
          rsList={formData.rs_list}
          onShowAddRsDialog={() => bus.$emit('showAddRsDialog', rsTableList.value)}
        />
      ),
    },
  ]);

  // 获取 rs 列表
  const getAllRsList = async (accountId: string) => {
    if (!accountId) return;
    const res = await businessStore.getAllRsList({
      filter: {
        op: QueryRuleOPEnum.AND,
        rules: [
          {
            field: 'account_id',
            op: QueryRuleOPEnum.EQ,
            value: accountId,
          },
        ],
      },
      page: {
        start: 0,
        limit: 500,
      },
    });
    rsTableList.value = res.data.details;
  };

  watch(
    () => formData.account_id,
    (id) => {
      getAllRsList(id);
    },
  );

  return {
    formItemOptions,
  };
};
