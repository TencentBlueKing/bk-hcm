import { defineComponent } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import { Button } from 'bkui-vue';

export default defineComponent({
  setup(props, ctx) {
    return () => (
      <div class={'create-second-account-wrapper'}>
        <DetailHeader class={'header'}>
          <span class={'header-title'}>创建二级账号</span>
        </DetailHeader>
        <CommonCard title={() => '基础信息'} class={'info-card'}>
          123
        </CommonCard>
        <CommonCard title={() => '账号信息'} class={'info-card'}>
          123
        </CommonCard>
        <Button theme='primary' class={'mr8 ml24'}>提交</Button>
        <Button>取消</Button>
      </div>
    );
  },
});