import {
  defineComponent,
  ref,
} from 'vue';
import { Table, Input, Select, Button } from 'bkui-vue';
import { POLICY_STATUS } from '@/constants';
import {
  useI18n,
} from 'vue-i18n';
import StepDialog from '@/components/step-dialog/step-dialog';
const { Option } = Select;

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

  emits: ['update:isShow'],

  setup(props, { emit }) {
    const {
      t,
    } = useI18n();

    // 状态
    const tableData = ref([{ id: 1, policy: 'reject' }]);
    const columns: any[] = [
      { label: t('优先级'),
        field: 'id',
        render: ({ data }: any) => <Input class="mt25" v-model={ data.id }></Input>,
      },
      { label: t('策略'),
        field: 'policy',
        render: ({ data }: any) => {
          return (
            <Select class="mt25" v-model={data.policy}>
                {POLICY_STATUS.map(ele => (
                <Option value={ele.id} label={ele.name} key={ele.id} />
                ))}
          </Select>
          );
        },
      },
      { label: t('协议端口'),
        field: 'port',
        render: ({ data }: any) => {
          return (
                <>
                <Select v-model={data.policy}>
                    {POLICY_STATUS.map(ele => (
                    <Option value={ele.id} label={ele.name} key={ele.id} />
                    ))}
                </Select>
                <Input v-model={ data.id }></Input>
                </>
          );
        },
      },
      { label: t('类型'),
        field: 'type',
        render: ({ data }: any) => {
          return (
              <Select class="mt25" v-model={data.policy}>
                  {POLICY_STATUS.map(ele => (
                  <Option value={ele.id} label={ele.name} key={ele.id} />
                  ))}
            </Select>
          );
        },
      },
      { label: t('源地址'),
        field: 'id',
        render: ({ data }: any) => {
          return (
                  <>
                  <Select v-model={data.policy}>
                      {POLICY_STATUS.map(ele => (
                      <Option value={ele.id} label={ele.name} key={ele.id} />
                      ))}
                  </Select>
                  <Input v-model={ data.id }></Input>
                  </>
          );
        },
      },
      { label: t('描述'),
        field: 'id',
        render: ({ data }: any) => <Input class="mt25" v-model={ data.id }></Input>,
      },
      { label: t('操作'),
        field: 'id',
        render: () => {
          return (
                <div class="mt20">
                <Button text theme="primary">{t('复制')}</Button>
                <Button text theme="primary" class="ml20">{t('删除')}</Button>
                </div>
          );
        },
      },
    ];
    const steps = [
      {
        component: () => <>
            <Table
              class="mt20"
              row-hover="auto"
              columns={columns}
              data={tableData.value}
            />
            <Button text theme="primary" class="ml20 mt20" onClick={handlerAdd}>{t('新增一条规则')}</Button>
          </>,
      },
    ];

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      handleClose();
    };

    const handlerAdd = () => {
      tableData.value.push({});
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

