import { h } from 'vue';
import { InfoBox } from 'bkui-vue';
import { useLoadBalancerStore } from '@/store';
import { VendorEnum } from '@/common/constant';
import classes from './export-infobox.module.scss';

type UseExportParams = {
  target: 'lb' | 'listener';
  vendor: VendorEnum;
  listeners: Array<{ lb_id: string; lbl_ids?: string[] }>;
  single?: { name: string };
};

export const useExport = (params: UseExportParams) => {
  const { vendor, target = 'lb', listeners = [], single } = params;

  const isLb = target === 'lb';

  const lbIds = listeners.map((item) => item.lb_id);
  const lblIds = listeners.flatMap((item) => item.lbl_ids);

  const { exportPreCheck, exportClb } = useLoadBalancerStore();

  const invokeExport = () => {
    const stats = () =>
      h(
        'div',
        { class: classes['infobox-stats'] },
        single
          ? ['名称：', single.name]
          : ['已选择', h('em', isLb ? lbIds.length : lblIds.length), isLb ? '个负载均衡' : '个监听器'],
      );

    const confirmTips = () =>
      h('div', { class: classes['infobox-tips'] }, [
        '导出数据包括',
        h('b', '监听器信息，URL规则信息(HTTP/HTTPS协议)，监听器绑定的RS信息。'),
      ]);

    const errorTips = (props?: { content: string }) => h('div', { class: classes['infobox-tips'] }, props.content);
    const loadingTips = () => h('div', '导出过程中请勿关闭本弹窗，或可直接终止导出');

    const infoBox = InfoBox({
      title: isLb ? `确认${!single ? '批量' : ''}导出负载均衡？` : `确认${!single ? '批量' : ''}导出监听器？`,
      width: 480,
      contentAlign: 'left',
      content: h('div', { class: classes['infobox-content'] }, [stats(), confirmTips()]),
      confirmText: '导出',
      cancelText: '取消',
      onConfirm: async () => {
        try {
          // 开始预检，只保留一个按钮，主按钮loading
          infoBox.update({
            cancelText: undefined,
          });
          const checkResult = await exportPreCheck(vendor, listeners);

          if (!checkResult.pass) {
            // 预检不通过，显示错误信息
            infoBox.update({
              type: 'danger',
              title: '批量导出失败',
              content: h('div', { class: [classes['infobox-content'], classes['infobox-content-error']] }, [
                stats(),
                errorTips({ content: checkResult.reason }),
              ]),
              confirmText: '知道了',
              cancelText: undefined,
              onConfirm: () => Promise.resolve(),
            });
          } else {
            // 预检通过，开始导出
            const { download, cancelDownload } = await exportClb(vendor, listeners);

            // 先显示loading
            infoBox.update({
              type: 'loading',
              title: isLb ? '批量导出负载均衡中…' : '批量导出监听器中…',
              contentAlign: 'center',
              content: h('div', { class: classes['infobox-content'] }, [loadingTips()]),
              confirmText: undefined,
              cancelText: '终止导出',
              onClose: () => {
                cancelDownload();
              },
            });

            // 执行下载
            download()
              .then(() => {
                infoBox.hide();
                infoBox.destroy();
              })
              .catch((error: any) => {
                if (error?.code === 'ERR_CANCELED') {
                  return;
                }
                infoBox.update({
                  type: 'danger',
                  title: '导出失败',
                  contentAlign: 'center',
                  content: h('div', { class: classes['infobox-content'] }, [error?.message]),
                  confirmText: '关闭',
                  cancelText: undefined,
                  onConfirm: () => Promise.resolve(),
                });
              });
          }
        } finally {
          // 组件在onConfirm时会自动默认关闭，这里通过reject取消自动关闭
          return Promise.reject();
        }
      },
    });
  };

  return {
    invokeExport,
  };
};
