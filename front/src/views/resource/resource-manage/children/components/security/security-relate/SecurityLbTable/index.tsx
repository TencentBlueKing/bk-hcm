import { defineComponent } from 'vue';
import { Loading } from 'bkui-vue';
import useRemoteTable from '@/hooks/useRemoteTable';

export default defineComponent({
  name: 'SecurityLbTable',
  setup() {
    // /api/v1/cloud/security_group/{id}/common/list
    // /api/v1/bizs/{bk_biz_id}/cloud/security_group/{id}/common/list

    const { isLoading } = useRemoteTable('');

    return () => (
      <Loading loading={isLoading.value} class='security-lb-table'>
        security-lb-table
      </Loading>
    );
  },
});
