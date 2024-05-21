import { timeFormatter } from '@/common/util';
import { Ref, VNode, h, ref, watchEffect } from 'vue';
import { Close, RightTurnLine, Spinner, Success } from 'bkui-vue/lib/icon';
import { Button, Message } from 'bkui-vue';
import './index.scss';
import { useBusinessStore } from '@/store';
import { useRoute } from 'vue-router';

export type IProps = {
  flow: Ref<Flow>;
  tasks: Ref<Task[]>;
};

export interface Flow {
  id?: string; // 任务ID
  name?: string; // 任务名称
  state?: string; // 任务状态
  reason?: any; // 任务失败原因
  creator?: string; // 任务创建者
  reviser?: string; // 任务最后一次修改的修改者
  created_at?: string; // 任务创建时间，标准格式：2006-01-02T15:04:05Z
  updated_at?: string; // 任务最后一次修改时间，标准格式：2006-01-02T15:04:05Z
}

export interface Task {
  id?: string; // 子任务自增ID
  action_id?: string; // 子任务ID
  action_name?: string; // 子任务名称
  state?: string; // 子任务状态
  reason?: any; // 子任务失败原因
  creator?: string; // 子任务创建者
  reviser?: string; // 子任务最后一次修改的修改者
  created_at?: string; // 子任务创建时间，标准格式：2006-01-02T15:04:05Z
  updated_at?: string; // 子任务最后一次修改时间，标准格式：2006-01-02T15:04:05Z
}

export interface IFlowInfo {
  name?: string; // 异步任务名称
  num?: number; // 执行批次
  actions?: string[]; // 所有执行批次的ID
  successNum?: number; // 成功的批次
}

export interface ITimelineNode {
  tag?: string | VNode;
  content?: string | VNode;
  icon?: string | VNode | boolean;
  theme?: '' | 'success' | 'danger';
  nodeType: string;
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
  cancel = 'canceled', // 取消
  success = 'success', // 成功
  failed = 'failed', // 失败
}

export enum NodeState {
  pending = 'pending', // 等待中
  scheduled = 'scheduled', // 待调度
  running = 'running', // 执行中
  cancel = 'canceled', // 取消
  failed = 'failed', // 失败
  success = 'success', // 成功
}

export const useFlowNode = (props: IProps) => {
  const nodes: Ref<ITimelineNode[]> = ref([]);
  const flowInfo: Ref<IFlowInfo> = ref({});
  const isRetryLoading = ref(false);
  const businessStore = useBusinessStore();
  const route = useRoute();

  const getContent = (updated_at: string, state: NodeState, taskID?: string) => {
    return (
      <span class={'use-flow-node-content'}>
        {timeFormatter(updated_at)}
        {[NodeState.cancel, NodeState.failed].includes(state) ? (
          <Button class={'ml6'} text onClick={() => handleRetry(taskID)} theme='primary'>
            <RightTurnLine width={12} height={12} fill='#3A84FF' />
            重试
          </Button>
        ) : null}
      </span>
    );
  };

  const renderTag = (state: NodeState, name: string, reason: string) => {
    let tag = <span>{name}</span>;
    if ([NodeState.cancel, NodeState.failed].includes(state)) {
      tag = (
        <span class={'use-flow-node-error-tag'}>
          {name}{' '}
          <i
            class={'icon hcm-icon bkhcm-icon-alert ml4'}
            v-bk-tooltips={{
              content: reason,
            }}
          />
        </span>
      );
    }
    return tag;
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
    if (!props.tasks.value.length) return;
    nodes.value = [
      {
        tag: '单据提交',
        content: getContent(props.flow.value.created_at, NodeState.success),
        icon: <Success fill='#2DCB56' width={10.5} height={10.5} />,
        theme: 'success',
        nodeType: 'vnode',
      },
      ...props.tasks.value.map(({ state, updated_at, reason, id }, idx) => ({
        tag: renderTag(state as NodeState, `第 ${idx + 1} 批任务`, reason),
        content: getContent(updated_at, state as NodeState, id),
        icon: renderIcon(state),
        nodeType: 'vnode',
      })),
      {
        tag: <span>执行结束</span>,
        icon: renderIcon(props.flow.value.state),
        nodeType: 'vnode',
      },
    ];

    flowInfo.value = {
      name: FlowNodeNameMap[props.flow.value.name],
      num: props.tasks.value.length,
      actions: props.tasks.value.map(({ action_id }) => action_id),
      successNum: props.tasks.value.filter(({ state }) => state === 'success').length,
    };
  });

  const handleRetry = async (taskID: string) => {
    isRetryLoading.value = true;
    await businessStore.retryAsyncTask({
      lb_id: route.query.res_id,
      flow_id: props.flow.value.id,
      task_id: taskID,
    });
    isRetryLoading.value = false;
    Message({
      content: '触发重试成功',
      theme: 'success',
    });
  };

  return {
    nodes,
    flowInfo,
  };
};
