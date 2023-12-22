import { computed, defineComponent, reactive, ref } from 'vue';
import { Form, Select, Input, SearchSelect, Loading, Table } from 'bkui-vue';
import { useAccountStore } from '@/store';
import AccountSelector from '@/components/account-selector/index.vue';
import Empty from '@/components/empty';
import './index.scss';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  name: 'TargetGroupSidesliderContent',

  setup() {
    const accountStore = useAccountStore();

    const formData = reactive({
      bizId: -1,
      cloudAccountId: '',
      targetGroupName: '',
      protocal: '',
      port: 80,
      region: '',
      net: '',
      rs_list: [],
    });
    const selectedBizId = computed({
      get() {
        return accountStore.bizs;
      },
      set(val) {
        formData.bizId = val;
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
            v-model={formData.cloudAccountId}
            bizId={selectedBizId.value}
            type='resource'></AccountSelector>
        ),
      },
      [
        {
          label: '目标组名称',
          required: true,
          property: 'target_group_name',
          span: 12,
          content: () => (
            <Select v-model={formData.targetGroupName}>
              <Option name='1'>选项一</Option>
              <Option name='2'>选项二</Option>
              <Option name='3'>选项三</Option>
            </Select>
          ),
        },
        {
          label: '协议端口',
          required: true,
          property: 'protocol_port',
          span: 12,
          content: () => (
            <div class='flex-row'>
              <Select v-model={formData.protocal}>
                <Option name='1'>选项一</Option>
                <Option name='2'>选项二</Option>
                <Option name='3'>选项三</Option>
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
            <Select v-model={formData.region}>
              <Option name='1'>选项一</Option>
              <Option name='2'>选项二</Option>
              <Option name='3'>选项三</Option>
            </Select>
          ),
        },
        {
          label: '网络',
          required: true,
          property: 'net',
          span: 12,
          content: () => (
            <Select v-model={formData.net}>
              <Option name='1'>选项一</Option>
              <Option name='2'>选项二</Option>
              <Option name='3'>选项三</Option>
            </Select>
          ),
        },
      ],
      {
        label: 'RS 配置',
        required: true,
        property: 'rs_list',
        span: 24,
        content: () => (
          <>
            <div class='operation-wrap'>
              <div class='left-wrap'>
                <i class='hcm-icon bkhcm-icon-plus-circle-shape'></i>
                <span>添加 RS</span>
              </div>
              <div class='search-wrap'>
                <SearchSelect></SearchSelect>
              </div>
            </div>
            <Loading loading={isTableLoading.value}>
              <Table data={[]} columns={columns}>
                {{
                  empty: () => {
                    if (isTableLoading.value) return null;
                    return <Empty text='暂未添加实例' />;
                  },
                }}
              </Table>
            </Loading>
          </>
        ),
      },
    ]);

    const columns = [
      {
        label: '内网IP',
        field: 'private_ipv4_or_ipv6',
      },
      {
        label: '名称',
        field: 'rs_name',
      },
      {
        label: () => (
          <div>
            <span>协议</span>
            <i class='hcm-icon bkhcm-icon-bianji'></i>
          </div>
        ),
        field: 'port',
      },
      {
        label: () => (
          <div>
            <span>权重</span>
            <i class='hcm-icon bkhcm-icon-bianji'></i>
          </div>
        ),
        field: 'weight',
      },
    ];

    const isTableLoading = ref(false);

    return () => (
      <bk-container margin={0}>
        <Form formType='vertical'>
          {formItemOptions.value.map(item => (
            <bk-row>
              {Array.isArray(item) ? (
                item.map(subItem => (
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
