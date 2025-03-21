<script setup lang="ts">
import { computed } from 'vue';
import { APPLICATION_STATUS_MAP } from '../../apply-list/constants';
import { ApplicationStatus, IApplicationDetail } from '../index';

import StatusUnknown from '@/assets/image/Status-unknown.png';
import share from 'bkui-vue/lib/icon/share';
import copyToClipboard from '@/components/copy-to-clipboard/index.vue';

const props = defineProps<{ applicationDetail: IApplicationDetail }>();

const status = computed(() => props.applicationDetail?.status ?? '');
const message = computed(() => props.applicationDetail?.delivery_detail ?? '');
</script>

<template>
  <div class="status">
    <!-- icon -->
    <bk-loading
      v-if="['pending', 'delivering'].includes(status)"
      style="transform: scale(0.5)"
      mode="spin"
      theme="primary"
      loading
    />
    <i v-else-if="['rejected'].includes(status)" class="hcm-icon bkhcm-icon-38moxingshibai-01" />
    <i v-else-if="['pass', 'completed'].includes(status)" class="hcm-icon bkhcm-icon-7chenggong-01" />
    <i v-else-if="['deliver_error'].includes(status)" class="hcm-icon bkhcm-icon-close-circle-fill"></i>
    <img v-else :src="StatusUnknown" :style="{ width: '22px' }" />
    <!-- name -->
    <span class="status-name">{{ APPLICATION_STATUS_MAP[status] }}</span>
    <!-- message -->
    <bk-overflow-title v-if="status !== ApplicationStatus.pending" type="tips" class="message">
      {{ message }}
    </bk-overflow-title>
    <copy-to-clipboard :content="message" class="ml8" />
    <bk-link class="link" theme="primary" :href="applicationDetail.ticket_url" target="_blank">
      <div class="flex-row align-items-center">
        ITSM单据
        <share class="font-small ml4" />
      </div>
    </bk-link>
  </div>
</template>

<style scoped lang="scss">
.hcm-icon {
  font-size: 21px;
  color: #3a84ff;
}

.bkhcm-icon-7chenggong-01 {
  color: #2dcb56;
}

.bkhcm-icon-38moxingshibai-01,
.bkhcm-icon-close-circle-fill {
  color: #cc4053;
}

.status {
  display: flex;
  align-items: center;
  padding: 0 24px;
  height: 52px;
  background-color: #fff;

  .status-name {
    flex-shrink: 0;
    margin-left: 8px;
    color: #313238;
  }

  .message {
    margin-left: 16px;
    max-width: 60%;
    color: $danger-color;
  }

  .link {
    margin-left: auto;
    flex-shrink: 0;
  }
}
</style>
