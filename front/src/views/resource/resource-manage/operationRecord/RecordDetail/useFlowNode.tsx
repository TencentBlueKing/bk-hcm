import { timeFormatter } from '@/common/util';
import { Ref, VNode, ref, watchEffect } from 'vue';
import { Close, Spinner, Success } from 'bkui-vue/lib/icon';

export type IProps = Array<{
  id: string; // 任务ID
  name: string; // 任务名称
  state: NodeState; // 任务状态
  reviser: string; // 修改者
  updated_at: string; // 更新时间
}>;

export interface ITimelineNode {
  tag?: string;
  content?: string;
  icon?: string | VNode | boolean;
  theme?: '' | 'success' | 'danger';
}

export const FlowNodeNameMap = {
  tg_add_rs: '批量添加rs',
  tg_remove_rs: '批量移除rs',
  tg_modify_port: '批量修改端口',
  tg_modify_weight: '批量修改权重',
  apply_tg_listener_rule: '应用目标组到监听器/规则上',
};

export enum TaskState {
  pending = 'pending', // 等待中
  running = 'running', // 执行中
  rollback = 'rollback', // 回滚
  cancel = 'cancel', // 取消
  success = 'success', // 成功
  failed = 'failed', // 失败
}

export enum NodeState {
  pending = 'pending', // 等待中
  scheduled = 'scheduled', // 待调度
  running = 'running', // 执行中
  cancel = 'cancel', // 取消
  failed = 'failed', // 失败
  success = 'success', // 成功
}

export const useFlowNode = (props: IProps) => {
  const nodes: Ref<ITimelineNode[]> = ref([]);

  const getContent = (updated_at: string) => {
    return `<span style="font-size: 12px;color: #979BA5;">${timeFormatter(updated_at)}</span>`;
  };

  const renderIcon = (state: string) => {
    let icon = false as VNode | boolean;
    switch (state) {
      case NodeState.pending:
      case NodeState.scheduled:
      case NodeState.running:
        icon = <Spinner fill='#3A84FF' width={16} height={16} />;
        break;
      case NodeState.cancel:
      case NodeState.failed:
        icon = <Close fill='#EA3636' width={10.5} height={10.5} />;
        break;
      case NodeState.success:
        icon = <Success fill='#2DCB56' width={10.5} height={10.5} />;
        break;
    }
    return icon;
  };

  watchEffect(() => {
    if (!props.length) return;
    nodes.value = [
      {
        tag: '单据提交',
        content: '<span style="font-size: 12px;color: #979BA5;">2019-12-15 11:00</span>',
        icon: <Success fill='#2DCB56' width={10.5} height={10.5} />,
        theme: 'success',
      },
      ...props.map(({ name, state, updated_at }) => ({
        tag: FlowNodeNameMap[name] || '--',
        content: getContent(updated_at),
        icon: renderIcon(state),
      })),
      {
        tag: '<span>执行结束</span>',
      },
    ];
  });

  return {
    nodes,
  };
};
