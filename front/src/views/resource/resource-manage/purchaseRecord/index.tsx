import CommonCard from '@/components/CommonCard';
import { defineComponent } from 'vue';

export default defineComponent({
  setup() {
    return () => <div>购买记录
      <CommonCard
        layout='grid'
        title={() => '666'}
      >
        <div>132132</div>
        <div>132132</div>
        <div>132132</div>
        <div>132132</div>
      </CommonCard>
    </div>;
  },
});
