<script setup lang="tsx">
import { exportTableToExcel } from '@/utils';
import { MAX_ROWS_PER_FILE } from '@/vendor/Export2Excel';
import { InfoBox, Message } from 'bkui-vue';
import { h, useCssModule } from 'vue';

interface ExportToExcelButtonProps {
  data?: any[];
  columns: any[];
  filename?: string;
  text?: string;
  name?: string;
  confirmTips?: string;
  content?: string;
  title?: string;
  confirmText?: string;
  cancelText?: string;
  showConfirmDialog?: boolean;
  disabled?: boolean;
  request?: (signal: AbortSignal) => Promise<any[]>;
  pickNum?: number;
  maxExportNum?: number;
}

defineOptions({ name: 'ExportToExcelBatchButton' });
const props = withDefaults(defineProps<ExportToExcelButtonProps>(), {
  filename: '导出文件',
  text: '导出',
  data: () => [],
  columns: () => [],
  confirmTips: '点击导出按钮后开始导出',
  confirmText: '导出',
  cancelText: '取消',
  name: '',
  showConfirmDialog: false,
  disabled: false,
  pickNum: 0,
  maxExportNum: 450000,
});

const emit = defineEmits<{
  confirm: [];
  fail: [error: Error];
  abort: [];
}>();

const classes = useCssModule();

// 用于取消导出操作的控制器
let abortController: AbortController | null = null;

// 同步导出
const exportToExcel = () => {
  exportTableToExcel(props.data, props.columns, props.filename).then(() => {
    Message({
      theme: 'success',
      message: '导出成功',
    });
  });
};

// 异步导出（弹窗流程）
const invokeExport = () => {
  const dataLength = props.data.length || props.pickNum;

  const overloadText = () => {
    return dataLength > MAX_ROWS_PER_FILE
      ? h('p', [
          `导出文件将拆分为 ${Math.ceil(
            dataLength / MAX_ROWS_PER_FILE,
          )} 个文件，单个最多为 ${MAX_ROWS_PER_FILE} 条记录`,
        ])
      : '';
  };

  const content = () =>
    h(
      'div',
      { class: classes['infobox-stats'] },
      props.content
        ? [props.content]
        : ['已选择', h('em', props.data.length || props.pickNum), `个${props.name}`, overloadText()],
    );

  const confirmTips = () => h('div', { class: classes['infobox-tips'] }, [props.confirmTips]);

  const errorTips = (content: string) => h('div', { class: classes['infobox-tips'] }, content);

  const loadingTips = () => h('div', '导出过程中请勿关闭本弹窗，或可直接终止导出');

  // 检查是否超过最大导出数量限制
  if (dataLength > props.maxExportNum) {
    InfoBox({
      type: 'danger',
      title: '批量导出失败',
      width: 480,
      contentAlign: 'left',
      content: h('div', { class: [classes['infobox-content'], classes['infobox-content-error']] }, [
        content(),
        errorTips(`导出数量已超过上限 ${props.maxExportNum} 条，请筛选条件后再提交导出`),
      ]),
      confirmText: '知道了',
      cancelText: undefined,
    });
    return;
  }

  const infoBox = InfoBox({
    title: props.title || `确认批量导出${props.name}记录？`,
    width: 480,
    contentAlign: 'left',
    content: h('div', { class: classes['infobox-content'] }, [content(), confirmTips()]),
    confirmText: props.confirmText,
    cancelText: props.cancelText,
    onConfirm: async () => {
      // 创建新的 AbortController
      abortController = new AbortController();
      const { signal } = abortController;

      try {
        // 开始导出，只保留一个按钮，主按钮loading
        infoBox.update({
          cancelText: undefined,
        });

        infoBox.update({
          type: 'loading',
          title: `批量导出中…`,
          contentAlign: 'center',
          content: h('div', { class: classes['infobox-content'] }, [loadingTips()]),
          confirmText: undefined,
          cancelText: '终止导出',
          closeIcon: false,
          onClose: () => {
            // 终止导出操作
            abortController?.abort();
            emit('abort');
          },
        });

        // 检查是否已被终止
        if (signal.aborted) {
          throw new Error('终止导出');
        }

        if (props.request) {
          const list = await props.request(signal);
          // 请求完成后再次检查是否被终止
          if (signal.aborted) {
            throw new Error('终止导出');
          }
          // 在开始文件下载前禁用终止按钮
          infoBox.update({
            cancelText: undefined,
          });
          await exportTableToExcel(list, props.columns, props.filename).then(() => {
            Message({
              theme: 'success',
              message: '导出成功',
            });
          });
        } else {
          await new Promise<void>((resolve, reject) => {
            const timeoutId = setTimeout(async () => {
              if (signal.aborted) {
                reject(new Error('终止导出'));
                return;
              }
              await exportTableToExcel(props.data, props.columns, props.filename).then(() => {
                Message({
                  theme: 'success',
                  message: '导出成功',
                });
              });

              resolve();
            }, 1000);

            // 监听终止信号
            const abortHandler = () => {
              clearTimeout(timeoutId);
              signal.removeEventListener('abort', abortHandler); // 使用相同的函数引用
              reject(new Error('终止导出'));
            };
            signal.addEventListener('abort', abortHandler);
          });
        }

        // 成功
        infoBox.hide();
        infoBox.destroy();
      } catch (error: any) {
        // 如果是终止导出，直接关闭弹窗
        if (signal.aborted || error?.message === '终止导出') {
          infoBox.hide();
          infoBox.destroy();
          return;
        }
        // 其他失败情况
        infoBox.update({
          type: 'danger',
          title: '批量导出失败',
          content: h('div', { class: [classes['infobox-content'], classes['infobox-content-error']] }, [
            content(),
            errorTips(error?.message),
          ]),
          confirmText: '知道了',
          cancelText: undefined,
          closeIcon: true,
          onConfirm: () => Promise.resolve(),
        });
      } finally {
        abortController = null;
        // 组件在onConfirm时会自动默认关闭，这里通过reject取消自动关闭
        return Promise.reject();
      }
    },
  });
};
// 点击处理
const handleClick = () => {
  if (props.showConfirmDialog) {
    invokeExport();
  } else {
    exportToExcel();
  }
};
</script>

<template>
  <bk-button :disabled="disabled" @click="handleClick">
    {{ props.text }}
  </bk-button>
</template>

<style lang="scss" module>
.infobox-stats {
  color: #313238;

  em {
    color: #3a84ff;
    font-style: normal;
    font-weight: 700;
    padding: 0.2em;
  }
}

.infobox-stats-error {
  em {
    color: #ea3636;
  }
}

.infobox-tips {
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 2px;
}

.infobox-content {
  display: grid;
  flex-direction: column;
  gap: 16px;
  color: #4d4f56;
  font-size: 14px;
}

.infobox-content-error {
  .infobox-stats {
    em {
      color: #ea3636;
    }
  }
}
</style>
