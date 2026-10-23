import http from '@/http';
import { ref } from 'vue';

export const useBusinessFavorite = () => {
  const favoriteSet = ref<Set<number>>(new Set());

  const get = async (id: number) => {
    const { data } = await http.get(`/api/v1/cloud/bizs/${id}/collections/bizs`);
    for (const id of data) {
      favoriteSet.value.add(id);
    }
  };

  const add = async (id: number) => {
    await http.post(`/api/v1/cloud/bizs/${id}/collections/bizs/create`, { bk_biz_id: id });
    favoriteSet.value.add(id);
  };

  const remove = async (id: number) => {
    await http.delete(`/api/v1/cloud/bizs/${id}/collections/bizs`, { data: { bk_biz_id: id } });
    favoriteSet.value.delete(id);
  };

  return {
    favoriteSet,
    get,
    add,
    remove,
  };
};
