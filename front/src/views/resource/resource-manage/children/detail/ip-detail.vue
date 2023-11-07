<script lang="ts" setup>
import { ref, inject,
  computed,
} from 'vue';
import DetailHeader from '../../common/header/detail-header';
import IpInfo from '../components/ip/ip-info.vue';
import AssignEip from '../dialog/assign-eip/assign-eip';

import {
  InfoBox,
} from 'bkui-vue';
import {
  useRoute,
  useRouter,
} from 'vue-router';
import useDetail from '../../hooks/use-detail';
import {
  useResourceStore,
} from '@/store/resource';
import bus from '@/common/bus';
import {
  useI18n,
} from 'vue-i18n';
import { IEip, EipStatus } from '@/typings';
import { CLOUD_VENDOR } from '@/constants/resource';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

const route = useRoute();
const router = useRouter();
const resourceStore = useResourceStore();
const {
  t,
} = useI18n();

const isShowAssignEip = ref(false);
const showDelete = ref(false);
const isDeleteing = ref(false);
const { whereAmI } = useWhereAmI();

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
};

const handleShowDeleteDialog = () => {
  showDelete.value = true;
};

const handleCloseDeleteEip = () => {
  showDelete.value = false;
};

const handleDeleteEip = () => {
  const postData: any = {
    eip_id: route.query.id,
  };
  if (['gcp', 'azure'].includes(detail.value.vendor)) {
    postData.network_interface_id = detail.value.instance_id;
  }
  isDeleteing.value = true;
  resourceStore
    .disassociateEip(postData)
    .then(() => {
      getDetail()
        .then(() => {
          handleCloseDeleteEip();
        });
    })
    .finally(() =>  {
      isDeleteing.value = false;
    });
};

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
  return !location.href.includes('business') && detail.value.bk_biz_id !== -1;
});

const isResourcePage: any = inject('isResourcePage');
const authVerifyData: any = inject('authVerifyData');


const actionName = computed(() => {   // 资源下没有业务ID
  return isResourcePage.value ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';
});

const actionDeleteName = computed(() => {   // 资源下没有业务ID
  return isResourcePage.value ? 'iaas_resource_delete' : 'biz_iaas_resource_delete';
});


// 权限弹窗 bus通知最外层弹出
const showAuthDialog = (authActionName: string) => {
  bus.$emit('auth', authActionName);
};

const canDelete = (data: IEip): boolean => {
  if (data.bk_biz_id !== -1 && whereAmI.value === Senarios.resource) return false;
  let res = false;
  const { status, vendor } = data;

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
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-header>
      弹性IP：ID（{{ detail.id }}）
      <template #right>
        <span v-if="!detail.instance_id" @click="showAuthDialog(actionName)">
          <bk-button
            class="w100 ml10"
            theme="primary"
            :disabled="disableOperation || !authVerifyData?.permissionAction[actionName]"
            @click="handleShowAssignEip"
          >
            {{ t('绑定') }}
          </bk-button>
        </span>
        <span v-else @click="showAuthDialog(actionName)">
          <bk-button
            class="w100 ml10"
            theme="primary"
            :disabled="disableOperation || detail.instance_type === 'OTHER'
              || !authVerifyData?.permissionAction[actionName]"
            @click="handleShowDeleteDialog"
          >
            {{ t('解绑') }}
          </bk-button>
        </span>
        <span @click="showAuthDialog(actionDeleteName)">
          <bk-button
            class="w100 ml10"
            theme="primary"
            :disabled="!canDelete(detail) && (!!detail.cvm_id || disableOperation || detail.instance_type === 'OTHER'
              || !authVerifyData?.permissionAction[actionDeleteName])"
            @click="handleShowDelete"
          >
            {{ t('删除') }}
          </bk-button>
        </span>
      </template>
    </detail-header>

    <ip-info :detail="detail" />

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
