# id_generator

id生成器，是用于生成 hcm mysql 资源 table 表主键id的。生成的id是一个递增的8位36进制字符串。

生成的ID总数：2,821,109,907,456

### 生成方式

#### 数据表初始化

1. 在 mysql 中建立 id_generator 表，用于管理每个资源生成的最大ID。
    ```mysql
    create table if not exists `id_generator`
    (
        `resource` varchar(64) not null,
        `max_id`   varchar(64) not null,
    
        primary key (`resource`)
    ) engine = innodb
      default charset = utf8mb4;
    ```
2. 为每一个需要生成ID的资源在 id_generator 表中初始化一个默认ID。
    ```mysql
   insert into id_generator(`resource`, `max_id`) values ('resource', '0');
    ```

#### 生成逻辑

1. 获取当前的最大ID，在 id_generator 表中为资源的ID生成器加行锁。
    ```mysql
   SELECT max_id from id_generator WHERE resource = "resource" FOR UPDATE
    ```
2. 把36进制的ID转成10进制的数值，然后增加需要生成的ID数量，然后将其转回36进制的string类型的ID。
3. 将新的最大ID更新到 id_generator 表中。
    ```mysql
   UPDATE id_generator SET max_id = "new_max_id" WHERE resource = "resource" AND max_id = "old_max_id"
    ```

### 提供函数
```
Batch(kt *kit.Kit, resource table.Name, count int) ([]string, error)
One(kt *kit.Kit, resource table.Name) (string, error)
```

- Batch：用于批量申请唯一id列表
- One：用于申请单个唯一id
