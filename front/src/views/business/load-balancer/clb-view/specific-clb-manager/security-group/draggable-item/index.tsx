import { Button, InfoBox } from 'bkui-vue';
import { PropType, defineComponent } from 'vue';
import './index.scss';

export default defineComponent({
  props: {
    securityItem: Object as PropType<{
      cloud_id: string;
      id: string;
      name: string;
    }>,
    idx: Number,
    securitySearchVal: String,
    handleUnbind: Function,
    selectedSecuirtyGroupsSet: Set,
  },
  setup(props) {
    // 高亮命中关键词
    const getHighLightNameText = (name: string, rootCls: string) => {
      return (
        <div
          class={rootCls}
          v-html={name?.replace(
            new RegExp(props.securitySearchVal, 'g'),
            `<span class='search-result-highlight'>${props.securitySearchVal}</span>`,
          )}></div>
      );
    };

    return () => (
      <div
        key={props.securityItem.cloud_id}
        class={
          props.selectedSecuirtyGroupsSet.has(props.securityItem.id)
            ? 'config-security-item-new'
            : 'config-security-item'
        }>
        <i class={'hcm-icon bkhcm-icon-grag-fill mr8 draggable-card-header-draggable-btn'}></i>

        <div class={'config-security-item-idx'}>{props.idx + 1}</div>
        <span class={'config-security-item-name'}>
          {props.securitySearchVal ? getHighLightNameText(props.securityItem.name, '') : props.securityItem.name}
        </span>
        <span class={'config-security-item-id'}>({props.securityItem.cloud_id})</span>
        <div class={'config-security-item-edit-block'}>
          <Button
            text
            theme='primary'
            class={'mr27'}
            onClick={() => {
              const url = `/#/business/security?cloud_id=${props.securityItem.cloud_id}`;
              window.open(url, '_blank');
            }}>
            去编辑
            <span class='icon hcm-icon bkhcm-icon-jump-fill ml5'></span>
          </Button>
          <Button
            class={'mr24'}
            text
            theme='danger'
            onClick={() => {
              InfoBox({
                infoType: 'warning',
                title: '是否确定解绑当前安全组',
                onConfirm() {
                  props.handleUnbind(props.securityItem.id);
                },
              });
            }}>
            <svg
              viewBox='0 0 1024 1024'
              width={11.45}
              height={11.45}
              class={'mr4'}
              fill='#EA3636'
              version='1.1'
              xmlns='http://www.w3.org/2000/svg'>
              <path
                fill-rule='evenodd'
                d='M286.2275905828571 466.8617596342857L346.4517251657142 527.0656921599999 195.92332068571426 677.5959793371429 346.4297720685714 828.1262665142856 496.9801303771428 677.5959793371429 557.1823111314285 737.7989134628571 376.5528159085714 918.4526189714285C368.56767780571425 926.4409673142857 357.73563904 930.9290422857142 346.44074861714284 930.9290422857142 335.1458581942857 930.9290422857142 324.3138194285714 926.4409673142857 316.3286813257143 918.4526189714285L105.57614218971428 707.6974460342857C88.95576212479999 691.0712107885714 88.95576212479999 664.1207471542857 105.57614218971428 647.4945119085714L286.2275905828571 466.8617596342857ZM271.2404106971428 210.97583908571426L813.0869855085714 752.8720991085713 752.8848040228571 813.0750332342856 211.0162761142857 271.17877321142856 271.2404106971428 210.97583908571426ZM392.31718473142854 571.8582944914285L452.17451739428566 631.7174696228572 362.39992393142853 721.47449344 302.5654023314285 661.6361472 392.31718473142854 571.8582944914285ZM707.7107156114284 105.57629805714285L918.463253942857 316.3314731885714C935.0843026285713 332.9578225371429 935.0843026285713 359.9090556342857 918.463253942857 376.53540498285713L737.8138024228571 557.1891126857142 677.6106232685714 496.9642247314285 828.1390299428571 346.4548937142857 677.6106232685714 195.90265270857142 527.0812203885713 346.4349359542857 466.85708580571423 286.23100342857146 647.5085348571429 105.57729594514285C664.1345616457143 88.95571016411428 691.0846888228571 88.95571016411428 707.7107156114284 105.57629805714285ZM661.6181525942857 302.5654023314285L721.47449344 362.4037493028571 631.7008925257143 452.18160128 571.8653791085713 392.32242614857137 661.6181525942857 302.5654023314285Z'
              />
            </svg>
            {props.selectedSecuirtyGroupsSet.has(props.securityItem.id) ? '移除' : '解绑'}
          </Button>
        </div>
      </div>
    );
  },
});
