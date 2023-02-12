<template>
  <div class="iam-apply-detail-wrapper">
    <template v-if="isShowPage">
      <BasicInfo :data="basicInfo" />
      <ApplyProcess :link="basicInfo.ticket_url" />
      <div class="action">
        <bk-button
          :loading="loading"
          @click="handleCancel"
          :disabled="isShowRevoke"
        >
          {{ t('撤销') }}
        </bk-button>
        <bk-button
          :disabled="isCloneDisabled"
          :loading="loading"
          @click="handleCancel"
          style="margin-left: 20px">
          {{ t('克隆') }}
        </bk-button>
      </div>
      <template v-if="isEmpty">
        <div class="apply-content-empty-wrapper">
          <img :src="emptyChart" />
        </div>
      </template>
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent, reactive, ref, toRefs, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import emptyChart from '@/assets/image/empty-chart.png';
import BasicInfo from '@/views/service/my-apply/components/basic-info/index.vue';
import ApplyProcess from '@/views/service/my-apply/components/apply-process/index.vue';
export default defineComponent({
  name: 'ApplyDetail',
  components: {
    BasicInfo,
    ApplyProcess,
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
  },
  emits: ['on-cancel'],
  setup(props, { emit }) {
    const { t } = useI18n();

    let detailInfo = reactive({
      initRequestQueue: ['detail'],
      status: '',
    });
    const basicInfo = ref({} as any);
    const tableList = ref([]);
    const isEmpty = ref(false);
    const isLoading = ref(false);
    const isShowPage = ref(true);

    const handleCancel = () => {
      emit('on-cancel');
    };

    const handleClone = () => {
      console.log('克隆');
    };


    watch(() => props.params, async (payload: Record<string, any>) => {
      if (Object.keys(payload).length) {
        basicInfo.value = { ...basicInfo.value, ...payload };
        detailInfo = Object.assign(detailInfo, { initRequestQueue: ['detail'], status: payload.status });
      } else {
        detailInfo = Object.assign(detailInfo, {
          initRequestQueue: [],
          status: '',
        });
        basicInfo.value = {};
        tableList.value = [];
        isShowPage.value = false;
      }
    }, {
      immediate: true,
    });

    const isShowRevoke = computed(() => {
      return !['pending', 'reject'].includes(detailInfo.status);
    });

    const isCloneDisabled = computed(() => {
      return !['pass', 'reject'].includes(detailInfo.status);
    });


    return {
      t,
      ...toRefs(detailInfo),
      basicInfo,
      tableList,
      isShowPage,
      isEmpty,
      isLoading,
      isShowRevoke,
      isCloneDisabled,
      emptyChart,
      handleCancel,
      handleClone,
    };
  },
});
</script>

<style lang="scss" scoped>
@import "./index.scss"
</style>
