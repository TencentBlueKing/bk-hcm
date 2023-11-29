import { defineComponent, ref, watch } from 'vue';
import { Form, Dialog } from 'bkui-vue'
import { useAccountStore } from '@/store';
import './index.scss';

export default defineComponent({
  name: 'scheme-edit-dialog',
  emits: ['update:show', 'confirm'],
  props: {
    schemeData: {
      type: Object,
      default: () => ({
        id: '',
        name: '',
        bk_biz_id: 0,
      }),
    },
    show: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '编辑方案',
    },
    confirmFn: Function,
  },
  setup(props, ctx) {
    const accountStore = useAccountStore();

    const localVal = ref({
      id: '',
      name: '',
      bk_biz_id: 0,
    });
    const bizList = ref([]);
    const bizLoading = ref(false);
    const formRef = ref<InstanceType<typeof Form>>();

    const rules = [{
      name: [
        { trigger: 'blur', message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾' },
      ]
    }];

    watch(() => props.show, val => {
      if(val) {
        const { id, name, bk_biz_id } = props.schemeData.value;
        localVal.value = { id, name, bk_biz_id };
        getBizList();
      }
    });

    const getBizList = async () => {
      bizLoading.value = true;
      const res = await accountStore.getBizListWithAuth();
      bizList.value = res.data;
      bizLoading.value = false;
    };

    const handleConfirm = async() => {
      await formRef.value.validate();
      const res = props.confirmFn(localVal.value);
      ctx.emit('confirm', res);
    };
    const handleClose = () => {
      ctx.emit('update:show', false);
    };
    return () => (
      <Dialog
        isShow={props.show}
        title={props.title}
        width={480}
        quickClose={false}
        onConfirm={handleConfirm}
        onClosed={handleClose}>
        <bk-form ref={formRef} form-type="vertical" model={localVal.value}>
          <bk-form-item label="方案名称" property="name" required={true}>
            <bk-input v-model={localVal.value.name} />
          </bk-form-item>
          <bk-form-item label="项目标签">
            <bk-select v-model={localVal.value.bk_biz_id} loading={bizLoading.value}>
              {bizList.value.map(item => {
                return (<bk-option key={item.id} value={item.id} label={item.name} />)
              })}
            </bk-select>
          </bk-form-item>
        </bk-form>
      </Dialog>
    );
  },
});