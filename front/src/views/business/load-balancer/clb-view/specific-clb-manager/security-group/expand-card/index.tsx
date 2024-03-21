import { Button, Table } from 'bkui-vue';
import { defineComponent, ref } from 'vue';
import './index.scss';
import { AngleDown, AngleUp } from 'bkui-vue/lib/icon';

export default defineComponent({
  props: {
    idx: {
      type: Number,
      required: true,
    },
    name: {
      type: String,
      required: true,
    },
    cloudId: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const isExpand = ref(true);
    return () => (
      <div>
        <div class={'security-rule-table-container'}>
          <div class={'security-rule-table-header'}>
            <div onClick={() => (isExpand.value = !isExpand.value)} class={'header-icon'}>
              {isExpand.value ? <AngleUp width={34} height={28} /> : <AngleDown width={34} height={28} />}
            </div>
            <div class={'config-security-item-idx'}>{props.idx}</div>
            <span class={'config-security-item-name'}>{props.name}</span>
            <span class={'config-security-item-id'}>({props.cloudId})</span>
            <div class={'config-security-item-btn'}>
              <Button theme='primary' text>查看更多</Button>
              <span class='icon hcm-icon bkhcm-icon-jump-fill ml5'></span>
            </div>
          </div>
          {isExpand.value ? (
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
          ) : null}
        </div>
      </div>
    );
  },
});
