<script lang="ts" setup>
import { ref } from 'vue';
import DetailHeader from '../../common/header/detail-header';
import IpInfo from '../components/ip/ip-info.vue';
import AssignEip from '../dialog/assign-eip/assign-eip';

import {
  InfoBox,
} from 'bkui-vue';
import {
  useRoute,
  useRouter
} from 'vue-router';
import useDetail from '../../hooks/use-detail';
import {
  useResourceStore
} from '@/store/resource';
import {
  computed,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';

const route = useRoute();
const router = useRouter();
const resourceStore = useResourceStore();
const {
  t,
} = useI18n();

const isShowAssignEip = ref(false);
const showDelete = ref(false);
const isDeleteing = ref(false);

const {
  loading,
  detail,
  getDetail,
} = useDetail(
  'eips',
  route.query.id as string,
);

const handleShowAssignEip = () => {
  isShowAssignEip.value = true;
}

const handleShowDeleteDialog = () => {
  showDelete.value = true;
}

const handleCloseDeleteEip = () => {
  showDelete.value = false;
}

const handleDeleteEip = () => {
  const postData: any = {
    eip_id: route.query.id,
  }
  if (['gcp', 'azure'].includes(detail.value.vendor)) {
    postData.network_interface_id = detail.value.instance_id
  }
  isDeleteing.value = true;
  resourceStore
    .disassociateEip(postData)
    .then(() => {
      getDetail()
        .then(() => {
          handleCloseDeleteEip()
        })
    })
    .finally(() =>  {
      isDeleteing.value = false;
    })
}

const handleShowDelete = () => {
  InfoBox({
    title: '请确认是否删除',
    subTitle: `将删除【${detail.value.id}】`,
    theme: 'danger',
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    onConfirm() {
      return resourceStore
        .deleteBatch(
          'eips',
          {
            ids: [detail.value.id],
          },
        ).then(() => {
          router.back();
        });
    },
  });
};

const disableOperation = computed(() => {
  return !location.href.includes('business') && detail.value.bk_biz_id !== -1
})
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-header>
      弹性IP：ID（{{ detail.id }}）
      <template #right>
        <bk-button
          v-if="!detail.instance_id"
          class="w100 ml10"
          theme="primary"
          :disabled="disableOperation"
          @click="handleShowAssignEip"
        >
          {{ t('绑定') }}
        </bk-button>
        <bk-button
          v-else
          class="w100 ml10"
          theme="primary"
          :disabled="disableOperation || detail.instance_type === 'OTHER'"
          @click="handleShowDeleteDialog"
        >
          {{ t('解绑') }}
        </bk-button>
        <bk-button
          class="w100 ml10"
          theme="primary"
          :disabled="!!detail.cvm_id || disableOperation || detail.instance_type === 'OTHER'"
          @click="handleShowDelete"
        >
          {{ t('删除') }}
        </bk-button>
      </template>
    </detail-header>

    <ip-info :detail="detail"/>

    <assign-eip
      v-if="detail.id"
      v-model:is-show="isShowAssignEip"
      :detail="detail"
      @success-assign="getDetail"
    />

    <bk-dialog
      title="解绑EIP"
      theme="danger"
      :is-show="showDelete"
      :quick-close="false"
      @closed="handleCloseDeleteEip"
    >
      <div>确定解绑EIP【{{ detail.id }}】吗</div>
      <template #footer>
        <section class="bk-dialog-footer">
          <bk-button theme="danger" :loading="isDeleteing" @click="handleDeleteEip">确定</bk-button>
          <bk-button class="bk-dialog-cancel" :disabled="isDeleteing" @click="handleCloseDeleteEip">取消</bk-button>
        </section>
      </template>
    </bk-dialog>
  </bk-loading>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
</style>
