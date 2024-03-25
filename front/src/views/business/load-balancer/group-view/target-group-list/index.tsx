import { defineComponent, onMounted, reactive, ref, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
// import components
import { Input, Popover, VirtualRender } from 'bkui-vue';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
import { useAccountStore } from '@/store';
// import static resources
import allIcon from '@/assets/image/all-vendors.png';
import './index.scss';

export default defineComponent({
  name: 'TargetGroupList',
  emits: ['changeActiveType'],
  setup(_, { emit }) {
    // use hooks
    const router = useRouter();
    const route = useRoute();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const accountStore = useAccountStore();
    // 搜索相关
    const searchValue = ref('');

    const allTargetGroupsItem = reactive({ isDropdownListShow: false });
    // handler - 切换目标组
    const handleTypeChange = (type: 'all' | 'specific', targetGroupId: string) => {
      if (targetGroupId === route.query.tgId) return;
      emit('changeActiveType', type);
      loadBalancerStore.setTargetGroupId(targetGroupId);
      // 将 target-group-id 放入路由中
      router.push({
        path: route.path,
        query: { ...route.query, tgId: targetGroupId || undefined },
      });
    };

    const renderDropdownActionList = (item: any) => {
      return (
        <Popover
          trigger='click'
          theme='light'
          renderType='shown'
          placement='bottom-start'
          arrow={false}
          extCls='more-action-dropdown-menu'
          onAfterHidden={({ isShow }) => (item.isDropdownListShow = isShow)}>
          {{
            default: () => (
              <div
                class={`more-action${item.isDropdownListShow ? ' click' : ''}`}
                onClick={() => (item.isDropdownListShow = !item.isDropdownListShow)}>
                <i class='hcm-icon bkhcm-icon-more-fill'></i>
              </div>
            ),
            content: () => (
              <div class='dropdown-action-list'>
                <div class='dropdown-action-item' onClick={() => {}}>
                  编辑
                </div>
                <div class='dropdown-action-item' onClick={() => {}}>
                  删除
                </div>
              </div>
            ),
          }}
        </Popover>
      );
    };

    onMounted(() => {
      loadBalancerStore.getTargetGroupList();
    });

    watch(
      () => route.query,
      (val) => {
        const { bizs } = val;
        if (!bizs) return;
        // 如果url中有bizs, 则存入store中
        accountStore.updateBizsId(Number(bizs));
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class='target-group-list'>
        <div class='search-wrap'>
          <Input v-model={searchValue.value} type='search' clearable placeholder='搜索目标组' />
        </div>
        <div class='group-list-wrap'>
          <div
            class={`all-groups-wrap${!route.query.tgId ? ' selected' : ''}`}
            onClick={() => handleTypeChange('all', '')}>
            <div class='left-wrap'>
              <img src={allIcon} alt='' class='prefix-icon' />
              <span>全部目标组</span>
            </div>
            <div class='right-wrap'>
              <div class='count'>{6654}</div>
              {renderDropdownActionList(allTargetGroupsItem)}
            </div>
          </div>
          <VirtualRender list={loadBalancerStore.allTargetGroupList} height='calc(100% - 36px)' lineHeight={36}>
            {{
              default: ({ data }: any) => {
                return data.map((item: any) => {
                  return (
                    <div
                      key={item.id}
                      class={`group-item-wrap${route.query.tgId === item.id ? ' selected' : ''}`}
                      // todo: 改 item.id 为目标组的 id 即可
                      onClick={() => handleTypeChange('specific', item.id)}>
                      <div class='left-wrap'>
                        <img src={allIcon} alt='' class='prefix-icon' />
                        <span>{item.name}</span>
                      </div>
                      <div class='right-wrap'>
                        <div class='count'>{item.count}</div>
                        {renderDropdownActionList(item)}
                      </div>
                    </div>
                  );
                });
              },
            }}
          </VirtualRender>
        </div>
      </div>
    );
  },
});
