import { computed, defineComponent, onMounted, reactive, ref } from 'vue';
import { Button, Form, Input, Upload } from 'bkui-vue';
import BkRadio, { BkRadioGroup } from 'bkui-vue/lib/radio';
import './index.scss';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import CommonSideslider from '@/components/common-sideslider';
import http from '@/http';
import { useAccountStore } from '@/store';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { FormItem } = Form;

export default defineComponent({
  name: 'CertManager',
  setup() {
    const accountStore = useAccountStore();

    const { columns } = useColumns('cert');
    const { CommonTable } = useTable({
      columns: [
        ...columns,
        {
          label: '操作',
          width: 120,
          render: () => (<span class='operate-text-btn'>删除</span>),
        },
      ],
      searchUrl: '',
      searchData: [
        {
          name: '资源ID',
          id: 'resourceId',
        },
        {
          name: '云厂商',
          id: 'cloudProvider',
        },
        {
          name: '证书类型',
          id: 'certificateType',
        },
        {
          name: '域名',
          id: 'domainName',
        },
        {
          name: '上传时间',
          id: 'uploadTime',
        },
        {
          name: '过期时间',
          id: 'expirationTime',
        },
        {
          name: '证书状态',
          id: 'certificateStatus',
        },
      ],
      tableData: [
        {
          resourceId: 'res-123',
          cloudProvider: '亚马逊AWS',
          certificateType: 'EV SSL',
          domainName: 'example.com',
          uploadTime: '2023-01-01 10:00:00',
          expirationTime: '2024-01-01 10:00:00',
          certificateStatus: '正常',
        },
        {
          resourceId: 'res-456',
          cloudProvider: '阿里云',
          certificateType: 'OV SSL',
          domainName: 'example.net',
          uploadTime: '2023-02-01 11:00:00',
          expirationTime: '2024-02-01 11:00:00',
          certificateStatus: '正常',
        },
        {
          resourceId: 'res-789',
          cloudProvider: '腾讯云',
          certificateType: 'DV SSL',
          domainName: 'example.org',
          uploadTime: '2023-03-01 12:00:00',
          expirationTime: '2024-03-01 12:00:00',
          certificateStatus: '已过期',
        },
      ],
    });
    const isCertUploadSidesliderShow = ref(false);
    const formModel = reactive({
      name: '' as string, // 证书名称
      type: '服务器证书' as string, // 证书类型
      cert_key: '' as string, // 证书信息
      primary_key: '' as string, // 私钥信息
    });
    const formItemOptions = computed(() => [
      {
        label: '证书名称',
        property: 'name',
        required: true,
        content: () => <Input v-model={formModel.name} />,
      },
      {
        label: '证书类型',
        property: 'type',
        required: true,
        content: () => (
          <BkRadioGroup v-model={formModel.type}>
            <BkRadio label='服务器证书'></BkRadio>
            <BkRadio label='客户端CA证书'></BkRadio>
          </BkRadioGroup>
        ),
      },
      {
        label: '证书上传',
        property: 'cert_key',
        required: true,
        content: () => (
          <>
            <Upload theme='button' tip='支持扩展名: .crt或.pem' />
            <Input v-model={formModel.cert_key} type='textarea' resize={false} rows={3}></Input>
          </>
        ),
      },
      {
        label: '私钥上传',
        property: 'primary_key',
        required: true,
        hidden: formModel.type === '客户端CA证书',
        content: () => (
          <>
            <Upload theme='button' tip='支持扩展名: .key' />
            <Input v-model={formModel.primary_key} type='textarea' resize={false} rows={3}></Input>
          </>
        ),
      },
    ]);

    const getCertList = async () => {
      const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${accountStore.bizs}/certs/list`);
      console.log(result);
    };

    onMounted(() => {
      getCertList();
    });
    return () => (
      <div class='cert-manager-page'>
        <div class='common-card-wrap'>
          <CommonTable>
            {{
              operation: () => (
                <Button theme='primary' onClick={() => (isCertUploadSidesliderShow.value = true)}>
                  上传证书
                </Button>
              ),
            }}
          </CommonTable>
        </div>
        <CommonSideslider
          v-model:isShow={isCertUploadSidesliderShow.value}
          title='证书上传'
          width='640'
          class='cert-upload-sideslider'>
          <Form formType='vertical'>
            {formItemOptions.value.map(({ label, property, required, content, hidden }) => {
              if (hidden) return null;
              return (
                <FormItem label={label} required={required} key={property}>
                  {content()}
                </FormItem>
              );
            })}
          </Form>
        </CommonSideslider>
      </div>
    );
  },
});
