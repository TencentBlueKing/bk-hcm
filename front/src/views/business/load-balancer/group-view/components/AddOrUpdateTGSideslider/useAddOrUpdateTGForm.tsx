import { ref, computed, watch, Ref } from 'vue';
// import components
import { Input, Select } from 'bkui-vue';
import AccountSelector from '@/components/account-selector/index.vue';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector';
import RegionVpcSelector from '@/views/service/service-apply/components/common/RegionVpcSelector';
import RsConfigTable from '../RsConfigTable';
// import stores
import { useAccountStore, useLoadBalancerStore } from '@/store';
// import types and constants
import { TARGET_GROUP_PROTOCOLS, VendorEnum } from '@/common/constant';

const { Option } = Select;

export default (formData: any, updateCount: Ref<number>, isEdit: Ref<boolean>, lbDetail: Ref<any>) => {
  // use stores
  const accountStore = useAccountStore();
  const loadBalancerStore = useLoadBalancerStore();

  const curVendor = computed({
    get() {
      return formData.vendor;
    },
    set(val) {
      formData.vendor = val;
    },
  });
  const curVpcId = computed({
    get() {
      return formData.vpc_id;
    },
    set(val) {
      formData.vpc_id = val;
    },
  });
  const deletedRsList = ref([]);

  const selectedBizId = computed({
    get() {
      return accountStore.bizs;
    },
    set(val) {
      formData.bk_biz_id = val;
    },
  });

  const disabledEdit = computed(() => updateCount.value === 2 && loadBalancerStore.currentScene !== 'edit');
  // 只有新增目标组的时候可以修改账号
  const canUpdateAccount = computed(() => loadBalancerStore.currentScene === 'add');
  // 只有新增目标组或目标组没有绑定rs时, 才可以修改地域和vpc
  const canUpdateRegionOrVpc = computed(
    () => loadBalancerStore.currentScene === 'add' || formData.rs_list.filter((item: any) => !item.isNew).length === 0,
  );
  // 表单相关
  const formRef = ref();
  const regionVpcSelectorRef = ref();
  const rules = {
    account_id: [
      {
        required: true,
        message: '云账户不能为空',
        trigger: 'change',
      },
    ],
    name: [
      {
        required: true,
        message: '目标组名称不能为空',
        trigger: 'blur',
      },
    ],
    region: [
      {
        required: true,
        message: '地域不能为空',
        trigger: 'change',
      },
    ],
    cloud_vpc_id: [
      {
        required: true,
        message: 'VPC不能为空',
        trigger: 'change',
      },
    ],
    protocol_port: [
      {
        required: true,
        trigger: 'change',
        message: '协议或端口不能为空',
        validator: () => {
          const { protocol, port } = formData;
          if (!protocol) {
            return false;
          }
          if (!port && port !== 0) {
            return false;
          }
          return true;
        },
      },
      {
        required: true,
        trigger: 'blur',
        message: '端口范围 1-65535 ',
        validator: () => {
          const { port } = formData;
          if (Number.isInteger(port) && port >= 1 && port <= 65535) {
            return true;
          }
          return false;
        },
      },
    ],
  };
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
          disabled={disabledEdit.value || !canUpdateAccount.value}
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
        property: 'protocol_port',
        span: 12,
        content: () => (
          <div class='flex-row'>
            <Select v-model={formData.protocol} disabled={disabledEdit.value}>
              {TARGET_GROUP_PROTOCOLS.map((protocol) => (
                <Option name={protocol} id={protocol} />
              ))}
            </Select>
            &nbsp;&nbsp;:&nbsp;&nbsp;
            <Input
              v-model={formData.port}
              disabled={isEdit.value || disabledEdit.value}
              type='number'
              class='no-number-control'
            />
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
            isDisabled={!formData.account_id || disabledEdit.value || !canUpdateRegionOrVpc.value}
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
            ref={regionVpcSelectorRef}
            v-model={formData.cloud_vpc_id}
            accountId={formData.account_id}
            region={formData.region}
            isDisabled={(!formData.account_id && !formData.region) || disabledEdit.value || !canUpdateRegionOrVpc.value}
            onChange={(vpcDetail: any) => (curVpcId.value = vpcDetail?.id || '')}
          />
        ),
      },
    ],
    {
      label: 'RS 配置',
      property: 'rs_list',
      span: 24,
      content: () => (
        <RsConfigTable
          v-model:rsList={formData.rs_list}
          v-model:deletedRsList={deletedRsList.value}
          accountId={formData.account_id}
          vpcId={curVpcId.value}
          port={formData.port}
          lbDetail={lbDetail.value}
        />
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
    () => {
      // region改变时, 过滤掉新增的rs, 保留原有的rs
      formData.rs_list = formData.rs_list.filter((item: any) => !item.isNew);
      // 重置cloud_vpc_id
      ['add', 'edit'].includes(loadBalancerStore.currentScene) && (formData.cloud_vpc_id = '');
    },
  );

  return {
    rules,
    formRef,
    formItemOptions,
    canUpdateRegionOrVpc,
    deletedRsList,
    regionVpcSelectorRef,
  };
};
