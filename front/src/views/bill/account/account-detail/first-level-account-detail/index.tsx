import { defineComponent, ref, watch } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import useBillStore from '@/store/useBillStore';

export default defineComponent({
  props: {
    accountId: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const detail = ref({});
    const billStore = useBillStore();
    const getDetail = async () => {
      const { data } = await billStore.root_account_detail(props.accountId);
      detail.value = data;
    };
    watch(
      () => props.accountId,
      () => {
        getDetail();
      },
      {
        immediate: true,
        deep: true,
      },
    );
    return () => (
      <div class={'account-detail-wrapper'}>
        <p class={'sub-title'}>帐号信息</p>

        <DetailInfo
          detail={detail.value}
          fields={[
            { prop: 'id', name: 'id' },
            { prop: 'name', name: '名字' },
            { prop: 'vendor', name: '云厂商' },
            { prop: 'cloud_id', name: '云ID' },
            { prop: 'email', name: '邮箱' },
            { prop: 'managers', name: '负责人' },
            { prop: 'bak_managers', name: '备份负责人' },
            { prop: 'site', name: '站点' },
            { prop: 'dept_id', name: '组织架构ID' },
            { prop: 'memo', name: '备忘录' },
            { prop: 'creator', name: '创建者' },
            { prop: 'reviser', name: '修改者' },
            { prop: 'created_at', name: '创建时间' },
            { prop: 'updated_at', name: '修改时间' },
          ]}
        />
        <p class={'sub-title'}>API 密钥</p>
        <div class={'detail-info'}>
          <div class='item'>
            <span class='label'>云密钥 ID：</span>
            <span class='value'>{'************'}</span>
          </div>
          <div class='item'>
            <span class='label'>云密钥：</span>
            <span class='value'>{'************'}</span>
          </div>
          <div class='item'>
            <span class='label'>所属账号 ID：</span>
            <span class='value'>{'************'}</span>
          </div>
        </div>
      </div>
    );
  },
});
