import business from '@/router/module/business';
import task from '@/views/task/route-config';

business.forEach((group) => {
  const index = group.children.findIndex((menu) => menu.name === 'businessRecord');
  if (index !== -1) {
    group.children.splice(index + 1, 0, ...task);
  }
});
export const businessViews = business;
