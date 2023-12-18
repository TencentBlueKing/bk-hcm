<template>
     <div class="template-warp">
      <div class="flex-row operate-warp justify-content-between align-items-center mb20">
        <div @click="handleAuth('account_import')">
          <bk-button
            theme="primary"
            :disabled="!authVerifyData.permissionAction.account_import">
            {{t('购买') }}
          </bk-button>
          <bk-button
            style="margin-left: 10px;" 
            @click="() => exampleSetting.dialog.isShow = true">
           分配
          </bk-button>
         <bk-dialog
         :is-show="exampleSetting.dialog.isShow"
         :title="'批量分配/主机分配'"
         :theme="'primary'"
         @closed="() => exampleSetting.dialog.isShow = false"
         @confirm="() => exampleSetting.dialog.isShow = false"
         >
         <p class="selected-host-count-tip">
        已选择 <span class="selected-host-count">{{  }}</span> 台主机，可选择所需分配的目标业务
      </p>
      <p class="mb6">目标业务</p>
      <!-- <bk-form-item property="name" label="业务" required>
    <bk-select
    v-model="formData.name"
    
  >
    <bk-option
    value="1"
          label="本科以下"
    />
  </bk-select>
  </bk-form-item> -->
  
        </bk-dialog>

          <bk-button
            style="margin-left: 10px;" 
            :disabled="!authVerifyData.permissionAction.account_import">
            {{t('批量删除') }}
          </bk-button>
        </div>

        <div class="flex-row input-warp justify-content-between align-items-center">
          <bk-search-select class="bg-white w280" 
          :conditions="[]" 
          v-model="searchValue" 
          :data="searchData"
          filterable
          auto-focus>
          </bk-search-select>
        </div>
      </div>
      
      <bk-loading
        :loading="loading"
      >
        <bk-table
          class="table-layout"
          :data="tableData"
          :is-row-select-enable="isRowSelectEnable"
          @column-sort="handleSortBy"
          remote-pagination
          :columns="columns"

          :pagination="{
            count: pageCount,
            limit: memoPageSize,
            current: memoPageIndex
          }"
          show-overflow-tooltip
          @page-value-change="handlePageValueChange"
          @page-limit-change="handlePageLimitChange"
          row-hover="auto"
          row-key="id"
        >
          <bk-table-column
          type="selection"
          sort
          :width="100"
          />

          <bk-table-column
            :label="t('负载均衡域名称')"
            fields="name"
            sort
            isDefaultShow: true,
          >
          <template #default="{ data }">
              {{data?.price || '--'}}{{data?.price_unit}}
            </template>
          </bk-table-column>
  
          <bk-table-column
            label="云厂商"
            fields="vendor"
            sort
            isDefaultShow: true,
          >
            <template #default="props">
              {{AccountType[props?.row.type]}}
            </template>
          </bk-table-column>

          <bk-table-column
            label="地域"
            fields="region"
            sort
            isDefaultShow: true,
          >
            
          </bk-table-column>

          <bk-table-column
            label="可用区域"
            fields="re"
            sort
            isDefaultShow: true,
          >
          <template #default="{ data }">
              {{data?.price || '--'}}{{data?.price_unit}}
            </template>
          </bk-table-column>

          <bk-table-column
            label="负载均衡域名"
            fields="domain"
            sort
            isDefaultShow: true,
          >
            
          </bk-table-column>
  
          <bk-table-column
            :label="t('负载均衡VIP')"
            fields="VIP"
            sort
            isDefaultShow: true,
          >
            <template #default="props">
              {{CloudType[props?.row?.vendor]}}
            </template>
          </bk-table-column>
  
          <bk-table-column
            label="网络类型"
            fields="network"
            sort
            isDefaultShow: true,
          >
            <!-- <template #default="props">
              {{SITE_TYPE_MAP[props?.row.site]}}
            </template> -->
          </bk-table-column>
         
          <bk-table-column
            :label="t('监听数量')"
            prop="count"
            sort
            isDefaultShow: true,
          >
            <template #default="props">
              {{props?.row.managers?.join(',')}}
            </template>
          </bk-table-column>

          <bk-table-column
            label="状态"
            fields="state"
            sort
            isDefaultShow: true,
          >
          </bk-table-column>

          <bk-table-column
            label="分配状态"
            fields="fpstate"
            sort
            isDefaultShow: true,
          >
          <template #default="{ data }">
              {{data?.price || '--'}}{{data?.price_unit}}
            </template>
          </bk-table-column>

          <bk-table-column
            label="所属网络"
            fields="bk_biz_id2"
            sort
            isDefaultShow: true,
          >
          </bk-table-column>
         
          <bk-table-column
            :label="t('IP版本')"
            prop="IP"
            isDefaultShow: true,
          >
            <template #default="{ data }">
              {{data?.price || '--'}}{{data?.price_unit}}
            </template>
          </bk-table-column>
          <bk-table-column
            :label="t('操作')"
            isDefaultShow: true,
          >
            <template #default="props">
              <div class="operate-button">
                <div @click="handleAuth('account_edit')">
                  <bk-button
                    text theme="primary" @click="handleJump('accountDetail', props?.row.id)"
                    :disabled="!authVerifyData.permissionAction.account_edit">
                    {{t('编辑')}}
                  </bk-button>
                </div>
                <bk-button class="ml15" text theme="primary">
                  {{t('删除')}}
                </bk-button>
              </div>
            </template>
          </bk-table-column>
        </bk-table>
    </bk-loading>
