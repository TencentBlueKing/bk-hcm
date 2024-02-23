import { InfoBox } from 'bkui-vue';
import { VNode } from 'vue';
export const confirmInstance = InfoBox({
  isShow: false,
  headerAlign: 'left',
  footerAlign: 'left',
  contentAlign: 'left',
});
const Confirm = (title: string, content: string | VNode, onConfirm: () => void, onClosed?: () => void) => {
  confirmInstance.update({
    title,
    subTitle: content,
    onConfirm,
    onClosed,
    headerAlign: 'left',
    footerAlign: 'left',
    contentAlign: 'left',
  });
  confirmInstance.show();
};
export default Confirm;
