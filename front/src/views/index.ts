import business from '@/router/module/business';
import service from '@/router/module/service';
import { billAccountManage, billManage } from '@/views/bill/route-config';

export const businessViews = business;
export const serviceViews = service;
export const billViews = [...billAccountManage, ...billManage];
