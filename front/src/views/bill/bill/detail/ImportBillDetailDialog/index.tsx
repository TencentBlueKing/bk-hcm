import { defineComponent, inject, PropType, reactive, Ref, ref, watch } from 'vue';
import cssModule from './index.module.scss';

import { Button, Dialog, Message, Upload } from 'bkui-vue';
import Amount from '../../components/amount';
import zenlayerIcon from '@/assets/image/zenlayer.png';

import { useI18n } from 'vue-i18n';
import { VendorEnum } from '@/common/constant';
import { billItemsImportPreview, billItemsImport } from '@/api/bill';
import { BillImportPreviewItems, CostMap } from '@/typings/bill';

export default defineComponent({
  name: 'ImportBillDetailDialog',
  props: { vendor: String as PropType<VendorEnum> },
  setup(props, { expose }) {
    const { t } = useI18n();
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const isShow = ref(false);
    const importPreviewInfo = reactive<{ cost_map: CostMap; count: number }>({ cost_map: null, count: 0 });
    let importItems: BillImportPreviewItems = [];

    const uploadRef = ref();
    let uploadFile: any = null;

    const triggerShow = (v: boolean) => {
      isShow.value = v;
    };

    const handlePreview = ({ file }: any) => {
      return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = async (e) => {
          try {
            const base64String = e.target.result as string;
            const res = await billItemsImportPreview(props.vendor, {
              bill_year: bill_year.value,
              bill_month: bill_month.value,
              excel_file_base64: base64String.replace(/^data:.*;base64,/, ''),
            });
            // 用于展示的信息
            importPreviewInfo.cost_map = res.data?.cost_map;
            importPreviewInfo.count = res.data?.items?.length || 0;
            // 用于上传的 items
            importItems = res.data?.items;
            uploadFile = file;
            resolve(res);
          } catch (error) {
            reject(translateErrorMessage(error));
          }
        };
        reader.readAsDataURL(file);
      });
    };

    const translateErrorMessage = (error: any) => {
      switch (error.code) {
        case 2000015:
          error.message = t('导入数据的月份和核算月份不匹配');
        case 2000016:
          error.message = t('Excel数据字段不符合要求');
      }
      return error;
    };

    const handleSuccess = (_res: any, file: any, fileList: any[]) => {
      const { uid } = file;
      const targetFile = fileList.find((item: any) => item.uid === uid);
      targetFile.statusText = `${t('共导入')} ${importPreviewInfo.count} ${t('条数据')} `;
    };

    const handleConfirm = async () => {
      triggerShow(false);
      await billItemsImport(props.vendor, {
        bill_year: bill_year.value,
        bill_month: bill_month.value,
        items: importItems,
      });
      Message({ theme: 'success', message: t('导入成功') });
    };

    const reset = () => {
      uploadRef.value.handleRemove(uploadFile);
      Object.assign(importPreviewInfo, { cost_map: null, count: 0 });
      importItems = [];
      uploadFile = null;
    };

    watch([bill_year, bill_month, () => props.vendor], () => {
      reset();
    });

    expose({ triggerShow });

    return () => (
      <Dialog v-model:isShow={isShow.value} title={t('导入')} width={700}>
        {{
          default: () => (
            <>
              <div class={cssModule.item}>
                <div class={cssModule.label}>{t('云厂商')}</div>
                <div class={cssModule.content}>
                  <img src={zenlayerIcon} alt='' class='mr8' />
                  zenlayer
                </div>
              </div>
              <div class={cssModule.item}>
                <div class={cssModule.label}>{t('核算月份')}</div>
                <div class={cssModule.content}>{`${bill_year.value}-${bill_month.value}`}</div>
              </div>
              <div>
                <div class='mb6'>{t('文件上传')}</div>
                <Upload
                  ref={uploadRef}
                  multiple={false}
                  limit={1}
                  accept='.xlsx'
                  validateName={/\.xlsx$/i}
                  customRequest={handlePreview}
                  onSuccess={handleSuccess}>
                  {{
                    tip: () => <div class={cssModule.uploadTip}>{t('仅支持.xlsx格式的文件')}</div>,
                  }}
                </Upload>
              </div>
            </>
          ),
          footer: () => (
            <div class={cssModule.footer}>
              <Amount class={cssModule.amounts} data={importPreviewInfo.cost_map} />
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
