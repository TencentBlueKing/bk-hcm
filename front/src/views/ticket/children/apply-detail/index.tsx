import { computed, defineComponent, onUnmounted, ref, watch } from 'vue';
import './index.scss';
import { useAccountStore } from '@/store';
import { useRoute } from 'vue-router';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { APPLICATION_TYPE_MAP } from '@/views/ticket/constants';
import Clb from './clb.vue';
import { applyContentRender } from './apply-content-render.plugin';

export enum ApplicationStatus {
  pending = 'pending',
  pass = 'pass',
  rejected = 'rejected',
  cancelled = 'cancelled',
  delivering = 'delivering',
  completed = 'completed',
  deliver_partial = 'deliver_partial',
  deliver_error = 'deliver_error',
}

export interface IApplicationDetail {
  id: string;
  source: string;
  sn: string;
  // 被替换为使用 operation 字段，接口还是会返回，但不再使用了
  type: string;
  // 新增 operation 字段，用于替代 type 字段，解决共用同一个审批流(之前的type字段)但需要区分不同的操作类型
  operation: string;
  status: ApplicationStatus;
  applicant: string;
  content: string;
  delivery_detail: string;
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  ticket_url: string;
  [key: string]: any;
}

export default defineComponent({
  setup() {
    const accountStore = useAccountStore();
    const { setTitle } = useBreadcrumb();
    const isLoading = ref(false);
    const currentApplyData = ref<Partial<IApplicationDetail>>({});
    const curApplyKey = ref('');
    const isCancelBtnLoading = ref(false);
    const route = useRoute();
    let interval: NodeJS.Timeout;
    const bizId = computed(() => Number(route.query.bizs));

    // 获取单据详情
    const getMyApplyDetail = async (id: string) => {
      isLoading.value = true;
      try {
        const res = await accountStore.getApplyAccountDetail(id, bizId.value);
        currentApplyData.value = res.data;
        curApplyKey.value = res.data.id;
        const subTitle = APPLICATION_TYPE_MAP[res.data.operation];
        setTitle(subTitle ? `申请单详情 - ${subTitle}` : '申请单详情');

        if ([ApplicationStatus.pending, ApplicationStatus.delivering].includes(res.data.status)) {
          // 避免重复创建 interval
          if (!interval) {
            interval = setInterval(() => getMyApplyDetail(route.query.id as string), 5000);
          }
        } else {
          clearTimer();
        }
      } finally {
        isLoading.value = false;
      }
    };

    const clearTimer = () => {
      clearInterval(interval);
      interval = null as unknown as NodeJS.Timeout;
    };

    onUnmounted(clearTimer);

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

    const render = () => {
      if (!currentApplyData.value?.operation) return null;
      if (['create_load_balancer'].includes(currentApplyData.value.operation)) {
        return <Clb applicationDetail={currentApplyData.value} loading={isLoading.value} />;
      }
      return (
        <div class={'apply-detail-container page-container'}>
          {applyContentRender(currentApplyData, curApplyKey, {
            cancelLoading: isCancelBtnLoading.value,
            onCancel: handleCancel,
          })}
        </div>
      );
    };

    return render;
  },
});
