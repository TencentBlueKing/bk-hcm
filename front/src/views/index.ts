import business from '@/router/module/business';
import service from '@/router/module/service';
import rollingServer from '@/views/rolling-server/route-config';
import greenChannel from '@/views/green-channel/route-config';
import stats from '@/views/stats/route-config';

export const businessViews = business;
export const serviceViews = service;

export const platformManagementViews = [...rollingServer, ...greenChannel, ...stats];
