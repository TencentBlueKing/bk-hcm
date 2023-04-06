import { defineComponent, ref, watch  } from 'vue';
import { Button } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useCommonStore } from '@/store';
import { useRoute } from 'vue-router';

import permissions from '@/assets/image/403.png';
import './403.scss';

export default defineComponent({
  setup() {
    const { t } = useI18n();
    const route = useRoute();
    const commonStore = useCommonStore();
    const authUrl = ref('');
    const urlLoading = ref<boolean>(false);
    const urlKey = ref<string>('');

    // 根据urlKey获取权限链接
    const getAuthActionUrl = async () => {
      const { authVerifyData } = commonStore;
      if (authVerifyData) {       // 权限矩阵数据
        const params = authVerifyData.urlParams[urlKey.value];    // 获取权限链接需要的参数
        if (params) {
          urlLoading.value = true;
          const res = await commonStore.authActionUrl(params);    // 列表的权限申请地址
          authUrl.value = res;
          urlLoading.value = false;
        }
      }
    };

    watch(() => route.params.id, (key: any, oldKey: any) => {
      if (key === oldKey) return;
      urlKey.value = key;
      getAuthActionUrl();
    }, { immediate: true });

    // 打开一个新窗口
    const handlePermissionJump = () => {
      window.open(authUrl.value);
    };

    return {
      handlePermissionJump,
      authUrl,
      urlLoading,
      urlKey,
      t,
    };
  },

  render() {
    return (
      <div>
        <div class="forbid-layout">
          <img src={permissions} alt="403" />
          <h2>{this.t('抱歉，您暂无该功能的权限')}</h2>
          <p class="mt10">{this.t('您还没有该功能的权限，可以点击下方的"申请功能权限"获得权限')}</p>
        </div>
        <div class="describe">
          <h2 class="mt20">
            权限申请说明：
          </h2>
          {this.urlKey === 'resource_find' && (   // 资源列表权限说明
            <>
              <p class="mt5 sub-describe">{this.t('该功能由平台资源的管理员维护，属于管理员的权限。')}</p>
              <p class="mt5 sub-describe">{this.t('如果您是业务方用户，无需申请该权限，请在「业务」菜单下直接使用。')}</p>
            </>
          )}
           {this.urlKey === 'biz_access' && ( // 业务列表权限说明
            <>
              <p class="mt5 sub-describe">{this.t('该功能下的资源，由业务自行维护，IaaS资源的创建，一般是由业务运维、SRE等操作')}</p>
              <p class="mt5 sub-describe">{this.t('如果您需要在业务下维护云资源，可以申请业务-IaaS资源下对应业务的权限')}</p>
            </>
           )}

          <h2 class="mt20">
          功能说明：
          </h2>
          {this.urlKey === 'resource_find' && (   // 资源列表功能说明
          <>
            <p class="mt5 sub-describe">{this.t('资源管理功能，屏蔽了各种不同云之间的底层差异，提供了统一的管理模式，方便资源管理员统一全局的管理功能')}</p>
            <p class="mt5 sub-describe">{this.t('具备同时管理多云多账号的云资源，支持多种不同资源的操作')}</p>
            <p class="mt5 sub-describe">{this.t('提供资源的生命周期管理，如资源的创建，回收，销毁等')}</p>
            <p class="mt5 sub-describe">{this.t('支持资源归属不同业务')}</p>
          </>
          )}
          {this.urlKey === 'biz_access' && (   // 业务列表功能说明
          <>
            <p class="mt5 sub-describe">{this.t('具备同时管理多云多账号的云资源，支持多种不同资源的操作')}</p>
            <p class="mt5 sub-describe">{this.t('提供资源的生命周期管理，如资源的创建，回收，销毁等')}</p>
            <p class="mt5 sub-describe">{this.t('主机、硬盘、VPC的新建，需要走申请流程，请到「服务」菜单操作')}</p>
            <p class="mt5 sub-describe">{this.t('其他资源，可以直接在业务下进行操作')}</p>
          </>
          )}
        </div>
        <div class="btn-warp">
        <Button class="mt20" theme="primary"
        loading={this.urlLoading}
        onClick={this.handlePermissionJump}>{this.t('申请权限')}</Button>
        </div>
      </div>
    );
  },
});
