import {
  defineComponent,
  ref,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
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

  setup(_, { emit }) {
    const business = ref([]);
    const businessList = ref([]);

    // use hooks
    const {
      t,
    } = useI18n();

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
                  value={item.value}
                  label={item.label}
                />
              </>)
            }
          </bk-select>
          {
            `${t('共转移') + business.value.length + t('个')}`
          }
        </>,
      },
    ];

    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      emit('handle-confirm');
      handleClose();
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
