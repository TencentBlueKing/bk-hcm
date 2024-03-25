import { defineComponent, ref } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { Switcher, Tag } from 'bkui-vue';
import {
  HOST_RUNNING_STATUS,
  HOST_SHUTDOWN_STATUS,
} from '@/views/resource/resource-manage/common/table/HostOperations';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import { CLOUD_HOST_STATUS } from '@/common/constant';

export default defineComponent({
  setup() {
    const detail = ref([]);

    const resourceFields = [
      {
        name: '名称',
        value: 'dommytest-20230403-0001',
        edit: true,
      },
      {
        name: '所属网络',
        value: 'dommytest-20230403-0001',
      },
      {
        name: 'ID',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '删除保护',
        render: () => (
          <div>
            <Switcher class={'mr5'} />
            <Tag>未开启</Tag>
          </div>
        ),
      },
      {
        name: '状态',
        render() {
          const data = {
            status: 'running',
          };
          return (
            <div>
              {HOST_SHUTDOWN_STATUS.includes(data.status) ? (
                data.status.toLowerCase() === 'stopped' ? (
                  <img src={StatusUnknown} class={'mr6'} width={14} height={14}></img>
                ) : (
                  <img src={StatusAbnormal} class={'mr6'} width={14} height={14}></img>
                )
              ) : HOST_RUNNING_STATUS.includes(data.status) ? (
                <img src={StatusNormal} class={'mr6'} width={14} height={14}></img>
              ) : (
                <img src={StatusUnknown} class={'mr6'} width={14} height={14}></img>
              )}
              <span>{CLOUD_HOST_STATUS[data.status] || data.status}</span>
            </div>
          );
        },
      },
      {
        name: 'IP版本',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '网络类型',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '创建时间',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '地域',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '可用区域',
        value: 'dommytest-20230403-0001',
      },
    ];

    const configFields = [
      {
        name: '负载均衡域名',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '实例计费模式',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '负载均衡VIP',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '带宽计费模式',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '规格类型',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '带宽上限',
        value: 'dommytest-20230403-0001',
      },
      {
        name: '运营商',
        value: 'dommytest-20230403-0001',
      },
    ];

    return () => (
      <div class={'clb-detail-continer'}>
        <div>
          <p class={'clb-detail-info-title'}>资源信息</p>
          <DetailInfo fields={resourceFields} detail={detail.value} class={'ml60'} />
        </div>
        <div>
          <p class={'clb-detail-info-title'}>配置信息</p>
          <DetailInfo fields={configFields} detail={detail.value} class={'ml60'} />
        </div>
      </div>
    );
  },
});
