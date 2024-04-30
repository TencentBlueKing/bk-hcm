import { defineComponent, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { Input, Message, VirtualRender } from 'bkui-vue';
import Confirm from '@/components/confirm';
import { useLoadBalancerStore, useAccountStore, useBusinessStore } from '@/store';
import useMoreActionDropdown from '@/hooks/useMoreActionDropdown';
import useList from '@/hooks/useList';
import { debounce, throttle } from 'lodash';
import bus from '@/common/bus';
import { LBRouteName } from '@/constants';
import { QueryRuleOPEnum } from '@/typings';
import allIcon from '@/assets/image/all-lb.svg';
import './index.scss';

export default defineComponent({
  name: 'TargetGroupList',
  setup() {
    // use hooks
    const router = useRouter();
    const route = useRoute();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const accountStore = useAccountStore();
    const businessStore = useBusinessStore();

    // 搜索相关
    const searchValue = ref('');

    const activeTargetGroupId = ref(''); // 当前选中的目标组id
    const allTargetGroupsItem = { type: 'all', isDropdownListShow: false }; // 全部目标组item

    // 获取目标组列表
    const rules = ref([]);
    const { list, pagination, getList, handleScrollEnd, reset, refresh } = useList(
      'target_groups',
      () => rules.value,
      true,
    );

    // handler - 切换目标组
    const handleTypeChange = (targetGroupId: string) => {
      // 如果两个目标组id相同，则不做切换
      if (targetGroupId === activeTargetGroupId.value) return;
      // 设置当前选中的目标组id
      activeTargetGroupId.value = targetGroupId;
      loadBalancerStore.setTargetGroupId(targetGroupId);
      // 导航
      router.push({
        name: targetGroupId ? LBRouteName.tg : LBRouteName.allTgs,
        query: { ...route.query, type: targetGroupId ? route.query.type : undefined },
        params: { id: targetGroupId || undefined },
      });
    };

    // 删除目标组
    const handleDeleteTargetGroup = (node: any) => {
      const { id, name } = node;
      Confirm('请确定删除目标组', `将删除目标组【${name}】`, () => {
        businessStore.deleteTargetGroups({ bk_biz_id: accountStore.bizs, ids: [id] }).then(() => {
          Message({ message: '删除成功', theme: 'success' });
          // 重新拉取目标组list
          refresh();
          // 跳转至全部目标组下
          handleTypeChange('');
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

    // 滚动触底加载下一页的目标组数据
    const scrollEndHandler = throttle((endIndex: number) => {
      if (endIndex === list.value.length) {
        // 如果 endIndex 等于总数，说明已经到底了，需要拉取更多数据
        handleScrollEnd();
      }
    }, 300);

    watch(
      () => route.params.id,
      (val) => {
        // 高亮状态保持
        if (!val) activeTargetGroupId.value = '';
        else activeTargetGroupId.value = val as string;
      },
      { immediate: true },
    );

    watch(
      searchValue,
      debounce((val) => {
        // 清空搜索结果
        reset();
        // 设置搜索条件
        if (val) {
          rules.value = [{ field: 'name', op: QueryRuleOPEnum.CIS, value: val }];
        } else {
          rules.value = [];
        }
        // 拉取搜索结果
        getList();
      }, 300),
    );

    onMounted(() => {
      bus.$on('refreshTargetGroupList', refresh);
    });

    onUnmounted(() => {
      bus.$off('refreshTargetGroupList');
    });

    return () => (
      <div class='target-group-list'>
        <div class='search-wrap'>
          <Input v-model={searchValue.value} type='search' clearable placeholder='搜索目标组' />
        </div>
        <div class='group-list-wrap'>
          <div
            class={`all-groups-wrap${!activeTargetGroupId.value ? ' selected' : ''}`}
            onClick={() => handleTypeChange('')}>
            <div class='base-info'>
              <img src={allIcon} alt='' class='prefix-icon' />
              <span class='text'>全部目标组</span>
            </div>
            <div class='ext-info'>
              <div class='count'>{pagination.count}</div>
              <div class='more-action' onClick={(e) => showDropdownList(e, allTargetGroupsItem)}>
                <i class='hcm-icon bkhcm-icon-more-fill'></i>
              </div>
            </div>
          </div>
          <VirtualRender
            list={list.value}
            height='calc(100% - 36px)'
            lineHeight={36}
            onContentScroll={([, pagination]) => scrollEndHandler(pagination.endIndex)}>
            {{
              default: ({ data }: any) => {
                return data.map((item: any) => {
                  return (
                    <div
                      key={item.id}
                      class={`group-item-wrap${activeTargetGroupId.value === item.id ? ' selected' : ''}`}
                      onClick={() => handleTypeChange(item.id)}>
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
