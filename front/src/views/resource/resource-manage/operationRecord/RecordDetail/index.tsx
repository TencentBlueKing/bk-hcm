import { defineComponent } from 'vue';
import './index.scss';
import DetailHeader from '../../common/header/detail-header';
import { useRoute } from 'vue-router';
import { Close, Spinner, Success } from 'bkui-vue/lib/icon';
import { Button, TimeLine } from 'bkui-vue';
import { useTable } from '@/hooks/useTable/useTable';

export default defineComponent({
  setup() {
    const route = useRoute();
    const { CommonTable } = useTable({
      searchOptions: {
        searchData: [
          {
            name: '内网IP',
            id: 'intranetIp',
          },
          {
            name: '公网IP',
            id: 'internetIp',
          },
          {
            name: '主机名称',
            id: 'hostName',
          },
          {
            name: '地域',
            id: 'region',
          },
          {
            name: '可用区',
            id: 'availabilityZone',
          },
          {
            name: '机型',
            id: 'machineType',
          },
          {
            name: '操作系统',
            id: 'operatingSystem',
          },
        ],
      },
      tableOptions: {
        columns: [
          {
            label: '内网IP',
            field: 'intranetIp',
          },
          {
            label: '公网IP',
            field: 'internetIp',
          },
          {
            label: '主机名称',
            field: 'hostName',
          },
          {
            label: '地域',
            field: 'region',
          },
          {
            label: '可用区',
            field: 'availabilityZone',
          },
          {
            label: '机型',
            field: 'machineType',
          },
          {
            label: '操作系统',
            field: 'operatingSystem',
          },
        ],
      },
      requestOption: {
        type: 'audits/async_task',
      },
    });
    return () => (
      <div class={'record-detail-container'}>
        <DetailHeader>
          <span class={'header-title'}>操作记录详情</span>
          <span class={'header-content'}>&nbsp;- {route.query.name}</span>
        </DetailHeader>
        <div class={'record-detail-info-card'}>
          <Success width={21} height={21} fill='#FFB848' />
          <span class={'info-card-prefix'}>部分执行成功</span>
          <span class={'info-card-num'}>80 / 100</span>
          <span class={'info-card-content'}>
            执行分为 <span class={'info-card-highlight-num'}> 4 </span> 个批次，可在每个批次查看具体状态
          </span>
          <Button class={'info-card-btn'}>终止任务</Button>
        </div>
        <div class={'main-wrapper'}>
          <div class={'main-side-card'}>
            <p class={'main-side-card-title'}>执行步骤</p>
            <TimeLine
              class={'main-side-card-timeline'}
              list={[
                {
                  tag: '单据提交',
                  content: '<span style="font-size: 12px;color: #979BA5;">2019-12-15 11:00</span>',
                  icon: <Success fill='#2DCB56' width={10.5} height={10.5} />,
                  theme: 'success',
                },
                {
                  tag: '<span style="color: #EA3636">第一批任务</span>',
                  content: '<span style="font-size: 12px;color: #979BA5;">2019-12-15 11:00<br/>0 / 100</span>',
                  icon: <Close fill='#EA3636' width={10.5} height={10.5} />,
                  theme: 'danger',
                },
                {
                  tag: '<span style="font-weight: 700;font-size: 16px;color: #313238;">第三批任务</span>',
                  content: '<span style="font-size: 12px;">2020-12-15 11:00</span>',
                  icon: <Spinner fill='#3A84FF' width={16} height={16} />,
                },
                {
                  tag: '<span>执行结束</span>',
                },
              ]}></TimeLine>
          </div>
          <div class={'mian-list-card'}>
            <CommonTable />
          </div>
        </div>
      </div>
    );
  },
});
