import { defineComponent, PropType, ref } from 'vue';

import { Button, Dialog } from 'bkui-vue';

import { useI18n } from 'vue-i18n';
import { BillsExportResData } from '@/typings/bill';

export default defineComponent({
  props: { cb: Function as PropType<() => Promise<BillsExportResData>>, fileName: String },
  setup(props) {
    const { t } = useI18n();

    const isShow = ref(false);
    const isLoading = ref(false);
    const downloadUrl = ref('');

    const handleExport = async () => {
      isLoading.value = true;
      try {
        const res = await props.cb();
        downloadUrl.value = res.data.download_url;
        isShow.value = true;
      } catch (error) {
        downloadUrl.value = '';
      } finally {
        isLoading.value = false;
      }
    };

    return () => (
      <>
        <Button onClick={handleExport} loading={isLoading.value}>
          {t('导出')}
        </Button>
        <Dialog v-model:isShow={isShow.value} title={props.fileName} quick-close dialogType='show'>
          <a href={downloadUrl.value} download={props.fileName} style={{ wordBreak: 'break-all' }}>
            {downloadUrl.value}
          </a>
        </Dialog>
      </>
    );
  },
});
