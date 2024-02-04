<template>
  <bk-loading :opacity="1" :loading="loading">
    <div class="iam-apply-detail-wrapper">
      <template v-if="isShowPage">
        <BasicInfo :data="basicInfo" />
        <ProcessStatus :data="basicInfo" />
        <ApplyProcess :link="basicInfo.ticket_url" />
        <div class="action">
          <bk-button :loading="cancelLoading" @click="handleCancel" :disabled="isShowRevoke">
            {{ t('撤销') }}
          </bk-button>
          <!-- <bk-button
          :disabled="isCloneDisabled"
          :loading="loading"
          @click="handleCancel"
          style="margin-left: 20px">
          {{ t('克隆') }}
        </bk-button> -->
        </div>
      </template>
    </div>
  </bk-loading>
</template>

<script lang="ts">
import { defineComponent, ref, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import emptyChart from '@/assets/image/empty-chart.png';
import BasicInfo from '@/views/service/my-apply/components/basic-info/index.vue';
import ApplyProcess from '@/views/service/my-apply/components/apply-process/index.vue';
import ProcessStatus from '@/views/service/my-apply/components/process-status/index.vue';
export default defineComponent({
  name: 'ApplyDetail',
  components: {
    BasicInfo,
    ApplyProcess,
    ProcessStatus,
  },
  props: {
    params: {
      type: Object,
      default: () => {
        return {} as any;
      },
    },
    loading: {
      type: Boolean,
      default: false,
    },
    cancelLoading: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['on-cancel'],
  setup(props, { emit }) {
    const { t } = useI18n();

    const basicInfo = ref({} as any);
    const tableList = ref([]);
    const isShowPage = ref(true);

    const handleCancel = () => {
      emit('on-cancel', basicInfo.value.id);
    };

    // const handleClone = () => {
    //   console.log('克隆');
    // };

    watch(
      () => props.params,
      async (payload: Record<string, any>) => {
        if (Object.keys(payload).length) {
          basicInfo.value = { ...basicInfo.value, ...payload };
        } else {
        }
      },
      {
        immediate: true,
      },
    );

    const isShowRevoke = computed(() => {
      return !['pending', 'reject'].includes(basicInfo.value.status);
    });

    const isCloneDisabled = computed(() => {
      return !['pass', 'reject'].includes(basicInfo.value.status);
    });

    return {
      t,
      basicInfo,
      tableList,
      isShowPage,
      isShowRevoke,
      isCloneDisabled,
      emptyChart,
      handleCancel,
      ProcessStatus,
    };
  },
});
</script>

<style lang="scss" scoped>
@import './index.scss';
</style>
