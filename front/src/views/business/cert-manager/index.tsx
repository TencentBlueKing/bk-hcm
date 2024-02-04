import { computed, defineComponent, reactive, ref, PropType } from 'vue';
import { Button, Form, Input, Upload, Message } from 'bkui-vue';
import BkRadio, { BkRadioGroup } from 'bkui-vue/lib/radio';
import './index.scss';
import { VendorEnum } from '@/common/constant';
import { DoublePlainObject, FilterType } from '@/typings';
import { useTable } from '@/hooks/useTable/useTable';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useAccountStore, useResourceStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import CommonSideslider from '@/components/common-sideslider';
import AccountSelector from '@/components/account-selector/index.vue';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';
import Confirm from '@/components/confirm';

const { FormItem } = Form;

export default defineComponent({
  name: 'CertManager',
  props: {
    filter: Object as PropType<FilterType>,
  },
  setup(props) {
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
            <Button
              text
              theme='primary'
              onClick={() => handleDeleteCert(data)}
              disabled={isResourcePage && data.bk_biz_id !== -1}
              v-bk-tooltips={{
                content: '该证书已分配业务, 仅可在业务下操作',
                disabled: isBusinessPage || data.bk_biz_id === -1,
              }}>
              删除
            </Button>
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
      searchOptions: {
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
      },
      tableOptions: {
        columns: tableColumns.value,
        extra: {
          isRowSelectEnable,
          onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
        },
      },
      requestOption: {
        type: 'certs',
        sortOption: {
          sort: 'cloud_created_time',
          order: 'DESC',
        },
      },
      bizFilter: props.filter,
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

    // 上传证书错误提示
    const uploadPublicKeyErrorText = ref('');
    const uploadPrivateKeyErrorText = ref('');
    // 错误提示映射
    const errorTextMap = {
      public_key: uploadPublicKeyErrorText,
      private_key: uploadPrivateKeyErrorText,
    };

    // 表单项配置
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
              limit={1}
              multiple={false}
              custom-request={({ file }: { file: any }) => handleUploadCertKey(file)}
              onDelete={() => handleUploadFileDelete('public_key')}
              onError={(_: any, fileList: any, error: Error) => handleUploadError(fileList, error, 'public_key')}
              onExceed={() => handleUploadExceed('public_key')}
            />
            {uploadPublicKeyErrorText.value && <div class='upload-error-text'>{uploadPublicKeyErrorText.value}</div>}
            <Input v-model={formModel.public_key} type='textarea' rows={5} class='upload-textarea-wrap'></Input>
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
              limit={1}
              multiple={false}
              custom-request={({ file }: { file: any }) => handleUploadPrimaryKey(file)}
              onDelete={() => handleUploadFileDelete('private_key')}
              onError={(_: any, fileList: any, error: Error) => handleUploadError(fileList, error, 'private_key')}
              onExceed={() => handleUploadExceed('private_key')}
            />
            {uploadPrivateKeyErrorText.value && <div class='upload-error-text'>{uploadPrivateKeyErrorText.value}</div>}
            <Input v-model={formModel.private_key} type='textarea' rows={5} class='upload-textarea-wrap'></Input>
          </>
        ),
      },
    ]);

    const resetForm = () => {
      Object.assign(formModel, {
        account_id: resourceAccountStore?.resourceAccount?.id || '',
        name: '',
        vendor: VendorEnum.TCLOUD,
        cert_type: 'SVR',
        public_key: '',
        private_key: '',
      });
      uploadPublicKeyErrorText.value = '';
      uploadPrivateKeyErrorText.value = '';
    };

    const showCreateCertSideslider = () => {
      isCertUploadSidesliderShow.value = true;
      resetForm();
    };

    // 回显证书内容
    const echoCertContent = (file: any, key: string) => {
      const fileReader = new FileReader();
      fileReader.onload = (e: any) => {
        formModel[key] = e.target.result;
      };
      fileReader.readAsText(file);
    };
    // 处理证书上传文件成功执行的事件
    const handleUploadCertKey = (file: any) => {
      echoCertContent(file, 'public_key');
      uploadPublicKeyErrorText.value = '';
    };
    // 处理密钥上传文件成功执行的事件
    const handleUploadPrimaryKey = (file: any) => {
      echoCertContent(file, 'private_key');
      uploadPrivateKeyErrorText.value = '';
    };
    // 处理文件上传失败的事件
    const handleUploadError = (fileList: any, error: Error, type: string) => {
      if (error.message === 'invalid filename') {
        errorTextMap[type].value = '请上传正确的证书文件，该证书将于 2s 后移除！';
        setTimeout(() => {
          fileList.pop();
          errorTextMap[type].value = '';
        }, 2000);
      }
    };
    // 处理文件上传个数超出限制后的事件
    const handleUploadExceed = (type: string) => {
      errorTextMap[type].value = '证书文件只支持上传 1 个，如需更换，请移除当前证书文件后再进行上传操作！';
    };
    // 处理
    const handleUploadFileDelete = (type: 'public_key' | 'private_key') => {
      formModel[type] = '';
      errorTextMap[type].value = '';
    };

    // 证书上传
    const handleCreateCert = async () => {
      await formRef.value.validate();
      await resourceStore.create('certs', {
        ...formModel,
        public_key: btoa(formModel.public_key),
        private_key: btoa(formModel.private_key),
      });
      Message({ theme: 'success', message: '证书上传成功' });
      isCertUploadSidesliderShow.value = false;
      await getListData();
    };
    // 删除指定证书
    const handleDeleteCert = async (cert: any) => {
      Confirm('请确定删除证书', `将删除证书【${cert.name}】`, () => {
        resourceStore.delete('certs', cert.id).then(() => {
          Message({ theme: 'success', message: '证书删除成功' });
          getListData();
        });
      });
    };

    return () => (
      <div class={`cert-manager-page${isResourcePage ? ' has-selection' : ''}`}>
        <div class='common-card-wrap'>
          <CommonTable>
            {{
              operation: () => (
                <>
                  <Button class='mw88' theme='primary' onClick={showCreateCertSideslider}>
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
