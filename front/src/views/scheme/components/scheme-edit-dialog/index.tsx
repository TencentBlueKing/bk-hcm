import { defineComponent, ref, watch } from 'vue';
import { Form, Dialog } from 'bkui-vue';
import { useAccountStore } from '@/store';
import './index.scss';

export default defineComponent({
  name: 'SchemeEditDialog',
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
  emits: ['update:show', 'confirm'],
  setup(props, ctx) {
    const accountStore = useAccountStore();

    const localVal = ref({
      id: '',
      name: '',
      bk_biz_id: -1,
    });
    const bizList = ref([]);
    const bizLoading = ref(false);
    const formRef = ref<InstanceType<typeof Form>>();
    const pending = ref(false);

    watch(
      () => props.show,
      (val) => {
        if (val) {
          const { id, name, bk_biz_id } = props.schemeData;
          localVal.value = { id, name, bk_biz_id };
          pending.value = false;
          getBizList();
        }
      },
    );

    const getBizList = async () => {
      bizLoading.value = true;
      const res = await accountStore.getBizListWithAuth();
      bizList.value = res.data;
      bizLoading.value = false;
    };

    const handleConfirm = async () => {
      await formRef.value.validate();
      pending.value = true;
      const res = await props.confirmFn(localVal.value);
      pending.value = false;
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
        isLoading={pending.value}
        onConfirm={handleConfirm}
        onClosed={handleClose}>
        <bk-form
          ref={formRef}
          form-type='vertical'
          model={localVal.value}
          rules={{
            name: [
              {
                trigger: 'change',
                message: '方案名称不能为空',
                validator: (val: string) => val.trim().length,
              },
            ],
          }}>
          <bk-form-item label='方案名称' property='name' required={true}>
            <bk-input v-model={localVal.value.name} />
          </bk-form-item>
          {/* <bk-form-item label="项目标签">
            <bk-select v-model={localVal.value.bk_biz_id} loading={bizLoading.value}>
              {bizList.value.map((item) => {
                return (<bk-option key={item.id} value={item.id} label={item.name} />);
              })}
            </bk-select>
          </bk-form-item> */}
        </bk-form>
      </Dialog>
    );
  },
});
