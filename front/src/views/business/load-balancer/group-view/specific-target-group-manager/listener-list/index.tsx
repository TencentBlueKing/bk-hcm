import { defineComponent } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import './index.scss';

export default defineComponent({
  name: 'ListenerList',
  setup() {
    const { columns, settings } = useColumns('targetGroupListener');
    const tableData = [
      {
        listener: 'HTTP Listener A',
        loadBalancer: 'Load Balancer 1',
        url: 'http://example.com',
        resourceType: 'VM',
        protocol: 'HTTP',
        port: '80',
        abnormalPortCount: '2',
        vpc: 'VPC-1',
        cloudProvider: 'AWS',
        region: 'us-east-1',
        availabilityZone: 'us-east-1a',
        ipAddressType: 'Public',
      },
      {
        listener: 'HTTPS Listener B',
        loadBalancer: 'Load Balancer 2',
        url: 'https://example.com',
        resourceType: 'Container',
        protocol: 'HTTPS',
        port: '443',
        abnormalPortCount: '0',
        vpc: 'VPC-2',
        cloudProvider: 'Azure',
        region: 'west-europe',
        availabilityZone: 'eu-west-3c',
        ipAddressType: 'Private',
      },
      {
        listener: 'TCP Listener C',
        loadBalancer: 'Load Balancer 3',
        url: 'tcp://example.org',
        resourceType: 'Bare-metal',
        protocol: 'TCP',
        port: '22',
        abnormalPortCount: '5',
        vpc: 'VPC-3',
        cloudProvider: 'GCP',
        region: 'asia-northeast1',
        availabilityZone: 'asia-northeast1-a',
        ipAddressType: 'Elastic',
      },
    ];
    const searchData = [
      {
        name: '绑定的监听器',
        id: 'listener',
      },
      {
        name: '关联的负载均衡',
        id: 'loadBalancer',
      },
      {
        name: '关联的URL',
        id: 'url',
      },
      {
        name: '资源类型',
        id: 'resourceType',
      },
      {
        name: '协议',
        id: 'protocol',
      },
      {
        name: '端口',
        id: 'port',
      },
      {
        name: '异常端口数',
        id: 'abnormalPortCount',
      },
      {
        name: '所在VPC',
        id: 'vpc',
      },
      {
        name: '云厂商',
        id: 'cloudProvider',
      },
      {
        name: '地域',
        id: 'region',
      },
      {
        name: '可用区域',
        id: 'availabilityZone',
      },
      {
        name: '资源类型',
        id: 'resourceType',
      },
      {
        name: 'IP地址类型',
        id: 'ipAddressType',
      },
    ];

    const { CommonTable } = useTable({
      searchOptions: {
        searchData,
      },
      tableOptions: {
        columns,
        reviewData: tableData,
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        type: '',
      },
    });
    return () => (
      <div class='listener-list-page'>
        <CommonTable></CommonTable>
      </div>
    );
  },
});
