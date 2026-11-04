import { PropType, computed, defineComponent, onBeforeUnmount, ref, watch } from 'vue';
import { Select, Tag } from 'bkui-vue';
import { debounce } from 'lodash';
import http from '@/http';
import { timeFormatter } from '@/common/util';
import { useSingleList } from '@/hooks/useSingleList';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { QueryRuleOPEnum } from '@/typings';
import './index.scss';

const { Option } = Select;

type CertValue = string | string[];

interface ICertItem {
  cloud_id: string;
  name?: string;
  cert_status?: string;
  domain?: string[];
  encrypt_algorithm?: string;
  cloud_expired_time?: string;
  [key: string]: any;
}

/**
 * 证书选择器
 * @prop modelValue - 证书 ID
 * @prop type - 证书类型 (SVR | CA)
 * @prop accountId - 云账户 ID
 */
export default defineComponent({
  name: 'CertSelector',
  props: {
    modelValue: [String, Array] as PropType<CertValue>,
    type: String as PropType<'SVR' | 'CA'>,
    accountId: String,
  },
  emits: ['update:modelValue'],
  setup(props, { emit, expose }) {
    const { getBusinessApiPath } = useWhereAmI();
    const url = `/api/v1/cloud/${getBusinessApiPath()}certs/list`;
    const selected = ref<CertValue>(props.modelValue ?? (props.type === 'SVR' ? [] : ''));
    const searchKeyword = ref('');
    const certCache = ref<Record<string, ICertItem>>({});
    const isSVRCert = computed(() => props.type === 'SVR');
    const selectedIds = computed(() => {
      const values = Array.isArray(selected.value) ? selected.value : [selected.value];
      return values.filter((id): id is string => typeof id === 'string' && !!id);
    });

    const mergeCertCache = (items: ICertItem[]) => {
      if (!items.length) return;
      const increment = Object.fromEntries(items.map((item) => [item.cloud_id, item]));
      certCache.value = { ...certCache.value, ...increment };
    };

    const {
      dataList: queryDataList,
      isDataLoad,
      isScrollLoading,
      handleScrollEnd,
      handleRefresh,
    } = useSingleList<ICertItem>({
      url,
      rules: () => {
        const rules = [
          { field: 'cert_type', op: QueryRuleOPEnum.EQ, value: props.type },
          { field: 'account_id', op: QueryRuleOPEnum.EQ, value: props.accountId },
        ];
        if (!searchKeyword.value) return rules;
        return [
          ...rules,
          {
            op: QueryRuleOPEnum.OR,
            rules: [
              { field: 'cloud_id', op: QueryRuleOPEnum.CIS, value: searchKeyword.value },
              { field: 'name', op: QueryRuleOPEnum.CIS, value: searchKeyword.value },
            ],
          },
        ];
      },
      immediate: true,
    });

    watch(queryDataList, mergeCertCache, { immediate: true });

    // 已选证书始终保留在列表中，搜索、刷新和分页不影响回显及调用方的算法校验。
    const dataList = computed(() => {
      const certMap = new Map<string, ICertItem>();
      queryDataList.value.forEach((item) => certMap.set(item.cloud_id, item));
      selectedIds.value.forEach((id) => {
        if (!certMap.has(id)) {
          certMap.set(id, certCache.value[id] ?? { cloud_id: id, name: '' });
        }
      });
      return [...certMap.values()].sort((a, b) => Number(a.cert_status || '1') - Number(b.cert_status || '1'));
    });

    const ensureSelectedCertificates = async () => {
      const missingIds = selectedIds.value.filter((id) => !certCache.value[id]);
      if (!missingIds.length) return;
      try {
        const res = await http.post(url, {
          filter: {
            op: QueryRuleOPEnum.AND,
            rules: [
              { field: 'cert_type', op: QueryRuleOPEnum.EQ, value: props.type },
              { field: 'account_id', op: QueryRuleOPEnum.EQ, value: props.accountId },
              { field: 'cloud_id', op: QueryRuleOPEnum.IN, value: missingIds },
            ],
          },
          page: {
            count: false,
            start: 0,
            limit: missingIds.length,
            sort: 'created_at',
            order: 'DESC',
          },
        });
        mergeCertCache(res.data?.details ?? []);
      } catch (error) {
        console.error(error);
      }
    };

    watch(selectedIds, ensureSelectedCertificates, { immediate: true });
    watch(selected, (val) => emit('update:modelValue', val), { deep: true });
    watch(
      () => props.modelValue,
      (val) => {
        selected.value = val ?? (props.type === 'SVR' ? [] : '');
      },
      { deep: true },
    );
    watch([() => props.accountId, () => props.type], () => {
      certCache.value = {};
      handleRefresh();
    });

    const handleRemoteSearch = debounce(async (keyword: string) => {
      searchKeyword.value = keyword.trim();
      try {
        await handleRefresh();
      } catch (error) {
        console.error(error);
      }
    }, 300);
    onBeforeUnmount(() => handleRemoteSearch.cancel());

    const formatCertLabel = (cert: ICertItem) => `${cert.cloud_id}（${cert.name || '--'}）`;
    const selectedFullLabel = computed(() =>
      selectedIds.value.map((id) => formatCertLabel(certCache.value[id] ?? { cloud_id: id })).join(','),
    );

    // 保持既有 expose 契约，两个调用方继续从 dataList 读取已选证书的加密算法。
    expose({ dataList, handleRefresh });

    return () => (
      <Select
        class='cert-selector'
        v-model={selected.value}
        filterable
        multiple={isSVRCert.value}
        multipleMode={isSVRCert.value ? 'tag' : 'default'}
        placeholder='请选择证书'
        searchPlaceholder='请输入证书ID/名称'
        noDataText='无匹配证书'
        noMatchText='无匹配证书'
        remoteMethod={handleRemoteSearch}
        onScroll-end={handleScrollEnd}
        loading={isDataLoad.value}
        scrollLoading={isScrollLoading.value}
        popoverMinWidth={560}
        inputTooltipsOptions={{ content: selectedFullLabel.value, disabled: !selectedFullLabel.value }}>
        {{
          default: () =>
            dataList.value.map((cert) => {
              const { cloud_id, name, cert_status, domain, encrypt_algorithm, cloud_expired_time } = cert;
              return (
                <Option key={cloud_id} id={cloud_id} name={formatCertLabel(cert)} disabled={cert_status === '3'}>
                  <div class='cert-selector-option'>
                    <div class='cert-selector-option-main'>
                      <span class='cert-selector-option-id'>{cloud_id}</span>
                      <Tag theme='info'>{encrypt_algorithm || '--'}</Tag>
                      {cert_status === '3' && <Tag theme='danger'>已过期</Tag>}
                    </div>
                    <div class='cert-selector-option-secondary'>
                      <span class='cert-selector-option-name' title={name || '--'}>
                        {name || '--'}
                      </span>
                      <span class='cert-selector-option-domain'>
                        主域名 {domain?.[0] || '--'}，备用域名 {domain?.[1] || '--'}
                      </span>
                      <span class='cert-selector-option-expire'>
                        到期 {timeFormatter(cloud_expired_time, 'YYYY-MM-DD')}
                      </span>
                    </div>
                  </div>
                </Option>
              );
            }),
          tagRender: (item: { value: string }) => {
            const cert = certCache.value[item.value] ?? { cloud_id: item.value, name: '' };
            return (
              <span class='cert-selector-selected-label' title={formatCertLabel(cert)}>
                <span>{cert.cloud_id}（</span>
                <span class='cert-selector-selected-name'>{cert.name || '--'}</span>
                <span>）</span>
              </span>
            );
          },
        }}
      </Select>
    );
  },
});
