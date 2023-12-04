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

export default defineComponent({
  setup() {
    const searchVal = ref('');
    const isCreateAccountDialogShow = ref(false);
    const route = useRoute();
    const router = useRouter();
    const {
      handleExpand,
      checkIsExpand,
      getAllVendorsAccountsList,
      accountsMatrix,
      isLoading,
    } = useAllVendorsAccounts();
    const { setAccountId } = useResourceAccount();
    const resourceAccountStore = useResourceAccountStore();

    const handleCancel = () => {
      // isCreateAccountDialogShow.value = false;
      router.push({
        query: {
          ...route.query,
          dialog: undefined,
        },
      });
    };
    const handleSubmit = () => {
    };
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
              onClick={() => {
                router.push({
                  query: {
                    ...route.query,
                    dialog: 'create_account',
                  },
                });
                // isCreateAccountDialogShow.value = true;
              }}>
              <div class={'flex-row align-items-center'}>
                <i class={'icon bk-icon icon-plus-circle mr3'} />
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
        <Loading loading={isLoading.value} style={{ height: 'calc(100% - 85px)' }}>
          {searchVal.value.length ? null : (
            <div
              class={`all-vendors specific-vendor ${
                !resourceAccountStore.resourceAccount?.id
                  ? ' actived-specfic-account'
                  : ''
              }`}
              onClick={() => setAccountId('')}>
              <img src={allVendors} alt='全部账号' class={'vendor-icon'} />
              <div>全部账号</div>
            </div>
          )}
          {computedAllVendorsAccount.value === 0 ? (
            <Exception
              class='exception-wrap-item exception-part'
              type='search-empty'
              scene='part'
              description='搜索为空'
            />
          ) : (
            <VendorAccounts
              accounts={accountsMatrix}
              handleExpand={handleExpand}
              handleSelect={setAccountId}
              checkIsExpand={checkIsExpand}
            />
          )}
        </Loading>
        <CreateAccount
          isShow={isCreateAccountDialogShow.value}
          onCancel={handleCancel}
          onSubmit={handleSubmit}
        />
      </div>
    );
  },
});
