const express = require('express');
const path = require('path');
const fs = require('fs');

const app = express();
const port = process.env.PORT || 5000;

// 设置项目根路径
const projectRoot = path.resolve(__dirname);
const distPath = path.join(projectRoot, '../dist');

// 配置静态资源服务 - 使用/static作为访问入口
const staticPath = path.join(distPath, '');
app.use(
  '/static',
  express.static(staticPath, {
    cacheControl: false,
  }),
);

// 主入口文件路径
const indexPath = path.join(distPath, 'index.html');
if (!fs.existsSync(indexPath)) {
  throw new Error('index.html not found in dist directory!');
}

// 读取index.html内容
let indexContent = '';
try {
  indexContent = fs.readFileSync(indexPath, 'utf8');
} catch (error) {
  console.error('Error reading index.html:', error);
  process.exit(1);
}

// 处理所有非静态资源的GET请求
app.get('*', (req, res) => {
  // 检查请求路径是否是静态资源路径
  if (req.path.startsWith('/static') || req.path.startsWith('/assets')) {
    // eslint-disable-next-line
    console.log('Static resource requested:', req.path);
    return res.status(404).send('Not Found');
  }

  try {
    // 简单路由分析
    const routeInfo = {
      path: req.path,
      method: req.method,
      timestamp: new Date().toISOString(),
      userAgent: req.get('User-Agent'),
      clientIP: req.ip,
    };

    // eslint-disable-next-line
    console.log(`Serving SPA for: ${routeInfo.path}`);

    // 返回index.html内容
    res.send(indexContent);
  } catch (error) {
    console.error('Error serving SPA:', error);
    res.status(500).send('Internal Server Error');
  }
});

// 启动服务器
app.listen(port, () => {
  // eslint-disable-next-line
  console.log(`
  🚀 Server running at: http://localhost:${port}
  📂 Static resources served from: ${staticPath}
  🏠 SPA served from: ${indexPath}
  `);
});
