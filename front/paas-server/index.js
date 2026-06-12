const express = require('express');
const path = require('path');
const fs = require('fs');
const https = require('https');
const http = require('http');

const app = express();
const port = process.env.PORT || 5001;
const httpsPort = process.env.HTTPS_PORT || 5443;

const publicPath = '/';

// 设置项目根路径
const projectRoot = path.resolve(__dirname);
const distPath = path.join(projectRoot, '../dist');

// 配置静态资源服务 - 使用/static作为访问入口
const staticPath = path.join(distPath, '');
app.use(
  `${publicPath}static`,
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
  if (req.path.startsWith(`${publicPath}static`) || req.path.startsWith(`${publicPath}assets`)) {
    return res.status(404).send('Not Found');
  }

  try {
    // 返回index.html内容，并将模板变量替换为对应的值
    const replaceVars = {
      '{{ .BK_HCM_AJAX_URL_PREFIX }}': '',
      '{{ .BK_LOGIN_URL }}': '',
      '{{ .BK_COMPONENT_API_URL }}': '',
      '{{ .BK_ITSM_URL }}': '',
      '{{ .VERSION }}': '',
      '{{ .BK_DOMAIN }}': '',
      '{{ .BK_HCM_LANGUAGE_URL_PREFIX }}': '',
      '{{ .BK_CMDB_CREATE_BIZ_DOCS_URL }}': '',
      '{{ .PUBLIC_PATH }}': publicPath,
      '{{ .USER_MANAGE_URL }}': '',
      // 可以在此处添加更多变量替换
    };
    let renderedContent = indexContent;
    for (const [key, value] of Object.entries(replaceVars)) {
      // 全局替换所有出现的变量
      renderedContent = renderedContent.split(key).join(value);
    }
    res.send(renderedContent);
  } catch (error) {
    console.error('Error serving SPA:', error);
    res.status(500).send('Internal Server Error');
  }
});

// 启动服务器
const startServer = () => {
  // 检查是否启用 HTTPS
  const sslKeyPath = process.env.SSL_KEY_PATH || path.join(__dirname, '');
  const sslCertPath = process.env.SSL_CERT_PATH || path.join(__dirname, '');
  const enableHttps = process.env.ENABLE_HTTPS === 'true' || (sslKeyPath && sslCertPath);

  if (enableHttps && sslKeyPath && sslCertPath) {
    // 读取 SSL 证书和私钥
    let sslOptions;
    try {
      sslOptions = {
        key: fs.readFileSync(sslKeyPath, 'utf8'),
        cert: fs.readFileSync(sslCertPath, 'utf8'),
      };
    } catch (error) {
      console.error('Error reading SSL certificate files:', error);
      console.error('Falling back to HTTP only');
      // 如果读取证书失败，回退到 HTTP
      startHttpServer();
      return;
    }

    // 启动 HTTPS 服务器
    https.createServer(sslOptions, app).listen(httpsPort, () => {
      // eslint-disable-next-line
      console.log(`
      🔒 HTTPS Server running at: https://localhost:${httpsPort}
      📂 Static resources served from: ${staticPath}
      🏠 SPA served from: ${indexPath}
      `);
    });

    // 可选：同时启动 HTTP 服务器用于重定向
    if (process.env.ENABLE_HTTP_REDIRECT === 'true') {
      http
        .createServer((req, res) => {
          const host = req.headers.host.replace(/:\d+$/, '');
          res.writeHead(301, { Location: `https://${host}:${httpsPort}${req.url}` });
          res.end();
        })
        .listen(port, () => {
          // eslint-disable-next-line
          console.log(`🔄 HTTP redirect server running at: http://localhost:${port}`);
        });
    }
  } else {
    // 启动 HTTP 服务器
    startHttpServer();
  }
};

const startHttpServer = () => {
  app.listen(port, () => {
    // eslint-disable-next-line
    console.log(`
    🚀 Server running at: http://localhost:${port}
    📂 Static resources served from: ${staticPath}
    🏠 SPA served from: ${indexPath}
    `);
  });
};

startServer();
