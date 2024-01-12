import { defineComponent, ref, reactive } from 'vue';
import { Input, VirtualRender, Popover } from 'bkui-vue';
import allIcon from '@/assets/image/all-vendors.png';
import AllGroupsManager from './all-groups-manager';
import SpecificTargetGroupManager from './specific-target-group-manager';
import './index.scss';

export default defineComponent({
  name: 'TargetGroupView',
  setup() {
    const searchValue = ref('');

    // 模拟左侧列表数据
    const renderList = Array(100)
      .fill({})
      .map((_, index: number) => {
        return {
          id: index + 1,
          name: `clb-group-${index + 1}`,
          count: Math.floor(Math.random() * 10) + 1,
          isDropdownListShow: false,
        };
      });

    const componentMap = {
      all: <AllGroupsManager />,
      specific: <SpecificTargetGroupManager />,
    };
    const renderComponent = (type: 'all' | 'specific') => {
      return componentMap[type];
    };
    const activeType = ref('all' as 'all' | 'specific');
    const highlightItemIndex = ref(-1);
    const allTargetGroupsItem = reactive({ isDropdownListShow: false });
    const handleTypeChange = (type: 'all' | 'specific', index: number) => {
      activeType.value = type;
      highlightItemIndex.value = index;
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
                <div class='dropdown-action-item'>编辑</div>
                <div class='dropdown-action-item'>删除</div>
              </div>
            ),
          }}
        </Popover>
      );
    };

    return () => (
      <div class='group-view-page'>
        <div class='left-container'>
          <div class='search-wrap'>
            <Input v-model={searchValue.value} type='search' clearable placeholder='搜索目标组' />
          </div>
          <div class='group-list-wrap'>
            <div
              class={`all-groups-wrap${highlightItemIndex.value === -1 ? ' selected' : ''}`}
              onClick={() => handleTypeChange('all', -1)}>
              <div class='left-wrap'>
                <img src={allIcon} alt='' class='prefix-icon' />
                <span>全部目标组</span>
              </div>
              <div class='right-wrap'>
                <div class='count'>{6654}</div>
                {renderDropdownActionList(allTargetGroupsItem)}
              </div>
            </div>
            <VirtualRender list={renderList} height='calc(100% - 36px)' lineHeight={36}>
              {{
                default: ({ data }: any) => {
                  return data.map((item: any, index: number) => {
                    return (
                      <div
                        key={item.id}
                        class={`group-item-wrap${highlightItemIndex.value === index ? ' selected' : ''}`}
                        onClick={() => handleTypeChange('specific', index)}>
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
        <div class='main-container'>{renderComponent(activeType.value)}</div>
      </div>
    );
  },
});
