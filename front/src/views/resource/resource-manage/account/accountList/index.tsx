import { defineComponent, ref, watch } from 'vue';
import { Button, Input, Loading } from 'bkui-vue';
import './index.scss';
import allVendors from '@/assets/image/all-vendors.png';
import CreateAccount from '../createAccount';
import VendorAccounts from './components/VendorAccounts';
import { useAllVendorsAccounts } from './useAllVendorsAccountsList';
import { useResourceAccount } from './useResourceAccount';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';

export default defineComponent({
  setup() {
    const searchVal = ref('');
    const isCreateAccountDialogShow = ref(false);
    const {
      handleExpand,
      getAllVendorsAccountsList,
      accountsMatrix,
      isLoading,
    } = useAllVendorsAccounts();
    const {
      setAccountId,
    } = useResourceAccount();
    const resourceAccountStore = useResourceAccountStore();

    const handleCancel = () => {
      isCreateAccountDialogShow.value = false;
    };
    const handleSubmit = () => {
      console.log(666);
    };
    // const rows = ref(new Array(9999).fill({}).map((_, idx) => {
    //   return { id: idx };
    // }));

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

    return () => (
      <div class={'account-list-container'}>
        <Input
          class={'search-input'}
          placeholder='搜索云厂商，云账号'
          type='search'
          clearable
          v-model={searchVal.value}></Input>
        <div class={'account-list-header'}>
          <p class={'header-title'}>账号列表</p>
          <div class={'header-btn'}>
            <Button text theme='primary' onClick={() => isCreateAccountDialogShow.value = true}>
              <i class={'icon bk-icon icon-plus-circle mr3'}/>
              接入
            </Button>
          </div>
        </div>
        <div class={`all-vendors specific-vendor ${!resourceAccountStore.resourceAccount?.id ? ' actived-specfic-account' : ''}`} onClick={() => setAccountId('')}>
          <img src={allVendors} alt='全部账号'class={'vendor-icon'} />
          <div>全部账号</div>
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
         <Loading
          loading={isLoading.value}
        >
          <VendorAccounts
            accounts={accountsMatrix}
            handleExpand={handleExpand}
            handleSelect={setAccountId}
          />
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
