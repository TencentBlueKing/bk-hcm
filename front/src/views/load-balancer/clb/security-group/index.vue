<script setup lang="ts">
import { inject, Ref } from 'vue';
import { IAuthSign } from '@/common/auth-service';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';

import SecurityGroup from '@/views/business/load-balancer/clb-view/specific-clb-manager/security-group';

interface IProps {
  lbId: string;
  details: ILoadBalancerDetails;
  uploadDetails: (id: string) => void;
}

defineProps<IProps>();
const clbOperationAuthSign = inject<Ref<IAuthSign | IAuthSign[]>>('clbOperationAuthSign');
</script>

<template>
  <security-group
    :id="lbId"
    :detail="details"
    :get-details="uploadDetails"
    :clb-operation-auth-sign="clbOperationAuthSign"
  />
</template>

<style scoped lang="scss"></style>
