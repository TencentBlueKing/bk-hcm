import { defineComponent, ref, watch, PropType } from 'vue';
import { Button } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useCommonStore } from '@/store';
import { useRoute } from 'vue-router';

import permissions from '@/assets/image/403.png';
import './403.scss';

export default defineComponent({
  props: {
    urlKeyId: String as PropType<string>,
  },
  setup(props) {
    const { t } = useI18n();
    const route = useRoute();
    const commonStore = useCommonStore();
    const authUrl = ref('');
    const urlLoading = ref<boolean>(false);
    const urlKey = ref<string>('');

    // 根据urlKey获取权限链接
    const getAuthActionUrl = async () => {
      const { authVerifyData } = commonStore;
      console.log(authVerifyData.urlParams, urlKey.value);
      if (authVerifyData) {
        // 权限矩阵数据
        const params = authVerifyData.urlParams[urlKey.value]; // 获取权限链接需要的参数
        if (params) {
          urlLoading.value = true;
          const res = await commonStore.authActionUrl(params); // 列表的权限申请地址
          authUrl.value = res;
          urlLoading.value = false;
        }
      }
    };

    watch(
      () => route.path,
      (path) => {
        if (['/scheme/recommendation'].includes(path)) {
          urlKey.value = 'cloud_selection_recommend';
          getAuthActionUrl();
        }
        if (['/scheme/deployment/list'].includes(path)) {
          urlKey.value = 'cloud_selection_find';
          getAuthActionUrl();
        }
      },
      {
        immediate: true,
      },
    );

    watch(
      () => route.params.id,
      (key: any, oldKey: any) => {
        if (key === oldKey) return;
        urlKey.value = key;
        getAuthActionUrl();
      },
      { immediate: true },
    );

    watch(
      () => props.urlKeyId,
      (key: any, oldKey: any) => {
        console.log(key, oldKey);
        if (key === oldKey) return;
        urlKey.value = key;
        getAuthActionUrl();
      },
      { immediate: true },
    );

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
        <div class='forbid-layout'>
          <img src={permissions} alt='403' />
          <h2>{this.t('抱歉，您暂无该功能的权限')}</h2>
          <p class='mt10'>{this.t('您还没有该功能的权限，可以点击下方的"申请权限"获得权限')}</p>
        </div>
        <div class='describe'>
          <h2 class='mt20'>权限申请说明：</h2>
          {this.urlKey === 'cloud_selection_recommend' && (
            <>
              <p class='mt5 sub-describe'>{'当前无“资源选型-选型推荐”的权限'}</p>
            </>
          )}
          {this.urlKey === 'cloud_selection_find' && (
            <>
              <p class='mt5 sub-describe'>{'当前无“部署方案-方案查看”的权限'}</p>
            </>
          )}
          {this.urlKey === 'resource_find' && ( // 资源列表权限说明
            <>
              <p class='mt5 sub-describe'>{this.t('该功能由平台资源的管理员维护，属于管理员的权限。')}</p>
              <p class='mt5 sub-describe'>{this.t('如果您是业务方用户，无需申请该权限，请在"业务"菜单下直接使用。')}</p>
            </>
          )}
          {this.urlKey === 'biz_access' && ( // 业务列表权限说明
            <>
              <p class='mt5 sub-describe'>
                {this.t('该功能下的资源，由业务自行维护，IaaS资源的创建，一般是由业务运维、SRE等操作')}
              </p>
              <p class='mt5 sub-describe'>
                {this.t('如果您需要在业务下维护云资源，可以申请业务-IaaS资源下对应业务的权限')}
              </p>
            </>
          )}
          {this.urlKey === 'account_find' && ( // 账号列表权限说明
            <>
              <p class='mt5 sub-describe'>
                {this.t('该功能用于管理云账号，如业务运维对云账号进行管理，可以对录入海垒的账号进行查看')}
              </p>
              <p class='mt5 sub-describe'>
                {this.t(
                  '如果您是业务下云账号的资源使用者，无需申请该权限。对账号的数据查看，可以申请对应账号的"账号查看"权限，无需申请账号"录入权限"',
                )}
              </p>
              <p class='mt5 sub-describe'>{this.t('如果您只需要录入账号，请在业务-服务申请-云账号录入')}</p>
            </>
          )}
          {this.urlKey === 'resource_audit_find' && ( // 审计列表权限说明
            <>
              <p class='mt5 sub-describe'>
                {this.t(
                  '如果您是业务运维、SRE等角色，业务下管理了多个云账号，可申请"业务审计查看"权限，查看业务下多个账号的审计信息。',
                )}
              </p>
              <p class='mt5 sub-describe'>
                {this.t(
                  '如果您的账号属于某个业务，您负责其中一个账号，可申请"资源审计查看"权限，单独查看该账号的审计信息',
                )}
              </p>
            </>
          )}
          {this.urlKey === 'recycle_bin_find' && ( // 回收站列表权限说明
            <>
              <p class='mt5 sub-describe'>{this.t('该功能由平台资源的管理员维护，属于管理员的权限')}</p>
              <p class='mt5 sub-describe'>
                {this.t('如果您是业务方用户，无需申请该权限，请在业务菜单主机、硬盘回收记录中查看回收信息')}
              </p>
            </>
          )}

          <h2 class='mt20'>功能说明：</h2>
          {this.urlKey === 'cloud_selection_recommend' && (
            <>
              <p class='mt5 sub-describe'>
                {
                  '资源选型，是根据业务需求，推荐出业务的部署地点，云资源方案的功能。当前页面访问受限，可到权限中心申请权限'
                }
              </p>
            </>
          )}
          {this.urlKey === 'cloud_selection_find' && (
            <>
              <p class='mt5 sub-describe'>
                {'部署方案，是系统推荐出的，由用户保存的选型推荐方案。当前页面访问受限，可到权限中心申请权限'}
              </p>
            </>
          )}
          {this.urlKey === 'resource_find' && ( // 资源列表功能说明
            <>
              <p class='mt5 sub-describe'>
                {this.t(
                  '资源管理功能，屏蔽了各种不同云之间的底层差异，提供了统一的管理模式，方便资源管理员统一全局的管理功能',
                )}
              </p>
              <p class='mt5 sub-describe'>{this.t('具备同时管理多云多账号的云资源，支持多种不同资源的操作')}</p>
              <p class='mt5 sub-describe'>{this.t('提供资源的生命周期管理，如资源的创建，回收，销毁等')}</p>
              <p class='mt5 sub-describe'>{this.t('支持资源归属不同业务')}</p>
            </>
          )}
          {this.urlKey === 'biz_access' && ( // 业务列表功能说明
            <>
              <p class='mt5 sub-describe'>{this.t('具备同时管理多云多账号的云资源，支持多种不同资源的操作')}</p>
              <p class='mt5 sub-describe'>{this.t('提供资源的生命周期管理，如资源的创建，回收，销毁等')}</p>
              <p class='mt5 sub-describe'>{this.t('主机、硬盘、VPC的新建，需要走申请流程，请到"服务"菜单操作')}</p>
              <p class='mt5 sub-describe'>{this.t('其他资源，可以直接在业务下进行操作')}</p>
            </>
          )}
          {this.urlKey === 'account_find' && ( // 账号列表功能说明
            <>
              <p class='mt5 sub-describe'>
                {this.t('资源账号：用于从云上同步、更新、操作、购买资源的账号，需要API密钥')}
              </p>
              <p class='mt5 sub-describe'>{this.t('登记账号：云上的普通登录用户，用于被安全审计的账号对象')}</p>
              <p class='mt5 sub-describe'>
                {this.t('安全审计账号：用于对云上资源进行安全审计的账号，需要API密钥，权限比资源账号低')}
              </p>
            </>
          )}
          {this.urlKey === 'resource_audit_find' && ( // 审计列表功能说明
            <>
              <p class='mt5 sub-describe'>
                {this.t(
                  '审计信息包括包括账号信息，IaaS资源想增删改查等。有2种区别：业务操作审计，业务下的审计信息；资源操作审计，以账号为粒度的审计信息。',
                )}
              </p>
            </>
          )}
          {this.urlKey === 'recycle_bin_find' && ( // 回收站列表功能说明
            <>
              <p class='mt5 sub-describe'>
                {this.t('资源回收的管理功能，对业务回收的主机、硬盘资源进行管理，如对销毁、恢复操作')}
              </p>
              <p class='mt5 sub-describe'>{this.t('资源恢复后，将恢复到从原回收的业务。')}</p>
              <p class='mt5 sub-describe'>
                {this.t('资源立即销毁，将从云上直接删除资源，销毁属于不可逆操作，请谨慎操作。')}
              </p>
            </>
          )}
        </div>
        <div class='btn-warp'>
          <Button class='mt20' theme='primary' loading={this.urlLoading} onClick={this.handlePermissionJump}>
            {this.t('申请权限')}
          </Button>
        </div>
      </div>
    );
  },
});
