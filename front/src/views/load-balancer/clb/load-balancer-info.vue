<script setup lang="ts">
import { inject, Ref } from 'vue';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';
import ClbInfo from '@/views/business/load-balancer/clb-view/specific-clb-manager/clb-detail';
import { IAuthSign } from '@/common/auth-service';

interface IProps {
  lbId: string;
  details: ILoadBalancerDetails;
  uploadDetails: (id: string) => void;
}

defineProps<IProps>();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');
const clbOperationAuthSign = inject<Ref<IAuthSign | IAuthSign[]>>('clbOperationAuthSign');
</script>

<template>
  <div class="clb-info-container">
    <clb-info
      :id="lbId"
      :current-global-business-id="currentGlobalBusinessId"
      :detail="details"
      :get-details="uploadDetails"
      :clb-operation-auth-sign="clbOperationAuthSign"
    />
  </div>
</template>

<style scoped lang="scss">
.clb-info-container {
  padding-top: 24px;
  height: 100%;
}
</style>
