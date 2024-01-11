import { Button, Dialog, Form, Input, Table } from 'bkui-vue';
import { PropType, defineComponent, ref } from 'vue';
import './index.scss';
const { FormItem } = Form;

export default defineComponent({
  props: {
    isShow: {
      required: true,
      type: Boolean,
    },
    handleClose: {
      required: true,
      type: Function as PropType<() => void>,
    },
  },
  setup(props) {
    const ipTableData = ref([
      {
        ipAddress: '192.168.1.1',
        note: '主服务器',
        actions: '修改',
      },
      {
        ipAddress: '192.168.1.2',
        note: '备份服务器',
        actions: '检查',
      },
      {
        ipAddress: '192.168.1.3',
        note: '数据库服务器',
        actions: '重启',
      },
    ]);
    return () => (
      <Dialog
        isShow={props.isShow}
        onClosed={() => props.handleClose()}
        onConfirm={() => props.handleClose()}
        title='新建参数模板'
        maxHeight={'720px'}
        width={1000}>
        <Form>
          <FormItem label='参数模板名称'>
            <Input placeholder='输入参数模板名称' />
          </FormItem>
          <FormItem label='参数模板类型'>
            <Input placeholder='输入参数模板名称' />
          </FormItem>
        </Form>
        <Table
          columns={[
            {
              label: 'IP地址',
              field: 'ipAddress',
              render: () => <Input placeholder='输入IP地址' />,
            },
            {
              label: '备注',
              field: 'note',
              render: () => <Input placeholder='备注信息' />,
            },
            {
              label: '操作',
              field: 'actions',
              render: ({ index }: {index: number}) => (
                <div>
                  <Button
                    text
                    class={'ml6'}
                    theme='primary'
                    onClick={() => {
                      ipTableData.value.splice(index, 1);
                      console.log(index, ipTableData.value);
                    }}>
                    删除
                  </Button>
                </div>
              ),
            },
          ]}
          data={ipTableData.value}
        />
        <Button
          text
          theme='primary'
          class={'mt20'}
          onClick={() => {
            ipTableData.value.push({
              ipAddress: '192.168.1.3',
              note: '数据库服务器',
              actions: '重启',
            });
          }}>
          新增一行
        </Button>
      </Dialog>
    );
  },
});
