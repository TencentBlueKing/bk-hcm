import { defineComponent, ref } from 'vue';
import { Button, Input } from 'bkui-vue';
import './index.scss';
import allVendors from '@/assets/image/all-vendors.png';

export default defineComponent({
  setup() {
    const searchVal = ref('');
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
            <Button text theme='primary'>
              <i class={'icon bk-icon icon-plus-circle mr3'} />
              接入
            </Button>
          </div>
        </div>
        <div class={'all-vendors specific-vendor'}>
          <img src={allVendors} alt='全部账号'class={'vendor-icon'} />
          <div>全部账号</div>
        </div>
      </div>
    );
  },
});
