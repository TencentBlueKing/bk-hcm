import { computed, defineComponent, reactive } from 'vue';
import { Form, Select, Input } from 'bkui-vue';
import { useAccountStore } from '@/store';
import AccountSelector from '@/components/account-selector/index.vue';
import RsConfigTable from '../rs-config-table';
import './index.scss';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  name: 'TargetGroupSidesliderContent',
  emits: ['showAddRsDialog'],
  setup(_props, { emit }) {
    const accountStore = useAccountStore();

    const formData = reactive({
      bizId: -1,
      cloudAccountId: '',
      targetGroupName: '',
      protocol: '',
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
              <Select v-model={formData.protocol}>
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
        content: () => <RsConfigTable onShowAddRsDialog={() => emit('showAddRsDialog')} />,
      },
    ]);

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
