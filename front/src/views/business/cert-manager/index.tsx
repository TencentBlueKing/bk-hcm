import { computed, defineComponent, reactive, ref } from 'vue';
import { Button, Form, Input, Upload, Message, PopConfirm } from 'bkui-vue';
import BkRadio, { BkRadioGroup } from 'bkui-vue/lib/radio';
import './index.scss';
import { VendorEnum } from '@/common/constant';
import { DoublePlainObject } from '@/typings';
import { useTable } from '@/hooks/useTable/useTable';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useAccountStore, useResourceStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import CommonSideslider from '@/components/common-sideslider';
import AccountSelector from '@/components/account-selector/index.vue';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';

const { FormItem } = Form;

export default defineComponent({
  name: 'CertManager',
  setup() {
    const { isResourcePage, isBusinessPage } = useWhereAmI();
    const accountStore = useAccountStore();
    const resourceStore = useResourceStore();
    const resourceAccountStore = useResourceAccountStore();

    const { selections, handleSelectionChange, resetSelections } = useSelection();

    const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
      if (isCheckAll) return true;
      return isCurRowSelectEnable(row);
    };
    const isCurRowSelectEnable = (row: any) => {
      if (isBusinessPage) return true;
      if (row.id) {
        return row.bk_biz_id === -1;
      }
    };

    const { columns } = useColumns('cert');
    const tableColumns = computed(() => {
      const result = [
        ...columns,
        {
          label: '操作',
          width: 120,
          render: ({ data }: { data: any }) => (
            <PopConfirm
              trigger='click'
              content={`是否删除证书「${data.name}」？`}
              onConfirm={() => handleDeleteCert(data)}>
              <span class='operate-text-btn'>删除</span>
            </PopConfirm>
          ),
        },
      ];
      if (isResourcePage) {
        result.unshift({
          type: 'selection',
          width: 32,
          minWidth: 32,
          onlyShowOnList: true,
          align: 'right',
        });
      }
      return result;
    });
    const { CommonTable, getListData } = useTable({
      columns: tableColumns.value,
      type: 'certs',
      searchData: [
        {
          name: '证书名称',
          id: 'name',
        },
        {
          name: '资源ID',
          id: 'cloud_id',
        },
        {
          name: '域名',
          id: 'domain',
        },
      ],
      tableExtraOptions: {
        isRowSelectEnable,
        onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
        onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
      },
    });
    const isCertUploadSidesliderShow = ref(false);
    const formRef = ref();
    const formModel = reactive({
      account_id: '' as string, // 账户ID
      name: '' as string, // 证书名称
      vendor: VendorEnum.TCLOUD, // 云厂商
      cert_type: 'SVR' as 'CA' | 'SVR', // 证书类型
      public_key: '' as string, // 证书信息
      private_key: '' as string, // 私钥信息
    });
    const formItemOptions = computed(() => [
      {
        label: '云账号',
        property: 'account_id',
        required: true,
        content: () => (
          <AccountSelector
            v-model={formModel.account_id}
            disabled={!!resourceAccountStore?.resourceAccount?.id}
            mustBiz={!isResourcePage}
            bizId={accountStore.bizs}
            type='resource'></AccountSelector>
        ),
      },
      {
        label: '证书名称',
        property: 'name',
        required: true,
        content: () => <Input v-model={formModel.name} />,
      },
      {
        label: '证书类型',
        property: 'cert_type',
        required: true,
        content: () => (
          <BkRadioGroup v-model={formModel.cert_type}>
            <BkRadio label='SVR'>服务器证书</BkRadio>
            <BkRadio label='CA'>客户端CA证书</BkRadio>
          </BkRadioGroup>
        ),
      },
      {
        label: '证书上传',
        property: 'public_key',
        required: true,
        content: () => (
          <>
            <Upload
              theme='button'
              tip='支持扩展名: .crt或.pem'
              validate-name={/\.(crt|pem)$/i}
              multiple={false}
              custom-request={({ file }: { file: any }) => handleUploadCertKey(file)}
              onDelete={() => (formModel.public_key = '')}
            />
            <Input v-model={formModel.public_key} type='textarea' resize={false} rows={3}></Input>
          </>
        ),
      },
      {
        label: '私钥上传',
        property: 'private_key',
        required: true,
        hidden: formModel.cert_type === 'CA',
        content: () => (
          <>
            <Upload
              theme='button'
              tip='支持扩展名: .key'
              validate-name={/\.key$/i}
              multiple={false}
              custom-request={({ file }: { file: any }) => handleUploadPrimaryKey(file)}
              onDelete={() => (formModel.private_key = '')}
            />
            <Input v-model={formModel.private_key} type='textarea' resize={false} rows={3}></Input>
          </>
        ),
      },
    ]);

    const resetFormParams = () => {
      Object.assign(formModel, {
        account_id: resourceAccountStore?.resourceAccount?.id || '',
        name: '',
        vendor: VendorEnum.TCLOUD,
        cert_type: 'SVR',
        public_key: '',
        private_key: '',
      });
    };

    const showCreateCertSideslider = () => {
      isCertUploadSidesliderShow.value = true;
      resetFormParams();
    };

    // 回显证书内容
    const echoCertContent = (file: any, key: string) => {
      const fileReader = new FileReader();
      fileReader.onload = (e: any) => {
        formModel[key] = e.target.result;
      };
      fileReader.readAsText(file);
    };
    // 证书上传文件之间执行的钩子
    const handleUploadCertKey = (file: any) => {
      echoCertContent(file, 'public_key');
    };
    // 密钥上传文件之间执行的钩子
    const handleUploadPrimaryKey = (file: any) => {
      echoCertContent(file, 'private_key');
    };

    // 处理参数
    const resolveFormParams = () => {
      // 证书内容转 base64
      Object.assign(formModel, {
        public_key: btoa(formModel.public_key),
        private_key: btoa(formModel.private_key),
      });
    };
    // 证书上传
    const handleCreateCert = async () => {
      await formRef.value.validate();
      resolveFormParams();
      await resourceStore.create('certs', formModel);
      Message({ theme: 'success', message: '证书上传成功' });
      isCertUploadSidesliderShow.value = false;
      await getListData();
    };
    // 删除指定证书
    const handleDeleteCert = async (cert: any) => {
      await resourceStore.delete('certs', cert.id);
      Message({ theme: 'success', message: '证书删除成功' });
      await getListData();
    };

    return () => (
      <div class={`cert-manager-page${isResourcePage ? ' has-selection' : ''}`}>
        <div class='common-card-wrap'>
          <CommonTable>
            {{
              operation: () => (
                <>
                  <Button theme='primary' onClick={showCreateCertSideslider}>
                    上传证书
                  </Button>
                  <BatchDistribution
                    selections={selections.value}
                    type={DResourceType.certs}
                    getData={() => {
                      getListData();
                      resetSelections();
                    }}
                  />
                </>
              ),
            }}
          </CommonTable>
        </div>
        <CommonSideslider
          v-model:isShow={isCertUploadSidesliderShow.value}
          title='证书上传'
          width='640'
          onHandleSubmit={handleCreateCert}
          class='cert-upload-sideslider'>
          <Form ref={formRef} formType='vertical' model={formModel}>
            {formItemOptions.value.map(({ label, property, required, content, hidden }) => {
              if (hidden) return null;
              return (
                <FormItem key={property} label={label} required={required} property={property}>
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
