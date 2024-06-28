import { defineComponent, ref } from 'vue';

import { Form, Table } from 'bkui-vue';
import PrimaryAccountSelector from '../../components/search/primary-account-selector';
import VendorRadioGroup from '@/components/vendor-radio-group';
import CommonSideslider from '@/components/common-sideslider';
import Amount from '../../components/amount';

import { useI18n } from 'vue-i18n';

export default defineComponent({
  setup(_, { expose }) {
    const { t } = useI18n();
    const isShow = ref(false);
    const modal = ref({ vendor: [] });

    const triggerShow = (v: boolean) => {
      isShow.value = v;
    };

    expose({ triggerShow });

    return () => (
      <CommonSideslider v-model:isShow={isShow.value} width={1280} title='新增调账'>
        {{
          default: () => (
            <Form formType='vertical'>
              <Form.FormItem label={t('云厂商')} required>
                <VendorRadioGroup />
              </Form.FormItem>
              <Form.FormItem label={t('一级账号')} required>
                <PrimaryAccountSelector vendor={modal.value.vendor} />
              </Form.FormItem>
              <Form.FormItem label={t('调账配置')} required>
                <Table
                  columns={[
                    { label: 'col1', field: 'col1' },
                    { label: 'col2', field: 'col2' },
                  ]}></Table>
              </Form.FormItem>
              <Form.FormItem label={t('结果预览')}>
                <Amount isAdjust showType='vertical' />
              </Form.FormItem>
            </Form>
          ),
        }}
      </CommonSideslider>
    );
  },
});
