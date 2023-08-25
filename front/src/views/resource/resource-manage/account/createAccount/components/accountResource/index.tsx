import { Switcher, Table } from 'bkui-vue';
import { defineComponent } from 'vue';

const TEST_DATA = new Array(10).fill({
  name: '主机',
  type: '系统内置',
  opertaion: [
    '购买', '分配',
  ],
  num: 666,
});

const TEST_COLUMNS = [
  {
    label: '资源名称',
    field: 'name',
  },
  {
    label: '插件类型',
    field: 'type',
  },
  {
    label: '操作',
    field: 'opertaion',
    rendor: ({ data }: any) => data.operation.join(','),
  },
  {
    label: '资源数量',
    field: 'num',
  },
  {
    label: '是否接入',
    rowspan: 9,
    rendor: () => (
      <Switcher/>
    ),
  },
];

export default defineComponent({
  setup() {
    return () => (
      <div>
        <Table
          data={TEST_DATA}
          columns={TEST_COLUMNS}
          border={['row', 'outer']}
        >
        </Table>
      </div>
    );
  },
});
