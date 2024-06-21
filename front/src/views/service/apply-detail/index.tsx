import { defineComponent, ref, watch } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import ApplyDetail from '@/views/service/my-apply/components/apply-detail/index.vue';
import { useAccountStore } from '@/store';
import { useRoute } from 'vue-router';

export default defineComponent({
  setup() {
    const accountStore = useAccountStore();
    const isLoading = ref(false);
    const currentApplyData = ref({});
    const curApplyKey = ref('');
    const isCancelBtnLoading = ref(false);
    const route = useRoute();

    // 获取单据详情
    const getMyApplyDetail = async (id: string) => {
      isLoading.value = true;
      try {
        const res = await accountStore.getApplyAccountDetail(id);
        currentApplyData.value = res.data;
        curApplyKey.value = res.data.id;
      } finally {
        isLoading.value = false;
      }
    };

    // 撤销单据
    const handleCancel = async (id: string) => {
      isCancelBtnLoading.value = true;
      try {
        await accountStore.cancelApplyAccount(id);
        getMyApplyDetail(id);
      } finally {
        isCancelBtnLoading.value = false;
      }
    };

    watch(
      () => route.query.id,
      (id) => {
        if (id) {
          getMyApplyDetail(id as string);
        }
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class={'apply-detail-container'}>
        <DetailHeader>
          <span>申请单详情</span>
        </DetailHeader>
        <div class={'apply-content-wrapper'}>
          <ApplyDetail
            params={currentApplyData.value}
            loading={isLoading.value}
            key={curApplyKey.value}
            cancelLoading={isCancelBtnLoading.value}
            onCancel={handleCancel}
          />
        </div>
      </div>
    );
  },
});
