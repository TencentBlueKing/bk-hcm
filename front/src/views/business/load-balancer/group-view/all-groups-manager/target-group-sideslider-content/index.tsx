import { computed, defineComponent, reactive, ref, watch } from 'vue';
import { Form, Select, Input } from 'bkui-vue';
import { useAccountStore, useBusinessStore } from '@/store';
import AccountSelector from '@/components/account-selector/index.vue';
import RsConfigTable from '../rs-config-table';
import './index.scss';
import { TARGET_GROUP_PROTOCOLS, VendorEnum } from '@/common/constant';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector';
import VpcSelector from '@/components/vpc-selector/index.vue';
import { QueryRuleOPEnum } from '@/typings';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  name: 'TargetGroupSidesliderContent',
  props: {
    rsTableData: {
      type: Array,
      required: true,
    },
    isEdit: {
      type: Boolean,
      default: false,
    },
    editData: {
      type: Object,
      default: {},
    },
  },
  emits: ['showAddRsDialog', 'change'],
  setup(props, { emit }) {
    const accountStore = useAccountStore();
    const businessStore = useBusinessStore();
    const rsList = ref([]);

    const formData = reactive({
      bk_biz_id: accountStore.bizs,
      account_id: '',
      name: '',
      protocol: '',
      port: 80,
      region: '',
      vpc_id: [],
    });
    const curVendor = ref(VendorEnum.TCLOUD);
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
        property: 'cloud_account_id',
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
          property: 'target_group_name',
          span: 12,
          content: () => <Input v-model={formData.name} />,
        },
        {
          label: '协议端口',
          required: true,
          property: 'protocol_port',
          span: 12,
          content: () => (
            <div class='flex-row'>
              <Select v-model={formData.protocol}>
                {TARGET_GROUP_PROTOCOLS.map((protocol) => (
                  <Option name={protocol} id={protocol}></Option>
                ))}
              </Select>
              &nbsp;&nbsp;:&nbsp;&nbsp;
              <Input v-model={formData.port}></Input>
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
          label: '网络',
          required: true,
          property: 'net',
          span: 12,
          content: () => (
            <VpcSelector
              v-model={formData.vpc_id}
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
          <RsConfigTable onShowAddRsDialog={() => emit('showAddRsDialog', rsList.value)} details={props.rsTableData} />
        ),
      },
    ]);

    watch(
      () => formData,
      () => {
        emit('change', formData);
      },
      {
        deep: true,
      },
    );

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
      rsList.value = res.data.details;
    };

    watch(
      () => formData.account_id,
      (id) => {
        getAllRsList(id);
      },
      {
        immediate: true,
      },
    );

    watch(
      () => props.editData,
      (data) => {
        if (props.isEdit) {
          formData.account_id = data.account_id;
          formData.name = data.name;
          formData.protocol = data.protocol;
          formData.port = data.port;
          formData.region = data.region;
          formData.vpc_id = data.vpc_id;
        }
      },
      {
        immediate: true,
        deep: true,
      },
    );

    return () => (
      <bk-container margin={0} class='target-group-sideslider-content'>
        <Form formType='vertical'>
          {formItemOptions.value.map((item) => (
            <bk-row>
              {Array.isArray(item) ? (
                item.map((subItem) => (
                  <bk-col span={subItem.span}>
                    <FormItem label={subItem.label} property={subItem.property} required={subItem.required}>
                      {subItem.content()}
                    </FormItem>
                  </bk-col>
                ))
              ) : (
                <bk-col span={item.span}>
                  <FormItem label={item.label} property={item.property} required={item.required}>
                    {item.content()}
                  </FormItem>
                </bk-col>
              )}
            </bk-row>
          ))}
        </Form>
      </bk-container>
    );
  },
});
