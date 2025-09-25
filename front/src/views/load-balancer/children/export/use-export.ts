import { h } from 'vue';
import { InfoBox } from 'bkui-vue';
import { useLoadBalancerStore } from '@/store';
import { VendorEnum } from '@/common/constant';
import classes from './export-infobox.module.scss';

type UseExportParams = {
  target: 'lb' | 'listener' | 'rs';
  vendor: VendorEnum;
  listeners: Array<{ lb_id: string; lbl_ids?: string[] }>;
  single?: { name: string };
  onlyExportListener?: boolean; // 是否只导出监听器信息
  check?: boolean; // 是否需要预检
  targetIds?: string[]; // target为Rs时，此项必填
};

const targetName: { lb: string; listener: string; rs: string } = {
  lb: '负载均衡',
  listener: '监听器',
  rs: 'RS',
};

export const useExport = (params: UseExportParams) => {
  const {
    vendor,
    target = 'lb',
    listeners = [],
    targetIds = [],
    single,
    onlyExportListener = false,
    check = true,
  } = params;

  const isLb = target === 'lb';
  const isRs = target === 'rs';

  const lbIds = listeners.map((item) => item.lb_id);
  const lblIds = listeners.flatMap((item) => item.lbl_ids);
  const checkedLength = () => {
    if (isLb) return lbIds.length;
    if (isRs) return targetIds.length;
    return lblIds.length;
  };

  const { exportPreCheck, exportClb, exportRS } = useLoadBalancerStore();

  const invokeExport = () => {
    const stats = () =>
      h(
        'div',
        { class: classes['infobox-stats'] },
        single ? ['名称：', single.name] : ['已选择', h('em', checkedLength()), `个${targetName[target]}`],
      );

    const confirmTips = () =>
      h(
        'div',
        { class: classes['infobox-tips'] },
        onlyExportListener || isRs
          ? ['导出时间可能较长，点击导出按钮后开始导出']
          : ['导出数据包括', h('b', '监听器信息，URL规则信息(HTTP/HTTPS协议)，监听器绑定的RS信息。')],
      );

    const errorTips = (props?: { content: string }) => h('div', { class: classes['infobox-tips'] }, props.content);
    const loadingTips = () => h('div', '导出过程中请勿关闭本弹窗，或可直接终止导出');

    const infoBox = InfoBox({
      title: `确认${!single ? '批量' : ''}导出${targetName[target]}？`,
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
          const checkResult = check ? await exportPreCheck(vendor, listeners) : { pass: true, reason: '' };

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
            const apiMethod = isRs ? exportRS : exportClb;
            const { download, cancelDownload } = await apiMethod(
              vendor,
              isRs ? targetIds : listeners,
              onlyExportListener,
            );

            // 先显示loading
            infoBox.update({
              type: 'loading',
              title: `批量导出${targetName[target]}中…`,
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
