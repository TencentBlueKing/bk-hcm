import { computed, defineComponent, ref, watch } from 'vue';
import { Button, Exception, Input, Loading } from 'bkui-vue';
import './index.scss';
import allVendors from '@/assets/image/all-vendors.png';
import CreateAccount from '../createAccount';
import VendorAccounts from './components/VendorAccounts';
import { useAllVendorsAccounts } from './useAllVendorsAccountsList';
import { useResourceAccount } from './useResourceAccount';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { useRoute, useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import { useVerify } from '@/hooks';
import PermissionDialog from '@/components/permission-dialog';

export default defineComponent({
  setup() {
    const searchVal = ref('');
    const isCreateAccountDialogShow = ref(false);
    const route = useRoute();
    const router = useRouter();
    const { handleExpand, checkIsExpand, getAllVendorsAccountsList, getVendorAccountList, accountsMatrix, isLoading } =
      useAllVendorsAccounts();
    const { setAccountId } = useResourceAccount();
    const resourceAccountStore = useResourceAccountStore();
    const { currentVendor, currentAccountVendor } = storeToRefs(resourceAccountStore);
    const {
      showPermissionDialog,
      handlePermissionConfirm,
      handlePermissionDialog,
      handleAuth,
      permissionParams,
      authVerifyData,
    } = useVerify();

    const handleCancel = () => {
      // isCreateAccountDialogShow.value = false;
      router.push({
        query: {
          ...route.query,
          dialog: undefined,
        },
      });
    };
    const handleSubmit = () => {};
    const computedAllVendorsAccount = computed(() => accountsMatrix.reduce((acc, { count }) => acc + count, 0));

    // 初始化账号列表/搜索账号
    watch(
      () => searchVal.value,
      (val) => {
        getAllVendorsAccountsList(val);
      },
      {
        immediate: true,
      },
    );

    watch(
      () => route.query.dialog,
      (dialog) => {
        if (dialog) isCreateAccountDialogShow.value = true;
        else isCreateAccountDialogShow.value = false;
      },
      {
        deep: true,
        immediate: true,
      },
    );

    return () => (
      <div class={'account-list-container'}>
        <Input
          class={'search-input'}
          placeholder='搜索云账号'
          type='search'
          clearable
          v-model={searchVal.value}></Input>
        <div class={'account-list-header'}>
          <p class={'header-title'}>账号列表</p>
          <div class={'header-btn'}>
            <Button
              text
              theme='primary'
              class={!authVerifyData.value?.permissionAction?.account_import ? 'hcm-no-permision-text-btn' : ''}
              onClick={() => {
                if (!authVerifyData.value?.permissionAction?.account_import) {
                  handleAuth('account_import');
                } else {
                  router.push({
                    query: {
                      ...route.query,
                      dialog: 'create_account',
                    },
                  });
                }
              }}>
              <div class={'flex-row align-items-center'}>
                <i class={'hcm-icon bkhcm-icon-plus-circle mr3'} />
                接入
              </div>
            </Button>
          </div>
        </div>

        {/* 新增一个子账号组件，能够虚拟滚动 */}
        {/* <VirtualRender
          list={rows.value}
          lineHeight={1}
          height={100}
          onContentScroll={(...args) => {
            console.log(args);
          }}
         >
          {{
            default: ({data}: any) => {
              return data.map(({id}: any) => <div>{id}</div>);
            }
          }}
         </VirtualRender> */}
        <Loading loading={isLoading.value} style={{ height: 'calc(100% - 87px)' }}>
          {searchVal.value.length ? null : (
            <div
              class={`all-vendors specific-vendor ${
                !(currentAccountVendor.value || currentVendor.value) ? ' actived-specfic-account' : ''
              }`}
              onClick={() => {
                setAccountId('');
                resourceAccountStore.setCurrentVendor(null);
                resourceAccountStore.setCurrentAccountVendor(null);
              }}>
              <img src={allVendors} alt='全部账号' class={'vendor-icon'} />
              <div>全部账号</div>
            </div>
          )}
          {computedAllVendorsAccount.value === 0 ? (
            <Exception
              class='exception-wrap-item exception-part'
              type='search-empty'
              scene='part'
              description='无搜索结果'
            />
          ) : (
            <VendorAccounts
              accounts={accountsMatrix}
              searchVal={searchVal.value}
              handleExpand={handleExpand}
              handleSelect={setAccountId}
              checkIsExpand={checkIsExpand}
              getVendorAccountList={getVendorAccountList}
            />
          )}
        </Loading>
        <CreateAccount isShow={isCreateAccountDialogShow.value} onCancel={handleCancel} onSubmit={handleSubmit} />
        {/* 申请权限 */}
        <PermissionDialog
          v-model:isShow={showPermissionDialog.value}
          params={permissionParams.value}
          onCancel={handlePermissionDialog}
          onConfirm={handlePermissionConfirm}
        />
      </div>
    );
  },
});
