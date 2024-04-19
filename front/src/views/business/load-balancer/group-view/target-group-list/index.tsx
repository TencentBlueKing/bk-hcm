import { defineComponent, onMounted, ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
// import components
import { Input, Message, VirtualRender } from 'bkui-vue';
import Confirm from '@/components/confirm';
// import stores
import { useLoadBalancerStore, useAccountStore, useBusinessStore } from '@/store';
// import hooks
import useMoreActionDropdown from '@/hooks/useMoreActionDropdown';
// import utils
import bus from '@/common/bus';
// import static resources
import allIcon from '@/assets/image/all-lb.svg';
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
    const businessStore = useBusinessStore();

    // 搜索相关
    const searchValue = ref('');

    const allTargetGroupsItem = { type: 'all', isDropdownListShow: false }; // 全部目标组item
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

    // 删除目标组
    const handleDeleteTargetGroup = (node: any) => {
      const { id, name } = node;
      Confirm('请确定删除目标组', `将删除目标组【${name}】`, () => {
        businessStore.deleteTargetGroups({ bk_biz_id: accountStore.bizs, ids: [id] }).then(() => {
          Message({ message: '删除成功', theme: 'success' });
          // 重新拉取目标组list
          loadBalancerStore.getTargetGroupList();
          // 跳转至全部目标组下
          handleTypeChange('all', '');
        });
      });
    };

    // more-action
    const typeMenuMap = {
      all: [{ label: '新增目标组', handler: () => bus.$emit('addTargetGroup') }],
      specific: [
        { label: '编辑', handler: () => {} }, // todo: 等产品确认编辑位置
        { label: '删除', handler: handleDeleteTargetGroup },
      ],
    };
    const { showDropdownList, currentPopBoundaryNodeKey } = useMoreActionDropdown(typeMenuMap);

    onMounted(() => {
      loadBalancerStore.getTargetGroupList();
    });

    return () => (
      <div class='target-group-list'>
        <div class='search-wrap'>
          <Input v-model={searchValue.value} type='search' clearable placeholder='搜索目标组' />
        </div>
        <div class='group-list-wrap'>
          <div
            class={`all-groups-wrap${!route.query.tgId ? ' selected' : ''}`}
            onClick={() => handleTypeChange('all', '')}>
            <div class='base-info'>
              <img src={allIcon} alt='' class='prefix-icon' />
              <span class='text'>全部目标组</span>
            </div>
            <div class='ext-info'>
              <div class='count'>{6654}</div>
              <div class='more-action' onClick={(e) => showDropdownList(e, allTargetGroupsItem)}>
                <i class='hcm-icon bkhcm-icon-more-fill'></i>
              </div>
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
                      onClick={() => handleTypeChange('specific', item.id)}>
                      <div class='base-info'>
                        <img src={allIcon} alt='' class='prefix-icon' />
                        <span class='text'>{item.name}</span>
                      </div>
                      <div class={`ext-info${currentPopBoundaryNodeKey.value === item.id ? ' show-dropdown' : ''}`}>
                        <div class='count'>{item.count}</div>
                        <div class='more-action' onClick={(e) => showDropdownList(e, { type: 'specific', ...item })}>
                          <i class='hcm-icon bkhcm-icon-more-fill'></i>
                        </div>
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
