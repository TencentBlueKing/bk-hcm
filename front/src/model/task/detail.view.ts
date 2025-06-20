import { h } from 'vue';
import { ModelProperty } from '@/model/typings';
import { TASK_DETAIL_STATUS_NAME } from '@/views/task/constants';
import { ITaskDetailItem } from '@/store/task';
import { TaskDetailStatus } from '@/views/task/typings';
import { timeFormatter } from '@/common/util';

export default [
  {
    id: 'task_management_id',
    name: '任务ID',
    type: 'string',
  },
  {
    id: 'created_at',
    name: '开始时间',
    type: 'datetime',
  },
  {
    id: 'updated_at',
    name: '结束时间',
    type: 'datetime',
    render: ({ row }: { row: ITaskDetailItem }) =>
      h(
        'span',
        [TaskDetailStatus.INIT, TaskDetailStatus.RUNNING].includes(row.state) ? '--' : timeFormatter(row.updated_at),
      ),
  },
  {
    id: 'state',
    name: '任务状态',
    type: 'enum',
    option: TASK_DETAIL_STATUS_NAME,
    meta: {
      display: {
        appearance: 'status',
      },
    },
  },
  {
    id: 'reason',
    name: '失败原因',
    type: 'string',
  },
  {
    id: 'param.clb_vip_domain',
    name: 'CLB VIP/域名',
    type: 'string',
  },
  {
    id: 'param.cloud_clb_id',
    name: 'CLB ID',
    type: 'string',
  },
  {
    id: 'param.cloud_lb_id',
    name: 'CLB ID',
    type: 'string',
  },
  {
    id: 'param.protocol',
    name: '协议',
    type: 'string',
  },
  {
    id: 'param.listener_port',
    name: '监听器端口',
    type: 'array',
  },
  {
    id: 'param.port',
    name: '监听器端口',
    type: 'string',
  },
  {
    id: 'param.domain',
    name: '域名',
    type: 'string',
  },
  {
    id: 'param.url_path',
    name: 'URL',
    type: 'string',
  },
  {
    id: 'param.health_check',
    name: '健康检查',
    type: 'bool',
    option: {
      trueText: '开启',
      falseText: '关闭',
    },
  },
  {
    id: 'param.scheduler',
    name: '均衡方式',
    type: 'enum',
    option: {
      WRR: '按权重轮询',
      LEAST_CONN: '最小连接数',
      IP_HASH: 'IP Hash',
    },
  },
  {
    id: 'param.session',
    name: '会话保持',
    type: 'number',
  },
  {
    id: 'param.ssl_mode',
    name: '证书认证方式',
    type: 'enum',
    option: {
      UNIDIRECTIONAL: '单向认证',
      MUTUAL: '双向认证',
    },
  },
  {
    id: 'param.cert_cloud_ids',
    name: '服务器证书',
    type: 'cert',
  },
  {
    id: 'param.ca_cloud_id',
    name: 'CA证书',
    type: 'ca',
  },
  {
    id: 'param.rs_ip',
    name: 'RSIP',
    type: 'string',
  },
  {
    id: 'param.inst_type',
    name: 'RS类型',
    type: 'enum',
    option: {
      CVM: 'CVM',
      ENI: 'ENI',
    },
  },
  {
    id: 'param.rs_port',
    name: 'RSPORT',
    type: 'array',
  },
  {
    id: 'param.weight',
    name: 'RS权重',
    type: 'number',
  },
  {
    id: 'param.validate_result',
    name: '参数校验',
    type: 'array',
  },
  {
    id: 'param.rs_list',
    name: 'rs信息',
    type: 'json',
  },
  {
    id: 'param.ip',
    name: 'RSIP',
    type: 'string',
  },
  {
    id: 'param.weight',
    name: '原权重',
    type: 'number',
  },
  {
    id: 'param.new_rs_weight',
    name: '新权重',
    type: 'number',
  },
  {
    id: 'param.url',
    name: 'URL',
    type: 'string',
  },
] as ModelProperty[];
