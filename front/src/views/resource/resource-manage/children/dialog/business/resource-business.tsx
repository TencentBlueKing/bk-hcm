import {
  defineComponent,
  ref,
  watch,
} from 'vue';
// import {
//   useI18n,
// } from 'vue-i18n';
import { useAccountStore } from '@/store';
import StepDialog from '@/components/step-dialog/step-dialog';
import './resource-business.scss';

export default defineComponent({
  components: {
    StepDialog,
  },

  props: {
    title: {
      type: String,
    },
    isShow: {
      type: Boolean,
    },
  },

  emits: ['update:isShow', 'handle-confirm'],

  setup(props, { emit }) {
    const business = ref([]);
    const businessList = ref([]);
    const accountStore = useAccountStore();

    // use hooks
    // const {
    //   t,
    // } = useI18n();

    watch(
      () => props.isShow,
      (val) => {
        if (val) {
          getBusinessList();
        }
      },
    );

    // 状态
    const steps = [
      {
        component: () => <>
          <bk-select
            v-model={business.value}
            class="business-select"
            filterable
          >
            {
              businessList.value.map((item, index) => <>
                <bk-option
                  key={index}
                  value={item.id}
                  label={item.name}
                />
              </>)
            }
          </bk-select>
          {/* {
            `${t('共转移') + business.value.length + t('个')}`
          } */}
        </>,
      },
    ];

    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      emit('handle-confirm', business.value);
      handleClose();
    };

    const getBusinessList = async () => {
      try {
        const res = await accountStore.getBizList();
        console.log(res);
        businessList.value = res?.data;
      } catch (error) {
        console.log(error);
      }
    };

    return {
      steps,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    return <>
      <step-dialog
        size="normal"
        title={this.title}
        isShow={this.isShow}
        steps={this.steps}
        onConfirm={this.handleConfirm}
        onCancel={this.handleClose}
      >
      </step-dialog>
    </>;
  },
});
