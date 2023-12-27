import { computed, defineComponent, reactive, ref } from 'vue';
import { Form, Select, Input, SearchSelect, Loading, Table } from 'bkui-vue';
import { useAccountStore } from '@/store';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import AccountSelector from '@/components/account-selector/index.vue';
import BatchUpdatePopconfirm from '@/components/batch-update-popconfirm';
import CommonDialog from '@/components/common-dialog';
import AddRsDialogContent from '../add-rs-dialog-content';
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

    // rs配置
    const isTableLoading = ref(false);
    const { columns, settings } = useColumns('rsConfig');
    const handleBatchUpdatePort = (_port: number) => {};
    const handleBatchUpdateWeight = (_weight: number) => {};
    const handleDeleteRs = () => {};
    const rsTableColumns = [
      ...columns,
      {
        label: () => (
          <>
            <span>端口</span>
            <BatchUpdatePopconfirm title='端口' onUpdateValue={handleBatchUpdatePort}></BatchUpdatePopconfirm>
          </>
        ),
        field: 'port',
        isDefaultShow: true,
      },
      {
        label: () => (
          <>
            <span>权重</span>
            <BatchUpdatePopconfirm title='权重' onUpdateValue={handleBatchUpdateWeight}></BatchUpdatePopconfirm>
          </>
        ),
        field: 'weight',
        isDefaultShow: true,
      },
      {
        label: '',
        width: 80,
        render: () => <i class='hcm-icon bkhcm-icon-minus-circle-shape' onClick={handleDeleteRs}></i>,
      },
    ];
    const rsTableSettings = Object.assign(settings.value);
    rsTableSettings.checked.push('port', 'weight');
    rsTableSettings.fields.push(
      { label: '端口', field: 'port', isDefaultShow: true },
      { label: '权重', field: 'weight', isDefaultShow: true },
    );
    const rsConfigData = [
      {
        privateIp: '10.0.0.1',
        publicIp: '203.0.113.10',
        name: '服务器A',
        region: '华北1区',
        resourceType: 'VM',
        network: 'VPC-XYZ',
        port: 8080,
        weight: 20,
      },
      {
        privateIp: '10.0.1.2',
        publicIp: '203.0.113.20',
        name: '数据库B',
        region: '华东2区',
        resourceType: 'RDS',
        network: 'VPC-ABC',
        port: 3306,
        weight: 10,
      },
      {
        privateIp: '10.0.2.3',
        publicIp: '203.0.113.30',
        name: '负载均衡C',
        region: '华南3区',
        resourceType: 'LoadBalancer',
        network: 'VPC-DEF',
        port: 80,
        weight: 30,
      },
    ];

    const isAddRsDialogShow = ref(false);
    const handleAddRs = () => {};
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
        content: () => (
          <>
            <div class='operation-wrap'>
              <div class='left-wrap' onClick={() => (isAddRsDialogShow.value = true)}>
                <i class='hcm-icon bkhcm-icon-plus-circle-shape'></i>
                <span>添加 RS</span>
              </div>
              <div class='search-wrap'>
                <SearchSelect></SearchSelect>
              </div>
            </div>
            <Loading loading={isTableLoading.value}>
              <Table data={rsConfigData} columns={rsTableColumns} settings={settings.value} showOverflowTooltip>
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

    return () => (
      <bk-container margin={0} class='target-group-sideslider-content'>
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
        <CommonDialog
          v-model:isShow={isAddRsDialogShow.value}
          title='添加 RS'
          width={640}
          onHandleConfirm={handleAddRs}>
          <AddRsDialogContent />
        </CommonDialog>
      </bk-container>
    );
  },
});
