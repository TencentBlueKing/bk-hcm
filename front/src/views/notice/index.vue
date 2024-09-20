<script setup lang="ts">
import NoticeComponent from '@blueking/notice-component';
import '@blueking/notice-component/dist/style.css';
import { useCommonStore } from '@/store';

const commonStore = useCommonStore();
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

// 是否含有跑马灯类型公告， 如果有跑马灯， navigation的高度可能需要减去40px， 避免页面出现滚动条
// 公告列表change事件回调， isShow代表是否含有跑马灯类型公告
const showAlertChange = function (isShow: boolean) {
  commonStore.setIsNoticeAlert(isShow);
};
</script>

<template>
  <NoticeComponent
    :api-url="`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/notice/current_announcements`"
    @show-alert-change="showAlertChange"
  />
</template>

<style scoped></style>
