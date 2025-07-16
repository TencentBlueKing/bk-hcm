<script setup lang="ts">
import StatusSuccess from '@/assets/image/success-account.png';
import StatusFailure from '@/assets/image/failed-account.png';
import StatusLoading from '@/assets/image/status_loading.png';
import { BINDING_STATUS_NAME, BindingStatus, LAYER_7_LISTENER_PROTOCOL, ListenerProtocol } from '../../constants';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps<{ value: BindingStatus; protocol: ListenerProtocol; isDomain?: boolean }>();

const { t } = useI18n();

const isLayer7ListenerProtocol = computed(() => LAYER_7_LISTENER_PROTOCOL.includes(props.protocol));
const isBinding = computed(() => props.value === BindingStatus.BINDING);
const icon = computed(() => {
  if (isBinding.value) return StatusLoading;
  return props.value === BindingStatus.SUCCESS ? StatusSuccess : StatusFailure;
});
</script>

<template>
  <template v-if="isLayer7ListenerProtocol || isDomain">
    <i
      class="hcm-icon bkhcm-icon-38moxingshibai-01 text-gray font-normal cursor mr8"
      v-bk-tooltips="{ content: t('HTTP/HTTPS监听器的同步状态，请到URL列表查看') }"
    ></i>
    <span>--</span>
  </template>
  <template v-else>
    <div v-if="value" class="binding-status-wrap">
      <img class="status-icon" :class="{ 'spin-icon': value === 'binding' }" :src="icon" alt="" />
      <span>{{ BINDING_STATUS_NAME[value] }}</span>
    </div>
    <template v-else>--</template>
  </template>
</template>

<style scoped lang="scss">
.binding-status-wrap {
  display: flex;
  align-items: center;

  .status-icon {
    margin-right: 7px;
    width: 14px;
    height: 14px;
  }
}
</style>
