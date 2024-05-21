import { defineComponent, PropType, ref } from 'vue';
import { DEPLOYMENT_ARCHITECTURE_MAP } from '@/constants';
import { ISchemeListItem } from '@/typings/scheme';
import SchemeUserProportionShowDialog from '@/views/scheme/scheme-recommendation/components/scheme-user-proportion-show-dialog';

import './index.scss';
import { Info } from 'bkui-vue/lib/icon';

export default defineComponent({
  name: 'SchemeInfoCard',
  props: {
    schemeDetail: Object as PropType<ISchemeListItem>,
  },
  setup(props) {
    const infos = [
      [
        { id: 'user_distribution', name: '用户据分布地区' },
        { id: 'biz_type', name: '业务类型' },
        { id: 'network', name: '用户网络容忍' },
      ],
      [
        { id: 'user_rate', name: '用户分布占比' },
        { id: 'deployment_architecture', name: '部署架构' },
      ],
    ];
    const isProportionDialogShow = ref(false);

    const getValue = (id: string) => {
      switch (id) {
        case 'user_distribution':
          return props.schemeDetail.user_distribution.map((item) => item.name).join(', ');
        case 'network':
          return `网络延迟 < ${props.schemeDetail.cover_ping}ms`;
        case 'deployment_architecture':
          return props.schemeDetail.deployment_architecture.map((item) => DEPLOYMENT_ARCHITECTURE_MAP[item]).join(', ');
        default:
          return props.schemeDetail[id];
      }
    };

    return () => (
      <div class='scheme-info-card'>
        <div class='info-list'>
          {infos.map((group, index) => {
            return (
              <div class='group' key={index}>
                {group.map((item) => {
                  return (
                    <div class='info-item' key={item.id}>
                      <span class='label'>{item.name}：</span>
                      <span class='info-item-value'>
                        {item.id === 'user_rate' ? (
                          <bk-button
                            text
                            theme='primary'
                            onClick={() => {
                              isProportionDialogShow.value = true;
                            }}>
                            查看详情
                          </bk-button>
                        ) : (
                          getValue(item.id)
                        )}

                        {item.id === 'network' ? (
                          <Info v-bk-tooltips={{ content: '用户到 IDC 的网络质量容忍' }} class={'ml6'} />
                        ) : (
                          ''
                        )}
                      </span>
                    </div>
                  );
                })}
              </div>
            );
          })}
        </div>
        {/* <div class="recreate-btn">
          <bk-button outline theme="primary">重新生成</bk-button>
        </div> */}
        <SchemeUserProportionShowDialog
          v-model:isShow={isProportionDialogShow.value}
          treeData={props.schemeDetail.user_distribution}
        />
      </div>
    );
  },
});
