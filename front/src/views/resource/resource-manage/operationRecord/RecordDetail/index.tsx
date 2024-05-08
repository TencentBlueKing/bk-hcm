import { defineComponent, ref, watch } from 'vue';
import './index.scss';
import DetailHeader from '../../common/header/detail-header';
import { useRoute } from 'vue-router';
import { Success } from 'bkui-vue/lib/icon';
import { Button, Select, Tab, TimeLine } from 'bkui-vue';
import { useTable } from '@/hooks/useTable/useTable';
import { useBusinessStore } from '@/store';
import { useFlowNode } from './useFlowNode';
import { useWhereAmI } from '@/hooks/useWhereAmI';
const { Option } = Select;
const { TabPanel } = Tab;

export default defineComponent({
  setup() {
    const route = useRoute();
    const { isResourcePage } = useWhereAmI();
    const businessStore = useBusinessStore();
    const tasks = ref([]);
    const flow = ref({});
    const isEnd = ref(false);
    const { nodes, flowInfo } = useFlowNode({
      flow,
      tasks,
    });
    const activeTab = ref('result');
    const actionId = ref('1');
    const isRetryLoading = ref(false);
    const { CommonTable } = useTable({
      searchOptions: {
        searchData: [
          {
            name: '内网IP',
            id: 'private_ip_addresses',
          },
          {
            name: '公网IP',
            id: 'public_ip_addresses',
          },
          {
            name: '主机名称',
            id: 'inst_name',
          },
          {
            name: '可用区',
            id: 'zone',
          },
          {
            name: '机型',
            id: 'inst_type',
          },
        ],
      },
      tableOptions: {
        columns: [
          {
            label: '内网IP',
            field: 'private_ip_addresses',
            render({ data }: any) {
              return <div>{data.private_ip_address}</div>;
            },
          },
          {
            label: '公网IP',
            field: 'public_ip_addresses',
            render({ data }: any) {
              return <div>{data.public_ip_address}</div>;
            },
          },
          {
            label: '主机名称',
            field: 'inst_name',
          },
          {
            label: '可用区',
            field: 'zone',
          },
          {
            label: '机型',
            field: 'inst_type',
          },
        ],
      },
      requestOption: {
        type: 'audits/async_task',
        extension: {
          flow_id: route.query.flow,
          audit_id: +route.query.id,
          action_id: actionId.value || flowInfo.value.actions?.[0] || '1',
        },
        dataPath: 'data.tasks[0].params.targets',
        async resolvePaginationCountCb(countData: any) {
          return countData.tasks?.[0].params.targets.length;
        },
      },
    });

    const getFlow = async (auditId: string, flowId: string) => {
      const res = await businessStore.getAsyncFlowList({
        audit_id: +auditId,
        flow_id: flowId,
      });
      flow.value = res.data.flow;
      tasks.value = res.data.tasks;
    };

    watch(
      () => [route.query.id, route.query.flow],
      ([id, flow]) => {
        getFlow(id as string, flow as string);
      },
      {
        immediate: true,
      },
    );

    const handleRetryTask = async () => {
      isRetryLoading.value = true;
      try {
        await businessStore.excuteTask({
          lb_id: route.query.res_id as string,
          flow_id: flow.value.id,
        });
        getFlow(route.query.id as string, route.query.flow as string);
      } finally {
        isRetryLoading.value = false;
      }
    };

    const handleEndTask = async () => {
      isRetryLoading.value = true;
      try {
        await businessStore.endTask({
          lb_id: route.query.res_id as string,
          flow_id: flow.value.id,
        });
        getFlow(route.query.id as string, route.query.flow as string);
      } finally {
        isRetryLoading.value = false;
      }
    };

    return () => (
      <div class={'record-detail-container'}>
        <DetailHeader>
          <span class={'header-title'}>操作记录详情</span>
          <span class={'header-content'}>&nbsp;- {flowInfo.value.name}</span>
        </DetailHeader>
        <div class={'record-detail-info-card'} style={isResourcePage ? { margin: '52px 0 0' } : null}>
          <Success
            width={21}
            height={21}
            fill={flowInfo.value.successNum === flowInfo.value.num ? '#2DCB56' : '#FFB848'}
          />
          <span class={'info-card-prefix'}>
            {flowInfo.value.successNum === flowInfo.value.num ? '全部执行成功' : '部分执行成功'}
          </span>
          <span class={'info-card-num'}>
            {flowInfo.value.successNum} / {flowInfo.value.num}
          </span>
          <span class={'info-card-content'}>
            执行分为 <span class={'info-card-highlight-num'}> {flowInfo.value.num} </span>{' '}
            个批次，可在每个批次查看具体状态
          </span>
          <Button
            loading={isRetryLoading.value}
            class={'info-card-btn'}
            onClick={() => {
              isEnd.value = !isEnd.value;
              if (!isEnd.value) handleRetryTask();
              else handleEndTask();
            }}
            theme={isEnd.value ? 'primary' : null}>
            {isEnd.value ? '重新执行' : '终止任务'}
          </Button>
        </div>
        <div
          class={'main-wrapper'}
          style={isResourcePage ? { margin: '16px 0 0', height: 'calc(100% - 120px)' } : null}>
          <div class={'main-side-card'}>
            <p class={'main-side-card-title'}>执行步骤</p>
            <TimeLine class={'main-side-card-timeline'} list={nodes.value}></TimeLine>
          </div>
          <Tab type='card-grid' class={'mian-list-card'} v-model:active={activeTab.value}>
            <TabPanel name={'result'} label={'执行结果'}>
              <CommonTable>
                {{
                  operation: () => (
                    <Select v-model={actionId.value} clearable={false}>
                      {flowInfo.value.actions?.map((id) => (
                        <Option name={`第${id}批`} id={id} key={id}></Option>
                      ))}
                    </Select>
                  ),
                }}
              </CommonTable>
            </TabPanel>
          </Tab>
        </div>
      </div>
    );
  },
});
