import { defineComponent, reactive, ref } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { Button, Form, Input, Sideslider, Upload } from 'bkui-vue';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  name: 'certificateSideslider',
  setup() {
    const { columns, settings } = useColumns('certificate');
    // const searchValue = ref('');
    const searchData: any = [];
    const searchUrl = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vpcs/list`;
    const { CommonTable } = useTable({
      columns,
      settings: settings.value,
      searchData,
      searchUrl,
    });
    const isShow = ref(false);
    const formRef = ref();
    const formData = reactive({
      name: '',
    });
    const limit = 1;
    const uploader = ref(null);

    function BkMessage(arg0: { theme: string; message: string }) {}

    const handleExceed = (files: any, fileList: any) => {
      console.log(files, fileList, 'handleExceed');
      BkMessage({
        theme: 'error',
        message: `最多上传${limit}个文件`,
      });
    };
    const handleRes = (response: { id: any }) => {
      if (response.id) {
        return true;
      }
      return false;
    };
    // const handleSubmit = () => {};

    // const triggerShow = () => {};

    const handleOpenSlider = () => {
      isShow.value = true;
    };

    return () => (
      <div style={{ padding: '20px' }}>
        <CommonTable>
          {{
            operation: () => (
              <>
                <Button
                  theme='primary'
                  onClick={handleOpenSlider}
                  style={{ width: '100px' }}>
                  上传证书
                </Button>
                <bk-sideslider
                  v-model:isShow={isShow.value}
                  title='证书上传'
                  quick-close>
                  <div style={{ margin: '30px' }}>
                    <bk-form
                      ref={formRef}
                      model='formData'
                      form-type='vertical'>
                      <bk-form-item label='证书名称:' property='name' required>
                        <bk-input
                          v-model={formData.name}
                          placeholder='请输入'
                          clearable
                        />
                      </bk-form-item>

                      <bk-form-item label='证书类型' required>
                        <bk-radio-group>
                          <bk-radio label='服务器证书' />
                          <bk-radio label='客户端CA证书' />
                        </bk-radio-group>
                      </bk-form-item>

                      <bk-form-item label='证书上传' required>
                        <bk-Upload
                          theme='button'
                          limit="limit"
                          onExceed={handleExceed}
                          tip="'支持扩展名：.crt或.pem'"></bk-Upload>

                        <bk-input
                          placeholder='未输入'
                          type='textarea'
                          disabled
                        />
                      </bk-form-item>

                      <bk-form-item label='密钥上传' required>
                        <bk-upload
                          theme='button'
                          limit="limit"
                          onExceed={handleExceed}
                          tip="'支持扩展名：.crt或.pem'"></bk-upload>
                        <bk-input
                          placeholder='未输入'
                          type='textarea'
                          disabled
                        />
                      </bk-form-item>
                    </bk-form>
                    {/* <Button theme='primary' onClick={handleSubmit}>
                      提交
                    </Button>
                    <Button onClick={() => triggerShow()}>取消</Button> */}
                  </div>
                </bk-sideslider>
              </>
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
