/**
 * @file prod server
 * 静态资源
 * 模块渲染输出
 * 注入全局变量
 * 添加html模板引擎
 */
 const Express = require('express');
 const path = require('path');
 const artTemplate = require('express-art-template');
 const cookieParser = require('cookie-parser');
 const history = require('connect-history-api-fallback');
 // const auth = require('./middleware/auth');
 // const installServices = require('./api');
 
 const app = new Express();
 const PORT = process.env.PORT || 5005;
 app.use(cookieParser());
 
 process.env.NODE_ENV = 'production'
 // 注入全局变量
 const GLOBAL_VAR = {
   SITE_URL: process.env.SITE_URL || '',
   BK_STATIC_URL: process.env.BK_STATIC_URL || '',
   // 当前应用的环境，预发布环境为 stag，正式环境为 prod
   BKPAAS_ENVIRONMENT: process.env.BKPAAS_ENVIRONMENT || '',
   // EngineApp名称，拼接规则：bkapp-{appcode}-{BKPAAS_ENVIRONMENT}
   BKPAAS_ENGINE_APP_NAME: process.env.BKPAAS_ENGINE_APP_NAME || '',
   // 内部版对应ieod，外部版对应tencent，混合云版对应clouds
   BKPAAS_ENGINE_REGION: process.env.BKPAAS_ENGINE_REGION || '',
   // APP CODE
   BKPAAS_APP_ID: process.env.BKPAAS_APP_ID || '',
   // 登录地址
   BK_LOGIN_URL: process.env.BK_LOGIN_URL || ''
 };
 
 const distDir = path.resolve(__dirname, '../dist');
 
 app.use(history({
   index: '/',
   rewrites: [
     {
       from: /\/(\d+\.)*\d+$/,
       to: '/',
     },
     {
       from: /\/\/+.*\..*\//,
       to: '/',
     },
     {
       from: /\/api\//,
       to: function(context) {
         return context.parsedUrl.href
       }
     }
   ],
 }));
 
 // 首页
 app.get('/', (req, res) => {
   const scriptName = (req.headers['x-script-name'] || '').replace(/\//g, '');
   // 使用子路径
   if (scriptName) {
     GLOBAL_VAR.BK_STATIC_URL = `/${scriptName}`;
     GLOBAL_VAR.SITE_URL = `/${scriptName}`;
   } else {
     // 使用系统分配域名
     GLOBAL_VAR.BK_STATIC_URL = '';
     GLOBAL_VAR.SITE_URL = '';
   }
   // 注入全局变量
   res.render(path.join(distDir, 'index.html'), GLOBAL_VAR);
 });
 
 app.use('/static', Express.static(path.join(distDir, '../dist/static')));
 // 配置视图
 app.set('views', path.join(__dirname, '../dist'));
 
 // 配置模板引擎
 // http://aui.github.io/art-template/zh-cn/docs/
 app.engine('html', artTemplate);
 app.set('view engine', 'html');
 
 // 注册api
 // app.use(auth);
 // installServices(app);
 
 // 配置端口
 app.listen(PORT, () => {
   console.log(`App is running in port ${PORT}`);
 });
 