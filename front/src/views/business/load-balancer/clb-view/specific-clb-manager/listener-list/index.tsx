import { defineComponent, ref, watch } from 'vue';
import './index.scss';
import { useTable } from '@/hooks/useTable/useTable';
import { Button, Form, Input, Radio, Select, Switcher, Tag } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import CommonSideslider from '@/components/common-sideslider';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { BkRadioGroup } from 'bkui-vue/lib/radio';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
// import custom hooks
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useI18n } from 'vue-i18n';

const { FormItem } = Form;

export default defineComponent({
  setup() {
    // use hooks
    const { t } = useI18n();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();

    const { columns, settings } = useColumns('listener');
    const { CommonTable, getListData } = useTable({
      searchOptions: {
        searchData: [
          {
            name: '监听器名称',
            id: 'listenerName',
          },
          {
            name: '协议',
            id: 'protocol',
          },
          {
            name: '端口',
            id: 'port',
          },
          {
            name: '均衡方式',
            id: 'balanceMode',
          },
          {
            name: '域名数量',
            id: 'domainCount',
          },
          {
            name: 'URL数量',
            id: 'urlCount',
          },
          {
            name: '同步状态',
            id: 'syncStatus',
          },
          {
            name: '操作',
            id: 'actions',
          },
        ],
      },
      tableOptions: {
        columns: [
          ...columns,
          {
            label: t('操作'),
            field: 'actions',
            render: () => (
              <div class='operate-groups'>
                <Button text theme='primary'>
                  {t('编辑')}
                </Button>
                <Button text theme='primary'>
                  {t('删除')}
                </Button>
              </div>
            ),
          },
        ],
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        type: `load_balancers/${loadBalancerStore.currentSelectedTreeNode.id}/listeners`,
      },
    });
    const isSliderShow = ref(false);

    watch(
      () => loadBalancerStore.currentSelectedTreeNode,
      (val) => {
        const { id, type } = val;
        if (type !== 'lb') return;
        // 只有当 type='lb' 时, 才去请求对应 lb 下的 listener 列表
        getListData([], `load_balancers/${id}/listeners`);
      },
    );

    return () => (
      <div>
        <CommonTable>
          {{
            operation: () => (
              <div class={'flex-row align-item-center'}>
                <Button theme={'primary'} onClick={() => (isSliderShow.value = true)}>
                  <Plus class={'f20'} />
                  {t('新增监听器')}
                </Button>
                <Button>{t('批量删除')}</Button>
              </div>
            ),
          }}
        </CommonTable>
        <CommonSideslider
          v-model:isShow={isSliderShow.value}
          title={'新增监听器'}
          width={640}
          onHandleSubmit={() => {}}>
          <Form formType='vertical'>
            <FormItem label='监听器名称' required>
              <Input placeholder='请输入' />
            </FormItem>
            <FormItem label='监听协议' required>
              <BkButtonGroup>
                <Button>TCP</Button>
                <Button>UDP</Button>
                <Button>HTTP</Button>
                <Button>HTTPS</Button>
              </BkButtonGroup>
            </FormItem>
            <FormItem label='监听端口' required>
              <Input placeholder='请输入' />
            </FormItem>

            <div class={'flex-row justify-content-between'}>
              <FormItem label='SNI' required>
                <Switcher />
              </FormItem>
              <FormItem label='SSL解析方式' required>
                <BkRadioGroup>
                  <Radio label='单向认证'></Radio>
                  <Tag theme='info'>推荐</Tag>
                  <Radio label='双向认证' class={'ml24'}></Radio>
                </BkRadioGroup>
              </FormItem>
            </div>
            <FormItem label='服务器证书' required>
              <Select></Select>
            </FormItem>
            <FormItem label='CA证书' required>
              <Select></Select>
            </FormItem>
            <FormItem label='默认域名' required>
              <Input placeholder='请输入' />
            </FormItem>
            <FormItem label='URL路径' required>
              <Input placeholder='请输入' />
            </FormItem>

            <FormItem label='均衡方式' required>
              <Select></Select>
            </FormItem>
            <FormItem label='监听器名称' required>
              <Input placeholder='请输入' />
            </FormItem>
            <div class={'flex-row'}>
              <FormItem label='会话保持' required>
                <Switcher />
              </FormItem>
              <FormItem label='保持时间' class={'ml24'} required>
                <Input placeholder='请输入' type='number' suffix='秒' />
              </FormItem>
            </div>
            <FormItem label='目标组' required>
              <Select></Select>
            </FormItem>
          </Form>
        </CommonSideslider>
      </div>
    );
  },
});