</div>
</template>

<script setup lang="ts">
import bkUi from 'bkui-vue'
  import { reactive, watch, toRefs, defineComponent, onMounted, ref, computed } from 'vue';
  import { useRouter } from 'vue-router';
  import { useI18n } from 'vue-i18n';
  import { useAccountStore } from '@/store';
  import { CloudType, AccountType } from '@/typings';
  import { ACCOUNT_TYPES, SITE_TYPES, SITE_TYPE_MAP, VENDORS } from '@/common/constant';
  import { useVerify } from '@/hooks';
  import { useMemoPagination, DEFAULT_PAGE_INDEX, DEFAULT_PAGE_SIZE } from '@/hooks/useMemoPagination';
  const { t } = useI18n();
      const router = useRouter();
      const accountStore = useAccountStore();
      const {
        setMemoPageSize,
        setMemoPageIndex,
        memoPageIndex,
        memoPageSize,
        memoPageStart,
      } = useMemoPagination();
      const state = reactive({
        isAccurate: false,    // 是否精确
        searchValue: [],
        searchData: [
          {
            name: '名称',
            field: 'name',
          },
          {
            name: '负载均衡域名',
            field: 'type',
            children: ACCOUNT_TYPES,
          },
          {
            name: '负载均衡VIP',
            field: 'vendor',
            children: VENDORS,
          },
          {
            name: '网络类型',
            field: 'site',
            children: SITE_TYPES,
          },
          {
            name: '监听数量',
            field: 'managers',
          },
          {
            name: 'IP版本',
            field: 'creator',
          },
        ],
        showDeleteBox: false,
        tableData: [],
        formData: [],
        columns:[],
        loading: true,
        dataId: null,
        CloudType,
        AccountType,
        filter: { op: 'and', rules: [] },
        type: '',
        btnLoading: false,
      });
  
      const pageCount = ref(0);
      const formRef = ref('');

      const fromData = ref({
        name: '',
      })
      const exampleSetting = ref({
        dialog: {
            isShow: false
        }
      })
      

  
      // 权限hook
      const {
        handleAuth,
        authVerifyData,
      } = useVerify();
  
      // 请求获取列表的总条数
      const getListCount = async () => {
        const params = {
          filter: state.filter,
          page: {
            count: true,
          },
        };
        const res = await accountStore.getAccountList(params);
        pageCount.value = res?.data.count || 0;
      };
  
      const getAccountList = async () => {
        state.loading = true;
        try {
          const params = {
            filter: state.filter,
            page: {
              count: false,
              limit: memoPageSize.value,
              start: memoPageStart.value,
              sort: 'created_at',
              order: 'DESC',
            },
          };
          const res = await accountStore.getAccountList(params);
          state.tableData = res.data.details;
        } catch (error) {
          console.log(error);
        } finally {
          state.loading = false;
        }
      };
  
      // 搜索数据
      watch(
        () => state.searchValue,
        (val, oldVal) => {
          console.log('val', val);
          state.filter.rules = val.reduce((p, v) => {
            if (v.type === 'condition') {
              state.filter.op = v.id || 'and';
            } else {
              console.log('v.values[0].id', v.values[0].id);
              if (v.id === 'managers') {
                p.push({
                  field: v.id,
                  op: 'json_contains',
                  value: v.values[0].id,
                });
              } else {
                p.push({
                  field: v.id,
                  op: state.isAccurate ? 'eq' : 'cs',
                  value: v.values[0].id,
                });
              }
            }
            return p;
          }, []);
          pageCount.value = 0;
          if (oldVal !== undefined) {
            setMemoPageIndex(DEFAULT_PAGE_INDEX);
            setMemoPageSize(DEFAULT_PAGE_SIZE);
          }
          /* 获取账号列表接口 */
          getListCount(); // 数量
          getAccountList(); // 列表
        },
        {
          deep: true,
          immediate: true,
        },
      );

      const init = () => {
        setMemoPageIndex(DEFAULT_PAGE_INDEX);
        setMemoPageSize(DEFAULT_PAGE_SIZE);
        state.isAccurate = false;
        state.searchValue = [];
        getAccountList();
      };
      // 弹窗确认
      const handleDialogConfirm = async (diaType: string) => {
        state.btnLoading = true;
        try {
          if (diaType === 'del') {    // 删除
            await accountStore.accountDelete(state.dataId);
          } 
          state.btnLoading = false;
          // 重新请求列表
          init();
        } catch (error) {
          console.log(error);
        } finally {
          state.btnLoading = false;
          state.showDeleteBox = false;
        }
      };
  
      // 跳转页面
      const handleJump = (routerName: string, id?: string, isDetail?: boolean) => {
        const routerConfig = {
          query: {},
          name: routerName,
        };
        if (id) {
          routerConfig.query = {
            id,
            isDetail,
          };
        }
        router.push(routerConfig);
      };
      const handlePageLimitChange = (limit: number) => {
        setMemoPageSize(limit);
        setMemoPageIndex(DEFAULT_PAGE_INDEX);
        getAccountList();
      };
  
      const handlePageValueChange = (value: number) => {
        setMemoPageIndex(value);
        getAccountList();
      };
  const handleSortBy = () => {
    console.log('handleSortBy');
    
  }

 
</script>

<style scoped lang="scss">
.operate-button{
    display: flex;
  }
  .btn-warp{
    margin-top: 30px;
    justify-content: end;
  }
    .sync-dialog-warp{
      height: 150px;
      .t-icon{
        height: 42px;
        width: 110px;
      }
      .logo-icon{
          height: 42px;
          width: 42px;
      }
      .arrow-icon{
        position: relative;
        flex: 1;
        overflow: hidden;
        height: 13px;
        line-height: 13px;
        .content{
          width: 130px;
          position: absolute;
          left: 200px;
          animation: 3s move infinite linear;
        }
      }
    }
  @-webkit-keyframes move {
    from {
          left: 0%;
      }
  
      to {
          left: 100%;
      }
  }
  
  @keyframes move {
      from {
          left: 0%;
      }
  
      to {
          left: 100%;
      }
  }
</style>