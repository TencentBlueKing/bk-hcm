<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { Message } from 'bkui-vue';
import useClipboard from 'vue-clipboard3';
import hljs from 'highlight.js';
import 'highlight.js/styles/github.css'; // 引入github得样式

interface IProps {
  show: boolean;
  accountId: string;
  id: string;
}

const props = withDefaults(defineProps<IProps>(), {
  show: false,
  accountId: '',
  id: '',
});

const emit = defineEmits(['close']);

const content = ref('');
const originContent = ref('');
const fullscreen = ref(false);
const { toClipboard } = useClipboard();
const show = computed(() => props.show);

const handleClose = () => {
  emit('close');
};

const handleCopy = async (content: string) => {
  try {
    await toClipboard(content);
    Message({ theme: 'success', message: '复制成功' });
  } catch (e) {
    Message({ theme: 'error', message: '复制失败' });
  }
};

// 给代码添加行号
const addLineNumbersForCode = (html: string) => {
  let num = 1;
  html = `<span class="ln-num" data-num="${num}"></span>${html}`;
  html = html.replace(/\r\n|\r|\n/g, function (a) {
    num = num + 1;
    return `${a}<span class="ln-num" data-num="${num}"></span>`;
  });
  html = `<span class="ln-bg"></span>${html}`;
  return html;
};

const getInfo = () => {
  // 测试
  originContent.value = `
   public boolean batchUpdatePlanVersionByIds(List<Long> planIds, String version, String userName, Long modifyTime) {
        int affected = context.update(TABLE)
            .set(TABLE.VERSION, version)
            .set(TABLE.LAST_MODIFY_USER, userName)
            .set(TABLE.LAST_MODIFY_TIME, ULong.valueOf(modifyTime))
            .where(TABLE.ID.in(planIds))
            .execute();
        return affected > 0;
         public boolean batchUpdatePlanVersionByIds(List<Long> planIds, String version, String userName, Long modifyTime) {
        int affected = context.update(TABLE)
            .set(TABLE.VERSION, version)
            .set(TABLE.LAST_MODIFY_USER, userName)
            .set(TABLE.LAST_MODIFY_TIME, ULong.valueOf(modifyTime))
            .where(TABLE.ID.in(planIds))
            .execute();
        return affected > 0;
         public boolean batchUpdatePlanVersionByIds(List<Long> planIds, String version, String userName, Long modifyTime) {
        int affected = context.update(TABLE)
            .set(TABLE.VERSION, version)
            .set(TABLE.LAST_MODIFY_USER, userName)
            .set(TABLE.LAST_MODIFY_TIME, ULong.valueOf(modifyTime))
            .where(TABLE.ID.in(planIds))
            .execute();
        return affected > 0;
         public boolean batchUpdatePlanVersionByIds(List<Long> planIds, String version, String userName, Long modifyTime) {
        int affected = context.update(TABLE)
            .set(TABLE.VERSION, version)
            .set(TABLE.LAST_MODIFY_USER, userName)
            .set(TABLE.LAST_MODIFY_TIME, ULong.valueOf(modifyTime))
            .where(TABLE.ID.in(planIds))
            .execute();
        return affected > 0;
        
 public boolean batchUpdatePlanVersionByIds(List<Long> planIds, String version, String userName, Long modifyTime) {
        int affected = context.update(TABLE)
            .set(TABLE.VERSION, version)
            .set(TABLE.LAST_MODIFY_USER, userName)
            .set(TABLE.LAST_MODIFY_TIME, ULong.valueOf(modifyTime))
            .where(TABLE.ID.in(planIds))
            .execute();
        return affected > 0;
    }`;
  content.value = addLineNumbersForCode(hljs.highlight(originContent.value, { language: 'java' }).value);
};

onMounted(() => {
  getInfo();
});
</script>

<template>
  <bk-sideslider
    v-model:is-show="show"
    class="log-detail-sideslider"
    @closed="handleClose"
    :width="fullscreen ? '100%' : 750"
  >
    <template #header>
      <div class="log-header">
        <div>执行日志</div>
        <div class="icon">
          <i
            class="hcm-icon bkhcm-icon-copy log-copy"
            color="#3A84FF"
            title="复制"
            @click="handleCopy(originContent)"
          ></i>
          <i
            class="hcm-icon bkhcm-icon-fullscreen log-fullscreen"
            color="#3A84FF"
            title="全屏"
            v-if="!fullscreen"
            @click="fullscreen = true"
          ></i>
          <i
            class="hcm-icon bkhcm-icon-zoomout log-zoomout"
            color="#3A84FF"
            title="还原"
            v-if="fullscreen"
            @click="fullscreen = false"
          ></i>
        </div>
      </div>
    </template>
    <div id="log" v-safe-html="content"></div>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.log-detail-sideslider {
  #log {
    white-space: pre;

    :deep(.ln-num) {
      &:before {
        content: attr(data-num);
        display: inline-block;
        width: 30px;
        color: #8c8f99;
        font-size: 12px;
        text-align: right;
        margin-right: 15px;
      }
    }
  }

  .log-header {
    display: flex;
    align-items: center;
    width: 100%;
    justify-content: space-between;
    padding-right: 24px;

    .hcm-icon {
      cursor: pointer;
      font-size: 16px;
    }
  }

  :deep(.bk-modal-body) {
    .bk-modal-content {
      height: 580px;
      overflow-y: auto;
      margin: 0 8px;
    }
    .bk-modal-footer {
      display: none;
    }
  }
}
</style>
