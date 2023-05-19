import { VendorEnum } from "@/common/constant";
import { computed } from "vue";
import { useRouter, useRoute } from "vue-router";

interface IDetail {
  vendor: VendorEnum;
  [key: string]: any;
}

interface IMeta {
  id: string;
  type: string;
  name: string;
}

export const useRouteLinkBtn = (
  data: IDetail,
  meta: IMeta
) => {
  const router = useRouter();
  const route = useRoute();
  const { id, name, type } = meta;
  const { vendor } = data;
  const computedId = computed(() => Array.isArray(data[id]) ? data[id][0] : data[id]);
  const computedName = computed(() => {
    let txt = Array.isArray(data[name]) ? data[name][0] : data[name];
    if(vendor === VendorEnum.AZURE) txt = txt.split('/').reverse()[0];
    return txt;
  });


  const handleClick = () => {
    const routeInfo = {
      query: { id: computedId.value, type }
    };
    if (route.path.includes('business')) {
      Object.assign(
        routeInfo,
        {
          name: `${type}BusinessDetail`,
        },
      );
    } else {
      Object.assign(
        routeInfo,
        {
          name: 'resourceDetail',
          params: {
            type,
          },
        },
      );
    }
    router.push(routeInfo);
  }

  return (
    <bk-button text theme="primary" onClick={handleClick}>
      { computedName.value }
    </bk-button>
  )
}