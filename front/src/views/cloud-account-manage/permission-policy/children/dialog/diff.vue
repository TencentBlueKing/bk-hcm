<script setup lang="ts">
import { computed, ref } from 'vue';
import { Message } from 'bkui-vue';
import useClipboard from 'vue-clipboard3';
import hljs from 'highlight.js';

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
const { toClipboard } = useClipboard();

const newContent = ref(
  JSON.stringify(
    {
      version: '2.0',
      statement: [
        { effect: 'allow', action: ['cvm:Describe*', 'cvm:Query*'], resource: '*1' },
        { effect: 'allow', action: ['cbs:Describe*'], resource: '*' },
        { effect: 'allow', action: ['vpc:Describe*', 'vpc:Query*'], resource: '*' },
        {
          effect: 'allow',
          action: ['cos:GetBucket', 'cos:GetObject', 'cos:HeadBucket', 'cos:HeadObject', 'cos:ListAllMyBuckets'],
          resource: '*',
        },
      ],
      statementd: [
        { effect: 'allow', action: ['cvm:Describe*', 'cvm:Query*'], resource: '*1' },
        { effect: 'allow', action: ['cbs:Describe*'], resource: '*' },
        { effect: 'allow', action: ['vpc:Describe*', 'vpc:Query*'], resource: '*' },
        {
          effect: 'allow',
          action: ['cos:GetBucket', 'cos:GetObject', 'cos:HeadBucket', 'cos:HeadObject', 'cos:ListAllMyBuckets'],
          resource: '*',
        },
      ],
      statements: [
        { effect: 'allow', action: ['cvm:Describe*', 'cvm:Query*'], resource: '*1' },
        { effect: 'allow', action: ['cbs:Describe*'], resource: '*' },
        { effect: 'allow', action: ['vpc:Describe*', 'vpc:Query*'], resource: '*' },
        {
          effect: 'allow',
          action: ['cos:GetBucket', 'cos:GetObject', 'cos:HeadBucket', 'cos:HeadObject', 'cos:ListAllMyBuckets'],
          resource: '*',
        },
      ],
    },
    null,
    2,
  ),
); // 策略最新版本
const cloudContent = ref(
  JSON.stringify(
    {
      version: '2.0',
      statement: [
        {
          effect: 'allow',
          action: ['cvm:Describe*', 'cvm:Query*'],
          resource: '*',
        },
        {
          effect: 'allow',
          action: ['cbs:Describe*'],
          resource: '*',
        },
        {
          effect: 'allow',
          action: ['vpc:Describe*', 'vpc:Query*'],
          resource: '*',
        },
      ],
      statementy: [
        { effect: 'allow', action: ['cvm:Describe*', 'cvm:Query*'], resource: '*1' },
        { effect: 'allow', action: ['cbs:Describe*'], resource: '*' },
        { effect: 'allow', action: ['vpc:Describe*', 'vpc:Query*'], resource: '*' },
        {
          effect: 'allow',
          action: ['cos:GetBucket', 'cos:GetObject', 'cos:HeadBucket', 'cos:HeadObject', 'cos:ListAllMyBuckets'],
          resource: '*',
        },
      ],
      asgag: [
        {
          aa: 33,
        },
      ],
    },
    null,
    2,
  ),
); // 云上当前版本

const show = computed(() => props.show);

// const getDiffContent = () => {
//   // 后续通过接口获得
// };

const handleCopy = async (content: string) => {
  try {
    await toClipboard(content);
    Message({ theme: 'success', message: '复制成功' });
  } catch (e) {
    Message({ theme: 'error', message: '复制失败' });
  }
};

const handleClose = () => {
  emit('close');
};
</script>

<template>
  <bk-dialog
    width="1200"
    v-model:is-show="show"
    title="策略内容对比"
    quick-close
    class="policy-diff-dialog"
    @closed="handleClose"
  >
    <template #default>
      <div class="name">二级账号： {{ props.accountId }}</div>
      <div class="diff-info">
        <div>
          <span>云上当前版本（v2）</span>
          <i
            class="hcm-icon bkhcm-icon-copy diff-copy"
            color="#3A84FF"
            title="复制"
            @click="handleCopy(cloudContent)"
          ></i>
        </div>
        <div>
          <span>策略最新版本（v3）</span>
          <i
            class="hcm-icon bkhcm-icon-copy diff-copy"
            color="#3A84FF"
            title="复制"
            @click="handleCopy(newContent)"
          ></i>
          <div class="diff-identify">
            <div class="add-content">新增内容</div>
            <div class="del-content">删除内容</div>
          </div>
        </div>
      </div>
      <bk-code-diff
        class="code-diff"
        :hljs="hljs"
        language="json"
        :new-content="newContent"
        :old-content="cloudContent"
        :diff-context="2000"
        diff-format="side-by-side"
        theme="light"
      />
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.policy-diff-dialog {
  .name {
    color: #313238;
    font-size: 12px;
    margin-bottom: 16px;
  }

  .diff-info {
    display: flex;
    width: 100%;
    background: #d7d9dd;
    color: #000;
    line-height: 32px;

    > div {
      flex: 1;
      padding-left: 24px;
    }

    .diff-copy {
      cursor: pointer;
    }

    .diff-identify {
      color: #4d4f56;
      float: right;

      .add-content,
      .del-content {
        display: inline-block;
        margin-left: 4px;
        margin-right: 10px;

        &::before {
          content: '';
          display: inline-block;
          width: 8px;
          height: 8px;
          border-radius: 50%;
          transform: translateX(-4px);
        }
      }

      .add-content::before {
        background: #2caf5e;
      }
      .del-content::before {
        background: #ff5656;
      }
    }
  }

  .code-diff {
    max-height: 500px;
    overflow-y: auto;
  }

  :deep(.bk-modal-footer) {
    display: none;
  }
}
</style>
