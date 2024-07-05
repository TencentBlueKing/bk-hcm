import { defineComponent } from 'vue';
import { Ediatable, HeadColumn } from '@blueking/ediatable';
import { useI18n } from 'vue-i18n';
export default defineComponent({
  props: {
    edit: Boolean,
  },
  setup(props, { slots }) {
    const { t } = useI18n();
    return () => (
      <Ediatable>
        {{
          default: () => (
            <>
              <HeadColumn required minWidth={120} width={450}>
                {t('调整方式')}
              </HeadColumn>
              <HeadColumn required minWidth={120} width={450}>
                {t('业务')}
              </HeadColumn>
              <HeadColumn required minWidth={120} width={450}>
                {t('二级账号')}
              </HeadColumn>
              <HeadColumn required minWidth={120} width={450}>
                {t('资源类型')}
              </HeadColumn>
              <HeadColumn required minWidth={120} width={450}>
                {t('金额')}
              </HeadColumn>
              <HeadColumn minWidth={120} width={450}>
                {t('备注')}
              </HeadColumn>
              {!props.edit && (
                <HeadColumn minWidth={120} width={450}>
                  {t('操作')}
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
