hcm-saas-vue3

# 本地开发
```
通过fork主仓库的方式进行开发
1. fork 主仓库的v1.0分支
2. 本地开发之后提pr至主仓库的v1.0
```

#### 安装依赖包
```
npm install
```

#### 配置host
```
127.0.0.1 dev.hcm.example.com（看具体环境使用的域名）推荐使用switchHosts
```

#### 本地开发环境变量
```
env.dev.js
index-dev.html 这两个文件问项目开发人员要
```

#### 如何访问
```
启动后，本地访问地址示例 `http://dev.hcm.example.com:{5000}` 5000为默认端口，可修改
````