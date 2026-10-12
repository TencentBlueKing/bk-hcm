<script lang="ts" setup>
import { ref, inject, computed, watchEffect } from 'vue';
import IpInfo from '../components/ip/ip-info.vue';
import AssignEip from '../dialog/assign-eip/assign-eip';

import { InfoBox, Message } from 'bkui-vue';
import { useRouter } from 'vue-router';
import useDetail from '../../hooks/use-detail';
import { useResourceStore } from '@/store/resource';
import { useI18n } from 'vue-i18n';
import { IEip, EipStatus } from '@/typings';
import { CLOUD_VENDOR } from '@/constants/resource';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { useResourceDetailAuth, useResourceDetailRoute } from '../../hooks/use-resource-detail-route';

const router = useRouter();
const resourceStore = useResourceStore();
const { t } = useI18n();

const isShowAssignEip = ref(false);
const showDelete = ref(false);
const isDeleteing = ref(false);
const { whereAmI } = useWhereAmI();
const { setTitle } = useBreadcrumb();

const { detailId } = useResourceDetailRoute();
const { loading, detail, getDetail } = useDetail('eips', detailId.value);

watchEffect(() => {
  if (detail.value?.id) {
    setTitle(`弹性IP：ID（${detail.value.id}）`);
  }
});

const handleShowAssignEip = () => {
  isShowAssignEip.value = true;
};

const handleShowDeleteDialog = () => {
  showDelete.value = true;
};

const handleCloseDeleteEip = () => {
  showDelete.value = false;
};

const handleDeleteEip = () => {
  const postData: any = {
    eip_id: detailId.value,
  };
  if (['gcp', 'azure'].includes(detail.value.vendor)) {
    postData.network_interface_id = detail.value.instance_id;
  }
  isDeleteing.value = true;
  resourceStore
    .disassociateEip(postData)
    .then(() => {
      getDetail().then(() => {
        handleCloseDeleteEip();
      });
    })
    .finally(() => {
      isDeleteing.value = false;
    });
};

const handleShowDelete = () => {
  InfoBox({
    title: '请确认是否删除',
    subTitle: `将删除【${detail.value.cloud_id}${detail.value.name ? ` - ${detail.value.name}` : ''}】`,
    theme: 'danger',
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    extCls: 'delete-resource-infobox',
    onConfirm() {
      return resourceStore
        .deleteBatch('eips', {
          ids: [detail.value.id],
        })
        .then(() => {
          Message({
            theme: 'success',
            message: '删除成功',
          });
          router.back();
        });
    },
  });
};

const disableOperation = computed(() => {
  return !location.href.includes('business') && detail.value.bk_biz_id !== -1;
});

const isResourcePage: any = inject('isResourcePage');
const { authPending, updateSign, deleteSign } = useResourceDetailAuth(computed(() => detail.value.account_id));

const hasNoRelateResource = ({ vendor, status }: IEip): boolean => {
  let res = false;
  switch (vendor) {
    case CLOUD_VENDOR.tcloud:
      if (status === EipStatus.UNBIND) res = true;
      break;
    case CLOUD_VENDOR.huawei:
      if ([EipStatus.BIND_ERROR, EipStatus.DOWN, EipStatus.ERROR].includes(status)) res = true;
      break;
    case CLOUD_VENDOR.aws:
      if (status === EipStatus.UNBIND) res = true;
      break;
    case CLOUD_VENDOR.gcp:
      if (status === EipStatus.RESERVED) res = true;
      break;
    case CLOUD_VENDOR.azure:
      if (status === EipStatus.UNBIND) res = true;
      break;
  }
  return res;
};
const canDelete = (data: IEip): boolean => {
  if (data.bk_biz_id !== -1 && whereAmI.value === Senarios.resource) return false;
  return hasNoRelateResource(data);
};

const bkToolTipsOptions = computed(() => {
  // 资源下，是否分配业务
  if (isResourcePage.value && detail.value?.bk_biz_id !== -1)
    return {
      content: '该弹性IP已分配到业务，仅可在业务下操作',
      disabled: detail.value.bk_biz_id === -1,
    };
  // 业务/资源下，弹性IP是否已经被资源绑定
  if (detail.value?.cvm_id || !hasNoRelateResource(detail.value) || detail.value.instance_type === 'OTHER')
    return {
      content: '该弹性IP已绑定资源，不可以删除',
      disabled: !(detail.value?.cvm_id || !hasNoRelateResource(detail.value) || detail.value.instance_type === 'OTHER'),
    };

  return {
    disabled: true,
  };
});
</script>

<template>
  <Teleport defer to="#breadcrumbExtra">
    <hcm-auth :sign="updateSign" :ignore="authPending" tag="span" v-slot="{ noPerm }">
      <bk-button
        v-if="!detail.instance_id"
        theme="primary"
        @click="handleShowAssignEip"
        :disabled="disableOperation || noPerm || authPending"
        v-bk-tooltips="{
          content: '该弹性IP已分配到业务，仅可在业务下操作',
          disabled: !disableOperation,
        }"
      >
        {{ t('绑定') }}
      </bk-button>
      <bk-button
        v-else
        theme="primary"
        :disabled="disableOperation || detail.instance_type === 'OTHER' || noPerm || authPending"
        @click="handleShowDeleteDialog"
      >
        {{ t('解绑') }}
      </bk-button>
    </hcm-auth>
    <hcm-auth :sign="deleteSign" :ignore="authPending" tag="span" v-slot="{ noPerm }">
      <bk-button
        theme="primary"
        :disabled="
          !canDelete(detail) ||
          !!detail.cvm_id ||
          disableOperation ||
          detail.instance_type === 'OTHER' ||
          noPerm ||
          authPending
        "
        @click="handleShowDelete"
        v-bk-tooltips="noPerm ? { disabled: true } : bkToolTipsOptions"
      >
        {{ t('删除') }}
      </bk-button>
    </hcm-auth>
  </Teleport>

  <div class="detail-content-wrap">
    <bk-loading :loading="loading">
      <ip-info :detail="detail" />
      <assign-eip v-if="detail.id" v-model:is-show="isShowAssignEip" :detail="detail" @success-assign="getDetail" />
    </bk-loading>
  </div>

  <bk-dialog title="解绑EIP" theme="danger" :is-show="showDelete" :quick-close="false" @closed="handleCloseDeleteEip">
    <div>确定解绑EIP【{{ detail.id }}】吗</div>
    <template #footer>
      <section class="bk-dialog-footer">
        <bk-button theme="danger" :loading="isDeleteing" @click="handleDeleteEip">确定</bk-button>
        <bk-button class="bk-dialog-cancel" :disabled="isDeleteing" @click="handleCloseDeleteEip">取消</bk-button>
      </section>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
</style>
