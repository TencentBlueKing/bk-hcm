import { defineComponent, onMounted, onUnmounted, provide, ref } from 'vue';
import Home from '@/views/home';
import NoticeComponent from '@blueking/notice-component';
import { useUserStore } from '@/store';

const { BK_HCM_AJAX_URL_PREFIX, ENABLE_NOTICE } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    // const router = useRouter();
    // 设置 rem
    const calcRem = () => {
      const doc = window.document;
      const docEl = doc.documentElement;
      const designWidth = 1580; // 默认设计图宽度
      const maxRate = 2560 / designWidth;
      const minRate = 1280 / designWidth;
      const clientWidth = docEl.getBoundingClientRect().width || window.innerWidth;
      const flexibleRem = Math.max(Math.min(clientWidth / designWidth, maxRate), minRate) * 100;
      docEl.style.fontSize = `${flexibleRem}px`;
      // 项目中没有做 rem 适配, 所以这里直接设置 rem 为 14px, 解决 MagicBox 适配 rem 失效问题
      docEl.style.fontSize = '14px';
    };
    const userStore = useUserStore();
    userStore.userInfo();

    // 是否含有跑马灯类型公告， 如果有跑马灯， navigation的高度可能需要减去40px， 避免页面出现滚动条
    const showAlert = ref(false);
    // 公告列表change事件回调， isShow代表是否含有跑马灯类型公告
    const showAlertChange = function (isShow: boolean) {
      showAlert.value = isShow;
    };
    provide('isNoticeAlert', showAlert);

    onMounted(() => {
      calcRem();
      window.addEventListener('resize', calcRem, false);
    });
    onUnmounted(() => {
      window.removeEventListener('resize', calcRem, false);
    });
    return () => (
      <div class='full-page flex-column'>
        {ENABLE_NOTICE && (
          <NoticeComponent
            apiUrl={`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/notice/current_announcements`}
            onShowAlertChange={showAlertChange}
          />
        )}
        <Home class='flex-1'></Home>
      </div>
    );
  },
});
