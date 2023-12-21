import { defineComponent, ref } from 'vue';
import { Input, Button, VirtualRender, Dropdown } from 'bkui-vue';
import { Plus, AngleDown } from 'bkui-vue/lib/icon';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { useTable } from '@/hooks/useTable/useTable';
import allIcon from '@/assets/image/all-vendors.png';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import CommonSideslider from '../components/common-sideslider';
import TargetGroupSidesliderContent from './target-group-sideslider-content';
import './index.scss';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  name: 'TargetGroupView',
  setup() {
    const { columns, settings } = useColumns('target-group');
    const searchData: ISearchItem[] = [
      {
        id: 'target_group_name',
        name: '目标组名称',
      },
      {
        id: 'clb_id',
        name: 'CLB ID',
      },
      {
        id: 'listener_id',
        name: '监听器ID',
      },
      {
        id: 'vip_address',
        name: 'VIP地址',
      },
      {
        id: 'vip_domain',
        name: 'VIP域名',
      },
      {
        id: 'port',
        name: '端口',
      },
      {
        id: 'protocol',
        name: '协议',
      },
      {
        id: 'rs_ip',
        name: 'RS的IP',
      },
    ];
    const searchUrl = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vpcs/list`;
    const { CommonTable } = useTable({
      columns,
      settings: settings.value,
      searchUrl,
      searchData,
    });
    const searchValue = ref('');

    const renderList = Array(100)
      .fill({})
      .map((_, index: number) => {
        return {
          id: index + 1,
          name: `clb-group-${index + 1}`,
          count: Math.floor(Math.random() * 10) + 1,
        };
      });

    // 新建目标组 sideslider
    const isTargetGroupSideslider = ref(false);
    const handleSubmit = () => {};

    return () => (
      <div class='group-view-page'>
        <div class='left-container'>
          <div class='search-wrap'>
            <Input
              v-model={searchValue.value}
              type='search'
              clearable
              placeholder='搜索目标组'
            />
          </div>
          <div class='group-list-wrap'>
            <div class='all-groups-wrap'>
              <div class='left-wrap'>
                <img src={allIcon} alt='' class='prefix-icon' />
                <span>全部目标组</span>
              </div>
              <div class='right-wrap'>
                <div class='count'>{6654}</div>
              </div>
            </div>
            <VirtualRender
              list={renderList}
              height='calc(100% - 36px)'
              lineHeight={36}>
              {{
                default: ({ data }: any) => {
                  return data.map((item: any) => {
                    return (
                      <div key={item.id} class='group-item-wrap'>
                        <div class='left-wrap'>
                          <img src={allIcon} alt='' class='prefix-icon' />
                          <span>{item.name}</span>
                        </div>
                        <div class='right-wrap'>
                          <div class='count'>{item.count}</div>
                        </div>
                      </div>
                    );
                  });
                },
              }}
            </VirtualRender>
          </div>
        </div>
        <div class='main-container'>
          <div class='common-card-wrap'>
            <CommonTable>
              {{
                operation: () => (
                  <>
                    <Button
                      theme='primary'
                      onClick={() => (isTargetGroupSideslider.value = true)}>
                      <Plus class='f20' />
                      新建
                    </Button>
                    <Dropdown trigger='click' placement='bottom-start'>
                      {{
                        default: () => (
                          <Button>
                            批量操作 <AngleDown class='f20' />
                          </Button>
                        ),
                        content: () => (
                          <DropdownMenu>
                            <DropdownItem>批量删除目标组</DropdownItem>
                            <DropdownItem>批量移除 RS</DropdownItem>
                            <DropdownItem>批量添加 RS</DropdownItem>
                          </DropdownMenu>
                        ),
                      }}
                    </Dropdown>
                  </>
                ),
              }}
            </CommonTable>
          </div>
        </div>
        <CommonSideslider
          title='新建目标组'
          width={960}
          v-model:isShow={isTargetGroupSideslider.value}
          onHandleSubmit={handleSubmit}>
          <TargetGroupSidesliderContent />
        </CommonSideslider>
      </div>
    );
  },
});
