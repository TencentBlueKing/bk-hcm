<script setup lang="ts">
import { computed } from 'vue';
import { ApplicationStatus, IApplicationDetail } from '../index';
import { LocationQueryRaw } from 'vue-router';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import share from 'bkui-vue/lib/icon/share';
import copyToClipboard from '@/components/copy-to-clipboard/index.vue';
import { APPLICATION_STATUS_MAP } from '@/views/ticket/constants';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_LOAD_BALANCER_OVERVIEW } from '@/constants/menu-symbol';
import qs from 'qs';

const props = defineProps<{ applicationDetail: IApplicationDetail }>();

const status = computed(() => props.applicationDetail?.status ?? '');
const message = computed(() => props.applicationDetail?.delivery_detail ?? '');

const toClbList = () => {
  const { delivery_detail, content } = props.applicationDetail;
  const { load_balancer_id } = JSON.parse(delivery_detail);
  const { bk_biz_id } = JSON.parse(content);
  if (!load_balancer_id || !bk_biz_id) return;

  const query: LocationQueryRaw = {
    [GLOBAL_BIZS_KEY]: bk_biz_id,
    filter: qs.stringify(
      {
        cloud_id: Array.isArray(load_balancer_id) ? load_balancer_id[0] : load_balancer_id,
      },
      {
        arrayFormat: 'comma',
        encode: false,
        allowEmptyArrays: true,
      },
    ),
  };
  routerAction.redirect({ name: MENU_BUSINESS_LOAD_BALANCER_OVERVIEW, query });
};
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
    <!-- link -->
    <template v-if="['pass', 'completed'].includes(status)">
      <share class="font-small ml8 to-clb-list" fill="#3a84ff" @click="toClbList" />
    </template>
    <!-- message -->
    <template v-if="status === ApplicationStatus.deliver_error">
      <bk-overflow-title type="tips" class="message">
        {{ message }}
      </bk-overflow-title>
      <copy-to-clipboard :content="message" class="ml8" />
    </template>
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

  .to-clb-list {
    cursor: pointer;
  }
}
</style>
