import { defineComponent, ref } from 'vue';
import { Button, DatePicker, Dialog, Upload } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';
import Amount from '../../components/amount';

export default defineComponent({
  name: 'ImportBillDetailDialog',
  setup(_, { expose }) {
    const { t } = useI18n();

    const isShow = ref(false);

    const triggerShow = (v: boolean) => {
      isShow.value = v;
    };

    const handleConfirm = async () => {
      triggerShow(false);
    };

    expose({ triggerShow });

    return () => (
      <Dialog v-model:isShow={isShow.value} title={t('导入')} width={640}>
        {{
          default: () => (
            <>
              <div class='flex-row mb30'>
                <div class='mr24'>{t('云厂商')}</div>
                <div>zenlayer</div>
              </div>
              <div class='mb30'>
                <div class='mb6'>{t('核算月份')}</div>
                <DatePicker class={cssModule.datePicker} appendToBody />
              </div>
              <div>
                <div class='mb6'>{t('文件上传')}</div>
                <Upload />
              </div>
            </>
          ),
          footer: () => (
            <div class={cssModule.footer}>
              <Amount class={cssModule.amounts} />
              <Button theme='primary' onClick={handleConfirm}>
                {t('确定')}
              </Button>
              <Button onClick={() => triggerShow(false)}>{t('取消')}</Button>
            </div>
          ),
        }}
      </Dialog>
    );
  },
});
