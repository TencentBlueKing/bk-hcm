declare global {
  interface Window {
    PROJECT_CONFIG: {
      [key: string]: any;
    };
    [key: string]: any;
  }
}

// export {} 将其标记为外部模块，模块是至少包含1个导入或导出语句的文件，我们必须这样做才能扩展全局范围
export {};
