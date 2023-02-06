import {
  defineComponent,
  ref,
} from 'vue';
import { Table, Input, Select, Button } from 'bkui-vue';
import { ACTION_STATUS, GCP_PROTOCOL_LIST } from '@/constants';
import Confirm from '@/components/confirm';
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

  emits: ['update:isShow', 'submit'],

  setup(props, { emit }) {
    const {
      t,
    } = useI18n();

    // 状态
    const tableData = ref<any>([{ priority: 100 }]);
    const columns: any[] = [
      { label: t('优先级'),
        field: 'priority',
        render: ({ data }: any) => <Input class="mt25" v-model={ data.priority }></Input>,
      },
      { label: t('策略'),
        field: 'action',
        render: ({ data }: any) => {
          return (
            <Select class="mt25" v-model={data.action}>
                {ACTION_STATUS.map(ele => (
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
                <Select v-model={data.protocol}>
                    {GCP_PROTOCOL_LIST.map(ele => (
                    <Option value={ele.id} label={ele.name} key={ele.id} />
                    ))}
                </Select>
                <Input v-model={ data.port }></Input>
                </>
          );
        },
      },
      // { label: t('源地址'),
      //   field: 'id',
      //   render: ({ data }: any) => {
      //     return (
      //             <>
      //             <Select v-model={data.policy}>
      //                 {POLICY_STATUS.map(ele => (
      //                 <Option value={ele.id} label={ele.name} key={ele.id} />
      //                 ))}
      //             </Select>
      //             <Input v-model={ data.id }></Input>
      //             </>
      //     );
      //   },
      // },
      { label: t('描述'),
        field: 'memo',
        render: ({ data }: any) => <Input class="mt25" v-model={ data.memo }></Input>,
      },
      { label: t('操作'),
        field: 'id',
        render: ({ data, row }: any) => {
          return (
                <div class="mt20">
                <Button text theme="primary" onClick={() => {
                  hanlerCopy(data);
                }}>{t('复制')}</Button>
                <Button text theme="primary" class="ml20" onClick={() => {
                  handlerDelete(data, row);
                }}>{t('删除')}</Button>
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
      emit('submit', tableData.value);
    };

    // 新增
    const handlerAdd = () => {
      tableData.value.push({});
    };

    // 删除
    const handlerDelete = (data: any, row: any) => {
      console.log('data', data);
      const index = row.__$table_row_index;
      Confirm('确定删除', '删除之后不可恢复', () => {
        tableData.value.splice(index, 1);
      });
    };

    // 复制
    const hanlerCopy = (data: any) => {
      tableData.value.push(data);
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

