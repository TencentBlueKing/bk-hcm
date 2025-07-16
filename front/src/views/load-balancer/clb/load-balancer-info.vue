<script setup lang="ts">
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';
import ClbInfo from '@/views/business/load-balancer/clb-view/specific-clb-manager/clb-detail';

interface IProps {
  lbId: string;
  details: ILoadBalancerDetails;
  uploadDetails: (id: string) => void;
}

defineProps<IProps>();
</script>

<template>
  <clb-info :id="lbId" :detail="details" :get-details="uploadDetails" />
</template>

<style scoped lang="scss"></style>
