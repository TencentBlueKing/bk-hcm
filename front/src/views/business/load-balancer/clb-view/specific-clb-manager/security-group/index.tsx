import { defineComponent, ref } from 'vue';
import './index.scss';
import { Button, Table, Tag } from 'bkui-vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import SearchInput from '@/views/scheme/components/search-input';
import CommonSideslider from '@/components/common-sideslider';
import CommonDialog from '@/components/common-dialog';

export default defineComponent({
  setup() {
    const rsCheckRes = ref('a');
    const securityRuleType = ref('in');
    const isSideSliderShow = ref(false);
    const hanldeSubmit = () => {};
    return () => (
      <div>
        <div class={'rs-check-selector-container'}>
          <div
            class={`${rsCheckRes.value === 'a' ? 'rs-check-selector-active' : 'rs-check-selector'}`}
            onClick={() => {
              rsCheckRes.value = 'a';
            }}>
            <Tag theme='warning'>2 次检测</Tag>
            <span>依次经过负载均衡和RS的安全组 2 次检测</span>
          </div>
          <div
            class={`${rsCheckRes.value === 'b' ? 'rs-check-selector-active' : 'rs-check-selector'}`}
            onClick={() => {
              rsCheckRes.value = 'b';
            }}>
            <Tag theme='warning'>1 次检测</Tag>
            <span>只经过负载均衡的安全组 1 次检测，忽略后端RS的安全组检测</span>
          </div>
        </div>
        <div class={'security-rule-container'}>
          <p>
            <span class={'security-rule-container-title'}>绑定安全组</span>
            <span class={'security-rule-container-desc'}>
              当负载均衡不绑定安全组时，其监听端口默认对所有 IP 放通。此处绑定的安全组是直接绑定到负载均衡上面。
            </span>
          </p>
          <div class={'security-rule-container-operations'}>
            <Button theme='primary' class={'mr12'} onClick={() => (isSideSliderShow.value = true)}>
              配置
            </Button>
            <Button>全部收起</Button>
            <div class={'security-rule-container-searcher'}>
              <BkButtonGroup class={'mr12'}>
                <Button
                  selected={securityRuleType.value === 'in'}
                  onClick={() => {
                    securityRuleType.value = 'in';
                  }}>
                  出站规则
                </Button>
                <Button
                  selected={securityRuleType.value === 'out'}
                  onClick={() => {
                    securityRuleType.value = 'out';
                  }}>
                  入站规则
                </Button>
              </BkButtonGroup>
              <SearchInput placeholder='请输入' />
            </div>
          </div>
          <div class={'specific-security-rule-tables'}>
            {[1, 2, 3].map(() => (
              <div class={'security-rule-table-container'}>
                <div class={'security-rule-table-header'}>title</div>
                <div class={'security-rule-table-panel'}>
                  <Table
                    stripe
                    data={[
                      {
                        target: 'any',
                        source: 'Any',
                        portProtocol: 'abnc',
                        policy: '允许',
                      },
                      {
                        target: '1abc',
                        source: '1ads',
                        portProtocol: 'abc',
                        policy: '拒绝',
                      },
                      {
                        target: 'abc',
                        source: 'abc',
                        portProtocol: 'Tabv3',
                        policy: '允许',
                      },
                    ]}
                    columns={[
                      {
                        label: '目标',
                        field: 'target',
                      },
                      {
                        label: '来源',
                        field: 'source',
                      },
                      {
                        label: '端口协议',
                        field: 'portProtocol',
                      },
                      {
                        label: '策略',
                        field: 'policy',
                      },
                    ]}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
        <CommonSideslider
          v-model:isShow={isSideSliderShow.value}
          title='配置安全组'
          width={'640'}
          onHandleSubmit={hanldeSubmit}>
          <div class={'config-security-rule-contianer'}>
            <div class={'config-security-rule-operation'}>
              <Button>新增绑定</Button>
              <SearchInput />
            </div>
            <div>
              {[1, 2, 3].map(() => (
                <div class={'config-security-item'}>testdddd</div>
              ))}
            </div>
          </div>
        </CommonSideslider>
        <CommonDialog />
      </div>
    );
  },
});
