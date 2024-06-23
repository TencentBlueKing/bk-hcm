import { InfoBox } from 'bkui-vue';
import { VNode } from 'vue';
import cssModule from './index.module.scss';

export const confirmInstance = InfoBox({
  isShow: false,
});
const Confirm = (title: string, content: string | VNode, onConfirm: () => void, onClosed?: () => void) => {
  confirmInstance.update({
    title,
    subTitle: content,
    onConfirm,
    onClosed,
    extCls: cssModule.confirm,
  });
  confirmInstance.show();
};
export default Confirm;
