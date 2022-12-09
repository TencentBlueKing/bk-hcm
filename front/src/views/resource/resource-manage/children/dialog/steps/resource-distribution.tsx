import {
  defineComponent,
  ref,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import './resource-distribution.scss';
import StepDialog from '@/components/step-dialog/step-dialog';
import {
  RESOURCE_TYPES,
} from '@/common/constant';

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
    hideRelateVPC: {
      type: Boolean,
    },
    chooseResourceType: {
      type: Boolean,
    },
  },

  emits: ['update:isShow'],

  setup(props, { emit }) {
    // use hooks
    const {
      t,
    } = useI18n();

    // 状态
    const business = ref('');
    const isDistributeVPC = ref(false);
    const businessList = ref([]);
    const tableData = ref([]);
    const resourceTypes = ref([]);
    const columns: any[] = [{ label: '23' }];
    const steps = [
      {
        title: t('业务与 VPC'),
        component: () => <>
          <section class="resource-head">
            <bk-select
              v-model={business.value}
              class="resource-select"
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
              !props.hideRelateVPC
                ? <bk-checkbox v-model={isDistributeVPC.value}>
                    { t('关联VPC同时分配') }
                  </bk-checkbox>
                : ''
            }
          </section>
          <bk-table
            class="mt20"
            row-hover="auto"
            columns={columns}
            data={tableData.value}
          />
        </>,
      },
      {
        title: t('信息确认'),
        component: () => <>
          <span class="confirm-message">
            <span>
              { `${t('目标业务')}:${business.value}` }
            </span>
            {
              !props.hideRelateVPC
                ? <span>
                    { `${t('关联VPC转移')}（${business.value}）${t('个')}` }
                  </span>
                : ''
            }
          </span>
          <bk-table
            class="mt20"
            row-hover="auto"
            columns={columns}
            data={tableData.value}
          />
        </>,
      },
    ];

    // 快速分配的时候需要选择资源类型
    if (props.chooseResourceType) {
      steps.unshift({
        title: t('资源类型'),
        component: () => <>
          <section>
            <span>分配资源类型</span>
            <bk-checkbox-group
              class="resource-types"
              v-model={resourceTypes.value}
            >
              {
                RESOURCE_TYPES.map((type) => {
                  return <bk-checkbox label={type.type}>{ t(type.name) }</bk-checkbox>;
                })
              }
            </bk-checkbox-group>
          </section>
        </>,
      });
    }

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
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
