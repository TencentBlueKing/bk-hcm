import { defineComponent, PropType, ref } from 'vue';

import { Button, Dialog } from 'bkui-vue';

import { useI18n } from 'vue-i18n';
import { BillsExportResData } from '@/typings/bill';

/**
 * 导出按钮
 * @prop {Function} cb - 获取下载链接的回调函数
 * @prop {String} title - 弹窗标题
 * @prop {String} content - 弹窗内容
 */
export default defineComponent({
  props: { cb: Function as PropType<() => Promise<BillsExportResData>>, title: String, content: String },
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
        <Dialog v-model:isShow={isShow.value} title={props.title} quick-close dialogType='show'>
          {props.content}，{t('点击')}
          <a href={downloadUrl.value} download={props.title} class='text-link'>
            {' '}
            {t('链接下载')}
          </a>
          。
        </Dialog>
      </>
    );
  },
});
