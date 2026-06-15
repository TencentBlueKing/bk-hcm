import { bkTooltips } from 'bkui-vue';
import safeHtml from './safeHtml';
// import overflowTitle from './overflowTitle';
const directives: Record<string, any> = {
  // 指令对象
  bkTooltips,
  safeHtml,
  // overflowTitle,
};

export default {
  install(app: any) {
    Object.keys(directives).forEach((key) => {
      app.directive(key, directives[key]);
    });
  },
};
