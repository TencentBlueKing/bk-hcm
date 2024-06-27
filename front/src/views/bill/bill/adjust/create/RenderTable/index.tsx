import { defineComponent } from 'vue';
import { Ediatable, HeadColumn } from '@blueking/ediatable';
export default defineComponent({
  props: {
    edit: Boolean,
  },
  setup(props, { slots }) {
    return () => (
      <Ediatable>
        {{
          default: () => (
            <>
              <HeadColumn required minWidth={120} width={450}>
                调整方式
              </HeadColumn>
              <HeadColumn required minWidth={120} width={450}>
                产品
              </HeadColumn>
              <HeadColumn required minWidth={120} width={450}>
                二级账号
              </HeadColumn>
              <HeadColumn required minWidth={120} width={450}>
                资源类型
              </HeadColumn>
              <HeadColumn required minWidth={120} width={450}>
                金额
              </HeadColumn>
              <HeadColumn minWidth={120} width={450}>
                备注
              </HeadColumn>
              {!props.edit && (
                <HeadColumn minWidth={120} width={450}>
                  操作
                </HeadColumn>
              )}
            </>
          ),
          data: slots.default?.(),
        }}
      </Ediatable>
    );
  },
});
