import { defineComponent, PropType, ref } from 'vue';

import { Button } from 'bkui-vue';

import { useI18n } from 'vue-i18n';

/**
 * 导出按钮
 * @prop {Function} cb - 获取下载链接的回调函数
 * @prop {String} title - 弹窗标题
 * @prop {String} content - 弹窗内容
 */
export default defineComponent({
  props: { cb: Function as PropType<() => Promise<void>>, title: String, content: String },
  setup(props) {
    const { t } = useI18n();

    const isLoading = ref(false);

    const handleExport = async () => {
      isLoading.value = true;
      try {
        await props.cb();
      } finally {
        isLoading.value = false;
      }
    };

    return () => (
      <>
        <Button
          onClick={handleExport}
          loading={isLoading.value}
          v-bk-tooltips={{
            content: t('单次导出最多20万条账单，超出条数后，当前暂不支持导出。'),
            placement: 'top-start',
          }}>
          {t('导出')}
        </Button>
      </>
    );
  },
});
