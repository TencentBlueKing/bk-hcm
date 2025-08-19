import { computed, defineComponent, onUnmounted, ref, watch } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import { useAccountStore } from '@/store';
import { useRoute } from 'vue-router';
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
  type: string;
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
    const isLoading = ref(false);
    const currentApplyData = ref<Partial<IApplicationDetail>>({});
    const curApplyKey = ref('');
    const isCancelBtnLoading = ref(false);
    const route = useRoute();
    let interval: NodeJS.Timeout;

    // 获取单据详情
    const getMyApplyDetail = async (id: string) => {
      isLoading.value = true;
      try {
        const res = await accountStore.getApplyAccountDetail(id);
        currentApplyData.value = res.data;
        curApplyKey.value = res.data.id;

        if ([ApplicationStatus.pending, ApplicationStatus.delivering].includes(res.data.status)) {
          clearInterval(interval);
          interval = setInterval(() => getMyApplyDetail(route.query.id as string), 5000);
        } else {
          clearInterval(interval);
        }
      } finally {
        isLoading.value = false;
      }
    };

    onUnmounted(() => {
      clearInterval(interval);
    });

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

    const subTitle = computed(() => {
      return APPLICATION_TYPE_MAP[currentApplyData.value?.type];
    });

    const render = () => {
      // 负载均衡详情
      if (!currentApplyData.value?.type) return null;
      if (['create_load_balancer'].includes(currentApplyData.value.type)) {
        return <Clb applicationDetail={currentApplyData.value} loading={isLoading.value} />;
      }
      return (
        <div class={'apply-detail-container'}>
          <DetailHeader>
            {{
              default: () => (
                <>
                  <span class={'title'}>申请单详情</span>
                  <span class={'sub-title'}>
                    &nbsp;-&nbsp;
                    {subTitle.value}
                  </span>
                </>
              ),
            }}
          </DetailHeader>
          <div class={'apply-content-wrapper'}>
            {applyContentRender(currentApplyData, curApplyKey, {
              cancelLoading: isCancelBtnLoading.value,
              onCancel: handleCancel,
            })}
          </div>
        </div>
      );
    };

    return render;
  },
});
