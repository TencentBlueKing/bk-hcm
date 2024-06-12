import { defineComponent, ref } from 'vue';
import { Dialog, Select, Upload } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  name: 'ImportBillDetailDialog',
  setup(_, { expose }) {
    const isShow = ref(false);

    const triggerShow = (v: boolean) => {
      isShow.value = v;
    };

    expose({ triggerShow });

    return () => (
      <Dialog v-model:isShow={isShow.value} title='导入' class='import-bill-detail-dialog' width={640}>
        <div class='flex-row mb30'>
          <div class='mr24'>云厂商</div>
          <div>zenlayer</div>
        </div>
        <div class='mb30'>
          <div class='mb6'>核算月份</div>
          <Select />
        </div>
        <div class='filter-upload-wrap'>
          <div class='mb6'>文件上传</div>
          <Upload />
        </div>
      </Dialog>
    );
  },
});
