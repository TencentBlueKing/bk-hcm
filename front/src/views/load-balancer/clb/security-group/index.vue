<script setup lang="ts">
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';
import SecurityGroup from '@/views/business/load-balancer/clb-view/specific-clb-manager/security-group';

interface IProps {
  lbId: string;
  details: ILoadBalancerDetails;
  uploadDetails: (id: string) => void;
}

defineProps<IProps>();
</script>

<template>
  <SecurityGroup :id="lbId" :detail="details" :get-details="uploadDetails" />
</template>

<style scoped lang="scss"></style>
